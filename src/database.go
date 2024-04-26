package src

import (
	"database/sql"
	"log"
)

var Db *sql.DB

func SetupDatabase() *sql.DB {
	var err error
	Db, err = sql.Open("sqlite3", "users.Db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			surname TEXT NOT NULL,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			gender TEXT NOT NULL,
			admin BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	return Db
}
