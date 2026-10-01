package user

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Model struct {
	ID            int
	Username      string
	Email         string
	Password      string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Claims struct {
	Email    string `json:"email" example:"joedoe@email.com"`
	Username string `json:"username" example:"slodkiadrianek"`
	Exp      int64
	jwt.RegisteredClaims
}
