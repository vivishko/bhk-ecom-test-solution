package utils

import (
	"log"
	"os"
)

func LoadEnv() {
    value := os.Getenv("POSTGRES_DB")
    if value == "" {
        log.Println("DB_HOST not set")
    }
}