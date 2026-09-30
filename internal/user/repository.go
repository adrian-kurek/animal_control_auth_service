package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	commonerrors "github.com/adrian-kurek/animal_control_auth_service/common/errors"
	"github.com/adrian-kurek/animal_control_auth_service/internal/auth"
)

const (
	errorMsg = "error"
	queryMsg = "query"
	argsMsg  = "args"
	emailMsg = "email"
)

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

func (ur *Repository) Create(ctx context.Context, user auth.NewUser, hashedPassword string) error {
	query := `INSERT INTO users (email,username,password,created_at,updated_at) VALUES ($1,$2,$3,now(),now())`

	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.loggerService.ErrorContext(ctx, commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.loggerService.ErrorContext(ctx, commonerrors.FailedToCloseStatement, "data", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, user.Email, user.Username, hashedPassword)
	if err != nil {
		ur.loggerService.ErrorContext(ctx, commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]string{
				"username": user.Username,
				emailMsg:   user.Email,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (ur *Repository) Update(ctx context.Context, username, email string) error {
	query := `UPDATE users SET username = $1 WHERE email = $2`

	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.loggerService.ErrorContext(ctx, commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.loggerService.ErrorContext(ctx, commonerrors.FailedToCloseStatement, "data", closeErr)
		}
	}()

	_, err = stmt.ExecContext(ctx, username, email)
	if err != nil {
		ur.loggerService.ErrorContext(ctx, commonerrors.FailedToExecuteInsertQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg: map[string]string{
				"username": username,
				emailMsg:   email,
			},
			errorMsg: err,
		})
		return err
	}

	return nil
}

func (ur *Repository) FindByEmail(ctx context.Context, email string) (Model, error) {
	query := "SELECT  id, email, username, password, email_verified, created_at, updated_at FROM USERS WHERE email = $1"
	stmt, err := ur.db.PrepareContext(ctx, query)
	if err != nil {
		ur.loggerService.ErrorContext(ctx, commonerrors.FailedToPrepareQuery, "data", map[string]string{
			queryMsg: query,
			errorMsg: err.Error(),
		})
		return Model{}, err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			ur.loggerService.ErrorContext(ctx, commonerrors.FailedToCloseStatement, "data", closeErr)
		}
	}()

	var user Model
	err = stmt.QueryRowContext(ctx, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ur.loggerService.InfoContext(ctx, "user not found", "data", map[string]any{
				emailMsg: email,
			})
			return Model{
				ID: 0,
			}, nil
		}
		ur.loggerService.ErrorContext(ctx, commonerrors.FailedToExecuteSelectQuery, "data", map[string]any{
			queryMsg: query,
			argsMsg:  []any{email},
			errorMsg: err,
		})
		return Model{}, err
	}
	return user, nil
}
