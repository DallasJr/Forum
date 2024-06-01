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

	// Accounts
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
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
			id TEXT PRIMARY KEY,
			creator TEXT NOT NULL,
			creation_date TIMESTAMP NOT NULL,
			FOREIGN KEY (creator) REFERENCES accounts(id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Answers
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS answers (
			id TEXT PRIMARY KEY,
			creator TEXT NOT NULL,
			creation_date TIMESTAMP NOT NULL,
			FOREIGN KEY (creator) REFERENCES accounts(id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Posts answers
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS posts_answers (
			post_id TEXT NOT NULL,
			answer_id TEXT NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (answer_id) REFERENCES answers(id)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Categorie
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
    		name TEXT PRIMARY KEY,
    		description TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS categories_posts (
			category_name TEXT NOT NULL,
			post_id TEXT NOT NULL,
			FOREIGN KEY (category_name) REFERENCES categories(name),
			FOREIGN KEY (post_id) REFERENCES posts(id)
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
	query := `SELECT id, name, surname, username, email, gender, created_at, power FROM accounts WHERE id = ?`
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
