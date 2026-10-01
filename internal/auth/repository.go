package auth

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	commonerrors "github.com/adrian-kurek/animal_control_auth_service/common/errors"
)

const refreshTokenExpiration = 7 * 24 * time.Hour

type Repository struct {
	db            *sql.DB
	loggerService *slog.Logger
}

func NewRepository(db *sql.DB, loggerService *slog.Logger) *Repository {
	return &Repository{
		db:            db,
		loggerService: loggerService,
	}
}

func (ar *Repository) InsertRefreshToken(
	ctx context.Context,
	ipAddress string,
	deviceInfo string,
	refreshToken string,
	userID int,
) error {
	query := `INSERT INTO refresh_tokens(user_id,token_hash,device_info,ip_address, expires_at, last_used_at) VALUES($1,$2,$3,$4,$5,$6)`

	stmt, err := ar.db.PrepareContext(ctx, query)
	if err != nil {
		ar.loggerService.Error(commonerrors.FailedToPrepareQuery, "data", map[string]string{
			"query": query,
			"error": err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ar.loggerService.Error(commonerrors.FailedToCloseStatement, closeErr)
		}
	}()

	expiration := time.Now().Add(refreshTokenExpiration)
	lastUsedAt := time.Now()

	_, err = stmt.ExecContext(ctx, userID, refreshToken, deviceInfo, ipAddress, expiration, lastUsedAt)
	if err != nil {
		ar.loggerService.Error(commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			"query": query,
			"args": map[string]any{
				"userId":     userID,
				"deviceInfo": deviceInfo,
				"ipAddress":  ipAddress,
			},
			"error": err,
		})
		return err
	}

	return nil
}
