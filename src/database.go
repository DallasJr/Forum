package src

import (
	"Forum/src/structs"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	Sessions = make(map[string]string)
	Mutex    sync.Mutex
)

var Db *sql.DB

func SetupDatabase() *sql.DB {
	var err error
	Db, err = sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
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

func GetUserFromSessionID(sessionID string) (structs.User, error) {
	Mutex.Lock()
	userID, exists := Sessions[sessionID]
	Mutex.Unlock()
	if !exists {
		return structs.User{}, nil // Session found
	}
	var user structs.User
	query := `SELECT id, name, surname, username, email, gender, created_at, admin FROM accounts WHERE id = ?`
	err := Db.QueryRow(query, userID).Scan(
		&user.Uuid, &user.Name, &user.Surname, &user.Username,
		&user.Email, &user.Gender, &user.CreationDate, &user.Admin,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return structs.User{}, nil // No user found
		}
		fmt.Println("query error:", err)
		return structs.User{}, err
	}
	return user, nil
}

func GetValidSession(r *http.Request) string {
	c, _ := r.Cookie("sessionID")
	Mutex.Lock()
	_, exists := Sessions[c.Value]
	Mutex.Unlock()
	if !exists {
		return ""
	} else {
		return c.Value
	}
}
