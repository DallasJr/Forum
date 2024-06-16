package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	defaultImage := "/static/image/default-category-image.jpg"

	for {
		_, err := src.Db.Exec("INSERT INTO categories (name, description, image) VALUES (?, ?, ?)", name, description, defaultImage)
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

	category, _ := src.GetCategory(categoryName)
	// Delete the category from the database
	_, err := src.Db.Exec("DELETE FROM categories WHERE name = ?", categoryName)
	if err != nil {
		http.Error(w, "Unable to delete category", http.StatusInternalServerError)
		return
	}
	if category.Image != "/static/image/default-category-image.jpg" {
		oldImagePath := fmt.Sprintf("src%s", category.Image)
		err := os.Remove(oldImagePath)
		if err != nil {
			http.Error(w, "Failed to delete old image", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/administration.html?message=Category%20deleted%20successfully.", http.StatusSeeOther)
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

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
	prevName := r.FormValue("prevName")
	category, _ := src.GetCategory(prevName)
	newName := r.FormValue("newName")
	newDescription := r.FormValue("newDescription")
	newImage := ""

	// Validate form values
	if newName == "" || newDescription == "" {
		http.Error(w, "New name and description are required", http.StatusInternalServerError)
		return
	}

	// Handle file upload for new image
	file, header, err := r.FormFile("newImage")
	if err == nil {
		defer file.Close()

		if category.Image != "/static/image/default-category-image.jpg" {
			oldImagePath := fmt.Sprintf("src%s", category.Image)
			err := os.Remove(oldImagePath)
			if err != nil {
				http.Error(w, "Failed to delete old image", http.StatusInternalServerError)
				return
			}
		}
		imagePath := fmt.Sprintf("src/static/image/category-%s-image%s", newName, filepath.Ext(header.Filename))
		out, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Unable to upload image", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Unable to save image", http.StatusInternalServerError)
			return
		}
		newImage = fmt.Sprintf("/static/image/category-%s-image%s", newName, filepath.Ext(header.Filename))
	} else {
		if prevName != newName {
			if category.Image != "/static/image/default-category-image.jpg" {
				oldImagePath := fmt.Sprintf("src%s", category.Image)
				newImagePath := fmt.Sprintf("src/static/image/category-%s-image%s", newName, filepath.Ext(category.Image))
				err := os.Rename(oldImagePath, newImagePath)
				if err != nil {
					http.Error(w, "Failed to rename image", http.StatusInternalServerError)
					return
				}
				newImage = fmt.Sprintf("/static/image/category-%s-image%s", newName, filepath.Ext(category.Image))
			}
		}
	}

	// Update the category in the database
	query := "UPDATE categories SET name = ?, description = ?"
	args := []interface{}{newName, newDescription}

	if newImage != "" {
		query += ", image = ?"
		args = append(args, newImage)
	}
	query += " WHERE name = ?"
	args = append(args, categoryName)

	_, err = src.Db.Exec(query, args...)
	if err != nil {
		http.Error(w, "Unable to update category", http.StatusInternalServerError)
		return
	}
	if prevName != newName {
		_, err = src.Db.Exec(`
			UPDATE posts
			SET category_name = ?
			WHERE category_name = ?
		`, newName, prevName)
		if err != nil {
			return
		}
	}

	http.Redirect(w, r, "/administration.html?message=Category%20updated%20successfully.", http.StatusSeeOther)
}
