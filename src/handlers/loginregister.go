package handlers

import (
	"Forum/src"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
)

// Generate a random session ID
func generateSessionID() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	var id string
	var hashedPassword string
	row := src.Db.QueryRow("SELECT id, password FROM accounts WHERE username = ?", username)
	err := row.Scan(&id, &hashedPassword)
	errorMessage := "Invalid username or password"
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, errorMessage, http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		http.Error(w, errorMessage, http.StatusUnauthorized)
		return
	}

	// Authentication successful, generate session ID
	sessionID := generateSessionID()

	// Store the session ID in the map
	src.Mutex.Lock()
	src.Sessions[sessionID] = id
	fmt.Println("logged: " + id)
	src.Mutex.Unlock()

	// Set session ID as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    sessionID,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func returnError(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/error.html", http.StatusSeeOther)
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

	nameOrSurnamePattern := regexp.MustCompile(`^[a-zA-Z-]+$`)
	if !nameOrSurnamePattern.MatchString(name) || len(name) < 1 || len(name) > 32 {
		returnError(w, r)
		return
	}
	if !nameOrSurnamePattern.MatchString(surname) || len(surname) < 1 || len(surname) > 32 {
		returnError(w, r)
		return
	}
	usernamePattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	if !usernamePattern.MatchString(username) || len(username) < 3 || len(username) > 16 {
		returnError(w, r)
		return
	}
	if len(password) < 8 || len(password) > 32 {
		returnError(w, r)
		return
	}
	hasNumber := regexp.MustCompile(`\d`)
	if !hasNumber.MatchString(password) {
		returnError(w, r)
		return
	}
	hasSpecialChar := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !hasSpecialChar.MatchString(password) {
		returnError(w, r)
		return
	}
	hasSpaces := regexp.MustCompile(`\s`)
	if hasSpaces.MatchString(password) {
		returnError(w, r)
		return
	}
	if gender != "male" && gender != "female" {
		returnError(w, r)
		return
	}

	uniqueId := uuid.Must(uuid.NewV4())
	userID := uniqueId.String()

	hashedPassword, er := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if er != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	_, err := src.Db.Exec("INSERT INTO accounts (id, name, surname, username, email, password, admin, gender) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userID, name, surname, strings.ToLower(username), email, hashedPassword, false, gender)
	if err != nil {
		returnError(w, r)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	// Authentication successful, generate session ID
	sessionID := generateSessionID()

	// Store the session ID in the map
	src.Mutex.Lock()
	src.Sessions[sessionID] = userID
	src.Mutex.Unlock()

	// Set session ID as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    sessionID,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func cookieExists(r *http.Request, cookieName string) bool {
	_, err := r.Cookie(cookieName)
	return !errors.Is(err, http.ErrNoCookie)
}

func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	if cookieExists(r, "sessionID") && src.GetValidSession(r) != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.ServeFile(w, r, "src/templates/login.html")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "sessionID",
		Value:  "",
		MaxAge: -1,  // Set MaxAge to -1 to delete the cookie
		Path:   "/", // Same path as the session cookie
	})

	// Redirect the user to the login page or any other page
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}

func serveRegisterPage(w http.ResponseWriter, r *http.Request) {
	if cookieExists(r, "sessionID") && src.GetValidSession(r) != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	/*tmpl := template.Must(template.ParseFiles("src/templates/register.html"))
	tmpl.Execute(w, ExportData)*/

	http.ServeFile(w, r, "src/templates/register.html")
}

func checkUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	var count int
	err := src.Db.QueryRow("SELECT COUNT(*) FROM accounts WHERE username = ?", username).Scan(&count)
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
	err := src.Db.QueryRow("SELECT COUNT(*) FROM accounts WHERE email = ?", email).Scan(&count)
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

func serveErrorPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "src/templates/error.html")
}
