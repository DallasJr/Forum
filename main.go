package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
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

	http.Handle("/src/templates/", http.StripPrefix("/src/templates/", http.FileServer(http.Dir("src/templates"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/check-username", checkUsernameAvailability)
	http.HandleFunc("/check-email", checkEmailAvailability)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/login.html", serveLoginPage)

	fmt.Println("http://localhost/")
	err = http.ListenAndServe("", nil)
	if err != nil {
		log.Fatal(err)
		return
	}
}

type exportData struct {
}

var ExportData exportData

func index(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/register.html"))
	tmpl.Execute(w, ExportData)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	surname := r.FormValue("surname")
	name := r.FormValue("name")
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	gender := r.FormValue("gender")
	userID := uuid.New().String()

	_, err := db.Exec("INSERT INTO users (id, name, surname, username, email, password, admin, gender) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userID, name, surname, username, email, password, false, gender)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}

func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "src/templates/login.html")
}

func checkUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		http.Error(w, "Failed to check username availability", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		fmt.Fprintf(w, "available")
	} else {
		fmt.Fprintf(w, "not available")
	}
}
func checkEmailAvailability(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		http.Error(w, "Failed to check email availability", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		fmt.Fprintf(w, "available")
	} else {
		fmt.Fprintf(w, "not available")
	}
}
