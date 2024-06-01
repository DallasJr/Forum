package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"html/template"
	"net/http"
	"strings"
)

type categoriesPageData struct {
	User       structs.User
	Categories []structs.Category
}
type categoryPageData struct {
	User     structs.User
	Category structs.Category
}

func serveCategoriesPage(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/categories.html"))

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	ExportData := categoriesPageData{}

	if cookieExists(r, "sessionID") {
		sessionID := src.GetValidSession(r)
		if sessionID == "" {
			logoutHandler(w, r)
			return
		}
		user, _ := src.GetUserFromSessionID(sessionID)
		ExportData.User = user
	}

	categories, err := getAllCategories()
	if err != nil {
		http.Error(w, "Unable to retrieve categories", http.StatusInternalServerError)
		return
	}
	ExportData.Categories = categories

	tmpl.Execute(w, ExportData)
}

func serveCategoryPage(w http.ResponseWriter, r *http.Request) {
	// Remove the prefix "/categories/"
	name := strings.TrimPrefix(r.URL.Path, "/categories/")
	if name == "" {
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/category.html"))

	ExportData := categoryPageData{}

	if cookieExists(r, "sessionID") {
		sessionID := src.GetValidSession(r)
		if sessionID == "" {
			logoutHandler(w, r)
			return
		}
		user, _ := src.GetUserFromSessionID(sessionID)
		ExportData.User = user
	}

	category, _ := src.GetCategory(name)
	ExportData.Category = category

	tmpl.Execute(w, ExportData)
}

func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/categories" || r.URL.Path == "/categories/" {
		serveCategoriesPage(w, r)
	} else {
		serveCategoryPage(w, r)
	}
}

func getAllCategories() ([]structs.Category, error) {
	rows, err := src.Db.Query("SELECT name, description FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []structs.Category
	for rows.Next() {
		var category structs.Category
		if err := rows.Scan(&category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}
