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

	// Users
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			surname TEXT NOT NULL,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			gender TEXT NOT NULL,
			power INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Posts
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			uuid TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			category_name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY(owner_id) REFERENCES users(uuid),
			FOREIGN KEY(category_name) REFERENCES categories(name)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Answers
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS answers (
			uuid TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			post_id TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY(owner_id) REFERENCES users(uuid),
			FOREIGN KEY(post_id) REFERENCES posts(uuid)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Categories
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
    		name TEXT PRIMARY KEY,
    		description TEXT
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
	query := `SELECT uuid, name, surname, username, email, gender, created_at, power FROM users WHERE uuid = ?`
	err := Db.QueryRow(query, userID).Scan(
		&user.Uuid, &user.Name, &user.Surname, &user.Username,
		&user.Email, &user.Gender, &user.CreationDate, &user.Power,
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
