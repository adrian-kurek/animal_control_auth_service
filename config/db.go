package config

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

const (
	defaultConnMaxIdleTime = 30 * time.Second
	defaultMaxOpenConns    = 20
	defaultMaxIdleConns    = 20
	defaultDBTimeout       = 5 * time.Second
)

type DB struct {
	Connection *sql.DB
}

func NewDB(databaseLink, dbDriver string) (*DB, error) {
	dbConnection, err := sql.Open(dbDriver, databaseLink)
	if err != nil {
		return nil, err
	}

	dbConnection.SetConnMaxIdleTime(defaultConnMaxIdleTime)
	dbConnection.SetMaxOpenConns(defaultMaxOpenConns)
	dbConnection.SetMaxIdleConns(defaultMaxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), defaultDBTimeout)
	defer cancel()

	if err = dbConnection.PingContext(ctx); err != nil {
		return nil, err
	}
	return &DB{
		Connection: dbConnection,
	}, nil
}

func (db *DB) Close() error {
	return db.Connection.Close()
}
