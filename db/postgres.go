package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@postgres/db_commands?sslmode=disable")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS commands (
        id SERIAL PRIMARY KEY,
        command TEXT NOT NULL,
        result TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}
