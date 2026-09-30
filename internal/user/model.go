package user

import "time"

type Model struct {
	ID            int
	Username      string
	Email         string
	Password      string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
