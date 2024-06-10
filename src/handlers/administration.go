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

func deleteCategory(w http.ResponseWriter, r *http.Request) {

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
	categoryName := r.URL.Path[len("/delete-category/"):]

	// Delete the category from the database
	_, err := src.Db.Exec("DELETE FROM categories WHERE name = ?", categoryName)
	if err != nil {
		http.Error(w, "Unable to delete category", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/administration.html?message=Category%20deleted%20successfully.", http.StatusSeeOther)
}

func updateCategory(w http.ResponseWriter, r *http.Request) {

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
	categoryName := r.URL.Path[len("/update-category/"):]

	// Get form values
	newName := r.FormValue("newName")
	newDescription := r.FormValue("newDescription")

	// Validate form values
	if newName == "" || newDescription == "" {
		http.Error(w, "New name and description are required", http.StatusBadRequest)
		return
	}

	// Update the category in the database
	_, err := src.Db.Exec("UPDATE categories SET name = ?, description = ? WHERE name = ?", newName, newDescription, categoryName)
	if err != nil {
		http.Error(w, "Unable to update category", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/administration.html?message=Category%20updated%20successfully.", http.StatusSeeOther)
}
