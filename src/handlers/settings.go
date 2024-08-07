package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"regexp"
)

type settingsPageData struct {
	User structs.User
}

func serveSettingsPage(w http.ResponseWriter, r *http.Request) {

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	ExportData := settingsPageData{}

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)
	ExportData.User = user

	tmpl := template.Must(template.ParseFiles("src/templates/settings.html"))
	tmpl.Execute(w, ExportData)
}

func passwordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/error.html", http.StatusSeeOther)
		return
	}

	currentPassword := r.FormValue("current-password")
	newPassword := r.FormValue("new-password")
	confirmNewPassword := r.FormValue("confirm-new-password")

	passwordError := "New password isn't compatible"
	// Limite de taille
	if len(newPassword) < 8 || len(newPassword) > 32 {
		http.Error(w, passwordError, http.StatusInternalServerError)
		return
	}
	// Vérifie la difficulté du mot de passe
	hasNumber := regexp.MustCompile(`\d`)
	if !hasNumber.MatchString(newPassword) {
		http.Error(w, passwordError, http.StatusInternalServerError)
		return
	}
	hasSpecialChar := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !hasSpecialChar.MatchString(newPassword) {
		http.Error(w, passwordError, http.StatusInternalServerError)
		return
	}
	if newPassword != confirmNewPassword {
		http.Error(w, passwordError, http.StatusInternalServerError)
		return
	}

	if currentPassword == newPassword {
		http.Error(w, "No changes detected", http.StatusInternalServerError)
		return
	}

	ExportData := settingsPageData{}
	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	ExportData.User, _ = src.GetUserFromSessionID(sessionID)
	// Récupère le mot de passe pour vérifier et comparer
	var hashedPassword string
	row := src.Db.QueryRow("SELECT password FROM users WHERE uuid = ?", ExportData.User.Uuid)
	err := row.Scan(&hashedPassword)
	errorMessage := "An error occurred"

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, errorMessage, http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(currentPassword))
	if err != nil {
		http.Error(w, "Incorrect password", http.StatusInternalServerError)
		return
	}
	// On crype le nouveau mot de passe
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	// Mise à jour du nouveau mot de passe dans la bdd
	_, err = src.Db.Exec("UPDATE users SET password = ? WHERE uuid = ?", newHashedPassword, ExportData.User.Uuid)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}
	http.Error(w, "Password changed successfully", http.StatusInternalServerError)
}

func emailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/error.html", http.StatusSeeOther)
		return
	}
	ExportData := settingsPageData{}
	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	ExportData.User, _ = src.GetUserFromSessionID(sessionID)

	email := r.FormValue("email")
	var dataEmail string
	row := src.Db.QueryRow("SELECT email FROM users WHERE uuid = ?", ExportData.User.Uuid)
	err := row.Scan(&dataEmail)
	if err != nil {
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}
	if email == dataEmail {
		http.Error(w, "No changes detected", http.StatusInternalServerError)
		return
	}
	// Mise à jour du nouveau email dans la bdd
	// Retourne une erreur s'il n'est pas unique
	_, err = src.Db.Exec("UPDATE users SET email = ? WHERE uuid = ?", email, ExportData.User.Uuid)
	if err != nil {
		http.Error(w, "Failed to update email", http.StatusInternalServerError)
		return
	}
	http.Error(w, "Email changed successfully", http.StatusInternalServerError)
}

func namesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/error.html", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	surname := r.FormValue("surname")

	nameError := "Not compatible"
	nameOrSurnamePattern := regexp.MustCompile(`^[a-zA-Z-]+$`)
	if !nameOrSurnamePattern.MatchString(name) || len(name) < 1 || len(name) > 32 {
		http.Error(w, nameError, http.StatusInternalServerError)
		return
	}
	if !nameOrSurnamePattern.MatchString(surname) || len(surname) < 1 || len(surname) > 32 {
		http.Error(w, nameError, http.StatusInternalServerError)
		return
	}

	ExportData := settingsPageData{}
	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	ExportData.User, _ = src.GetUserFromSessionID(sessionID)

	var dataName string
	var dataSurname string
	row := src.Db.QueryRow("SELECT name, surname FROM users WHERE uuid = ?", ExportData.User.Uuid)
	err := row.Scan(&dataName, &dataSurname)
	if err != nil {
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}
	if name == dataName && surname == dataSurname {
		http.Error(w, "No changes detected", http.StatusInternalServerError)
		return
	}

	_, err = src.Db.Exec("UPDATE users SET name = ?, surname = ? WHERE uuid = ?", name, surname, ExportData.User.Uuid)
	if err != nil {
		http.Error(w, "Failed to update names", http.StatusInternalServerError)
		return
	}
	http.Error(w, "Names changed successfully", http.StatusInternalServerError)
}

func genderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/error.html", http.StatusSeeOther)
		return
	}
	ExportData := settingsPageData{}
	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	ExportData.User, _ = src.GetUserFromSessionID(sessionID)

	gender := r.FormValue("gender")
	if gender != "male" && gender != "female" && gender != "other" {
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	var dataGender string
	row := src.Db.QueryRow("SELECT gender FROM users WHERE uuid = ?", ExportData.User.Uuid)
	err := row.Scan(&dataGender)
	if err != nil {
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}
	if gender == dataGender {
		http.Error(w, "No changes detected", http.StatusInternalServerError)
		return
	}
	_, err = src.Db.Exec("UPDATE users SET gender = ? WHERE uuid = ?", gender, ExportData.User.Uuid)
	if err != nil {
		http.Error(w, "Failed to update gender", http.StatusInternalServerError)
		return
	}
	http.Error(w, "Gender changed successfully", http.StatusInternalServerError)
}

func usernameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/error.html", http.StatusSeeOther)
		return
	}
	ExportData := settingsPageData{}
	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	ExportData.User, _ = src.GetUserFromSessionID(sessionID)

	username := r.FormValue("username")
	var dataUsername string
	row := src.Db.QueryRow("SELECT username FROM users WHERE uuid = ?", ExportData.User.Uuid)
	err := row.Scan(&dataUsername)
	if err != nil {
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}
	if username == dataUsername {
		http.Error(w, "No changes detected", http.StatusInternalServerError)
		return
	}
	_, err = src.Db.Exec("UPDATE users SET username = ? WHERE uuid = ?", username, ExportData.User.Uuid)
	if err != nil {
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}
	http.Error(w, "Username changed successfully", http.StatusInternalServerError)
}
