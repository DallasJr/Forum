package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type categoryPageData struct {
	User     structs.User
	Category structs.Category
	Posts    []structs.Post
}

func serveCategoryPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	name := strings.TrimPrefix(r.URL.Path, "/categories/")
	/*if name == "" {
		http.NotFound(w, r)
		return
	}*/
	tmpl := template.Must(template.ParseFiles("src/templates/category.html"))

	ExportData := categoryPageData{}

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

	posts, _ := src.GetPostsByCategory(name, 0, 5)

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

func showMorePosts(w http.ResponseWriter, r *http.Request) {
	categoryName := r.URL.Query().Get("category")
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}
	posts, err := src.GetPostsByCategory(categoryName, offset, 5)
	if err != nil {
		http.Error(w, "Unable to retrieve more posts", http.StatusInternalServerError)
		return
	}
	for i := range posts {
		posts[i].Title = structs.Shorten(posts[i].Title, 20)
		posts[i].Content = structs.Shorten(posts[i].Content, 20)
		if posts[i].Creator.Username != "Deleted User" {
			posts[i].Creator.Username = structs.Shorten(posts[i].Creator.Username, 16)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/categories" || r.URL.Path == "/categories/" {
		index(w, r)
	} else {
		serveCategoryPage(w, r)
	}
}

func getAllCategories() ([]structs.Category, error) {
	rows, err := src.Db.Query("SELECT name, description, image FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []structs.Category
	for rows.Next() {
		var category structs.Category
		if err := rows.Scan(&category.Name, &category.Description, &category.Image); err != nil {
			return nil, err
		}

		postCount, err := src.GetPostsCountByCategory(category.Name)
		if err != nil {
			return nil, err
		}
		category.PostsCount = postCount

		categories = append(categories, category)
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].PostsCount > categories[j].PostsCount
	})
	return categories, nil
}
