package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func SQLConnect() (*sql.DB, error) {
	if db == nil {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USERNAME"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)

		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			LogError(fmt.Errorf("error connecting to db : %v", err))
			return nil, err
		}

		err = db.Ping()
		if err != nil {
			LogError(fmt.Errorf("error pinging db : %v", err))
			return nil, err
		}
	}
	return db, nil
}
