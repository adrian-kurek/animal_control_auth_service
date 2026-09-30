package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	commonerrors "github.com/adrian-kurek/animal_control_auth_service/common/errors"
	"github.com/adrian-kurek/animal_control_auth_service/common/request"
	"github.com/adrian-kurek/animal_control_auth_service/common/response"
)

type authService interface {
	Register(ctx context.Context, user NewUser) error
}

const registerTimeout = 10 * time.Second

type Handler struct {
	authService   authService
	loggerService *slog.Logger
}

func NewHandler(authService authService, loggerService *slog.Logger) *Handler {
	return &Handler{
		authService:   authService,
		loggerService: loggerService,
	}
}

func (ah *Handler) handleTimeout(err error, path string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		ah.loggerService.Info("request timed out", "path", path)
		return commonerrors.RequestTimeout()
	}
	return err
}

func (ah *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), registerTimeout)
	defer cancel()

	newUser, err := request.ReadBody[NewUser](r)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	err = ah.authService.Register(ctx, *newUser)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]string{})

	return nil
}
