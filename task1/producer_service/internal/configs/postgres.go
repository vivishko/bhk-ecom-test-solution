package configs

import (
	"context"
	"fmt"
	"os"
	"producer_service/internal/utils"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPostgresDB() (*pgxpool.Pool, error) {
    utils.LoadEnv()

    dbHost := os.Getenv("POSTGRES_HOST")
    dbPort := "5432"
    dbUser := os.Getenv("POSTGRES_USER")
    dbPassword := os.Getenv("POSTGRES_PASSWORD")
    dbName := os.Getenv("POSTGRES_DB")

    dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
        dbUser, dbPassword, dbHost, dbPort, dbName)

    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("unable to parse config: %v", err)
    }

    dbpool, err := pgxpool.ConnectConfig(context.Background(), config)
    if err != nil {
        return nil, fmt.Errorf("unable to connect to database: %v", err)
    }

    return dbpool, nil
}
