package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"html/template"
	"net/http"
)

type administrationPageData struct {
	User       structs.User
	Categories []structs.Category
}

func serveAdministrationPage(w http.ResponseWriter, r *http.Request) {

	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if !cookieExists(r, "sessionID") {
		http.ServeFile(w, r, "src/templates/login.html")
		return
	}

	ExportData := administrationPageData{}

	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		logoutHandler(w, r)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)

	if user.Power == 0 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}

	ExportData.User = user

	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Unable to retrieve categories", http.StatusInternalServerError)
	}
	ExportData.Categories = categories

	tmpl := template.Must(template.ParseFiles("src/templates/administration.html"))
	tmpl.Execute(w, ExportData)
}

func addCategory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if !cookieExists(r, "sessionID") {
		http.ServeFile(w, r, "src/templates/login.html")
		return
	}

	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		logoutHandler(w, r)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)

	if user.Power != 2 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}
	name := "Name"
	description := "Description"
	i := 1
	baseName := name
	message := "Couldn't add category."

	for {
		_, err := src.Db.Exec("INSERT INTO categories (name, description) VALUES (?, ?)", name, description)
		if err != nil {
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintPrimaryKey) {
				name = fmt.Sprintf("%s%d", baseName, i)
				i++
			}
		} else {
			message = "Category added successfully."
			break
		}
	}
	http.Redirect(w, r, "/administration.html?message="+message, http.StatusSeeOther)
}
