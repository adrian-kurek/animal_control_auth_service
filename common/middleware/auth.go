package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	commonerrors "github.com/adrian-kurek/animal_control_auth_service/common/errors"
	"github.com/adrian-kurek/animal_control_auth_service/common/interfaces"
	"github.com/adrian-kurek/animal_control_auth_service/common/request"
	"github.com/adrian-kurek/animal_control_auth_service/internal/user"
	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenExpiration = 10 * time.Minute
	lengthOfRefreshToken  = 64
)

type TokenService struct {
	refreshSecret string
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	isUser        string
	loggerService *slog.Logger
	cacheService  interfaces.CacheService
}

func NewTokenService(
	refreshSecret, isUser string, loggerService *slog.Logger, cacheService interfaces.CacheService,
) *TokenService {
	return &TokenService{
		refreshSecret: refreshSecret,
		isUser:        isUser,
		loggerService: loggerService,
		cacheService:  cacheService,
	}
}

func (ts *TokenService) GenerateRefreshToken() ([]byte, error) {
	bytes := make([]byte, lengthOfRefreshToken)

	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (ts *TokenService) readTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ts.loggerService.Info("token is missing")
		return "", commonerrors.Unauthorized("failed to authorize a user")
	}

	return strings.Split(authHeader, " ")[1], nil
}

func (ts *TokenService) isTokenBlackListed(ctx context.Context, token string) error {
	cacheKey := "tokenBlacklist-" + token

	res, err := ts.cacheService.Exists(ctx, cacheKey)
	if err != nil {
		return err
	}

	if res > 0 {
		err = errors.New("token blacklisted")
		ts.loggerService.Info(err.Error())
		return commonerrors.Unauthorized(err.Error())
	}

	return nil
}

func (ts *TokenService) checkIsTokenValid(token *jwt.Token) error {
	if !token.Valid {
		err := errors.New("provided token is invalid")
		ts.loggerService.Info(err.Error())
		return commonerrors.Unauthorized(err.Error())
	}

	return nil
}

func (ts *TokenService) HashToken(token []byte) string {
	hasher := sha256.New()
	hasher.Write(token)
	return hex.EncodeToString(hasher.Sum(nil))
}

func (ts *TokenService) ParseClaimsFromToken(tokenString string) (*jwt.Token, user.Claims, error) {
	var claims user.Claims
	token, err := jwt.ParseWithClaims(
		tokenString, &claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodES256 {
				return nil, commonerrors.Unauthorized("invalid signing method")
			}
			return ts.publicKey, nil
		},
		jwt.WithIssuer(ts.isUser),
		jwt.WithAudience("auth-service"),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, user.Claims{}, err
	}
	return token, claims, nil
}

func (ts *TokenService) BlackListUser(r *http.Request) error {
	ctx := r.Context()

	token, err := ts.readTokenFromRequest(r)
	if err != nil {
		return err
	}

	tokenWithData, user, err := ts.ParseClaimsFromToken(token)
	if err != nil {
		ts.loggerService.Info("failed to read data properly")
		return commonerrors.Unauthorized("failed to read token")
	}

	err = ts.checkIsTokenValid(tokenWithData)
	if err != nil {
		return err
	}
	err = ts.isTokenBlackListed(ctx, token)
	if err != nil {
		return err
	}

	expirationTime := time.Until(user.ExpiresAt.Time)

	cacheKey := "tokenBlackList-" + token
	err = ts.cacheService.Set(ctx, cacheKey, "true", expirationTime)
	if err != nil {
		ts.loggerService.Info("failed to set data in cache", "error", err)
		return err
	}

	return nil
}

func (ts *TokenService) GenerateAccessToken(userData user.Model) (string, error) {
	now := time.Now()

	claims := user.Claims{
		Email:    userData.Email,
		Exp:      now.Add(accessTokenExpiration).Unix(),
		Username: userData.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(userData.ID),
			Issuer:  ts.isUser,
			Audience: jwt.ClaimStrings{
				"auth-service",
			},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenExpiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	token.Header["kid"] = "auth-key"
	return token.SignedString(ts.privateKey)
}

func (ts *TokenService) VerifyToken(r *http.Request) (*http.Request, error) {
	ctx := r.Context()
	if err := ctx.Err(); err != nil {
		return r, err
	}

	token, err := ts.readTokenFromRequest(r)
	if err != nil {
		return r, err
	}

	err = ts.isTokenBlackListed(ctx, token)
	if err != nil {
		return r, err
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return r, err
	}

	tokenWithClaims, claims, err := ts.ParseClaimsFromToken(token)
	if err != nil {
		return r, err
	}
	if err = ctx.Err(); err != nil {
		return r, err
	}

	err = ts.checkIsTokenValid(tokenWithClaims)
	if err != nil {
		return r, err
	}

	r = request.SetContext(r, "email", claims.Email)
	r = request.SetContext(r, "username", claims.Username)
	return r, nil
}
