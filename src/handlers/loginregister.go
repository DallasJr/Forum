package handlers

import (
	"Forum/src"
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
)

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

	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}

/*g := gender == "male"
user := structs.User{uniqueId, name, surname, username, email, password, g, time.Now(), false}
jsonBytes, err := json.Marshal(user)
if err != nil {
	fmt.Println("Couldn't save data")
	return
}
file, err := os.Create("data.json")
if err != nil {
	fmt.Println("Couldn't save data")
	return
}
defer file.Close()
_, err = file.Write(jsonBytes)
if err != nil {
	fmt.Println("Couldn't save data")
}*/

func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "src/templates/login.html")
}

func serveErrorPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "src/templates/error.html")
}

func serveRegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "src/templates/register.html")
	/*tmpl := template.Must(template.ParseFiles("src/templates/register.html"))
	tmpl.Execute(w, ExportData)*/
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
