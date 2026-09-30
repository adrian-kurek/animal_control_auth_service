package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/adrian-kurek/animal_control_auth_service/common/logger"
	"github.com/adrian-kurek/animal_control_auth_service/config"
	"github.com/adrian-kurek/animal_control_auth_service/internal/auth"
	"github.com/adrian-kurek/animal_control_auth_service/internal/server"
	"github.com/adrian-kurek/animal_control_auth_service/internal/user"
	"github.com/joho/godotenv"
)

func connectToDB() (*config.DB, error) {
	dbConnectionLink := os.Getenv("DB_LINK")

	db, err := config.NewDB(dbConnectionLink, "postgres")
	if err != nil {
		return &config.DB{}, err
	}

	return db, nil
}

func connectToCache() (*config.CacheService, error) {
	cacheConnectionLink := os.Getenv("CACHE_LINK")

	cacheService, err := config.NewCacheService(cacheConnectionLink)
	if err != nil {
		return &config.CacheService{}, err
	}
	return cacheService, nil
}

func bootstrapDependencies(loggerService *slog.Logger, db *config.DB, _ config.CacheService, port string) *server.HTTP {
	userRepository := user.NewRepository(db.Connection, loggerService)

	authService := auth.NewService(userRepository, loggerService)
	authHandler := auth.NewHandler(authService, loggerService)

	dependencies := server.NewDependencyConfig(port, *authHandler)
	return server.NewHTTP(dependencies)
}

func main() {
	logger := logger.Setup()

	err := godotenv.Load()
	if err != nil {
		logger.Error("Failed to load environment variables", "err:", err.Error())
		panic(err)
	}
	db, err := connectToDB()
	if err != nil {
		logger.Error("Failed to connecto to databse service", "err:", err.Error())
		panic(err)
	}
	defer db.Close()

	cacheService, err := connectToCache()
	if err != nil {
		logger.Error("Failed to connecto to cache service", "err:", err.Error())
		panic(err)
	}
	defer cacheService.Close()
	apiCtx, apiCtxCancel := context.WithCancel(context.Background())

	port := os.Getenv("PORT")
	httpServer := bootstrapDependencies(logger, db, *cacheService, port)
	go func() {
		logger.Info("Applicattion started", "port", port)
		if err = httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	defer apiCtxCancel()

	if err = httpServer.Shutdown(apiCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		err = db.Close()
		if err != nil {
			panic(err)
		}
		err = cacheService.Close()
		if err != nil {
			panic(err)
		}
	}

	logger.Info("server exited")
}
