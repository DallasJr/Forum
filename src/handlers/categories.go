package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type categoriesPageData struct {
	User       structs.User
	Categories []structs.Categorie
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
		if user.Username != "" {
			ExportData.User = user
		}
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
	category := strings.TrimPrefix(r.URL.Path, "/categories/")
	if category == "" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "You are viewing category: %s", category)
}

func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/categories" || r.URL.Path == "/categories/" {
		serveCategoriesPage(w, r)
	} else {
		serveCategoryPage(w, r)
	}
}

func getAllCategories() ([]structs.Categorie, error) {
	rows, err := src.Db.Query("SELECT name, description FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []structs.Categorie
	for rows.Next() {
		var category structs.Categorie
		if err := rows.Scan(&category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}
