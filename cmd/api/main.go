package main

import (
	"os"

	"github.com/adrian-kurek/animal_control_auth_service/common/logger"
	"github.com/adrian-kurek/animal_control_auth_service/config"
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
	logger.Info("Applicattion started")
}
