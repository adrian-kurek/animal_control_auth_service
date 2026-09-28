package main

import (
	"log"
	"os"

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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db, err := connectToDB()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()
}
