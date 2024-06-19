package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"html/template"
	"net/http"
	"strings"
)

type categoryPageData struct {
	User     structs.User
	Category structs.Category
	Posts    []structs.Post
}

func serveCategoryPage(w http.ResponseWriter, r *http.Request) {
	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	name := strings.TrimPrefix(r.URL.Path, "/categories/")
	tmpl := template.Must(template.ParseFiles("src/templates/category.html"))

	ExportData := categoryPageData{}

	// S'il y a un cookie de session, s'il n'est pas valide, on le supprime,
	if cookieExists(r, "sessionID") {
		sessionID := src.GetValidSession(r)
		if sessionID == "" {
			w, r = removeSession(w, r)
		}
		user, _ := src.GetUserFromSessionID(sessionID)
		ExportData.User = user
	}

	category, _ := src.GetCategory(name)
	ExportData.Category = category

	// Récupère tout les posts d'une category
	posts, _ := src.GetPostsByCategory(name)

	// Rétréci le titre et contenu ainsi que le pseudo si trop long à l'affichage
	for i := range posts {
		posts[i].Title = structs.Shorten(posts[i].Title, 20)
		posts[i].Content = structs.Shorten(posts[i].Content, 20)
		if posts[i].Creator.Username != "Deleted User" {
			posts[i].Creator.Username = structs.Shorten(posts[i].Creator.Username, 16)
		}
	}

	ExportData.Posts = posts

	tmpl.Execute(w, ExportData)
}

func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/categories" || r.URL.Path == "/categories/" {
		index(w, r)
	} else {
		serveCategoryPage(w, r)
	}
}
