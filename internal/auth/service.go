package auth

import (
	"context"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type userRepository interface {
	Create(ctx context.Context, user NewUser, hashedPassword string) error
}

type Service struct {
	loggerService  *slog.Logger
	userRepository userRepository
}

func NewService(userRepository userRepository, loggerService *slog.Logger) *Service {
	return &Service{
		loggerService:  loggerService,
		userRepository: userRepository,
	}
}

func (as *Service) Register(ctx context.Context, user NewUser) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return as.userRepository.Create(ctx, user, string(hashedPassword))
}
