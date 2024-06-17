package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"html/template"
	"net/http"
	"strings"
)

type mainPageData struct {
	User        structs.User
	Categories  []structs.Category
	RecentPosts []structs.Post
}

func SetupHandlers() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/static"))))
	http.Handle("/src/templates/", http.StripPrefix("/src/templates/", http.FileServer(http.Dir("src/templates"))))
	http.HandleFunc("/", index)

	//login register page
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/check-username", checkUsernameAvailability)
	http.HandleFunc("/check-email", checkEmailAvailability)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/login.html", serveLoginPage)
	http.HandleFunc("/register.html", serveRegisterPage)
	http.HandleFunc("/logout", logoutHandler)

	//settings page
	http.HandleFunc("/settings.html", serveSettingsPage)
	http.HandleFunc("/change-password", passwordHandler)
	http.HandleFunc("/change-email", emailHandler)
	http.HandleFunc("/change-gender", genderHandler)
	http.HandleFunc("/change-names", namesHandler)

	//categories page
	http.HandleFunc("/categories/", categoriesHandler)
	http.HandleFunc("/more-posts", showMorePosts)

	//post page
	http.HandleFunc("/create-post/", servePostCreatePage)
	http.HandleFunc("/create-post/submit", handleCreatePost)

	//administration page
	http.HandleFunc("/administration.html", serveAdministrationPage)
	http.HandleFunc("/add-category", addCategory)
	http.HandleFunc("/delete-category/", deleteCategory)
	http.HandleFunc("/update-category/", updateCategory)

	http.HandleFunc("/error.html", serveErrorPage)
}

func index(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/index.html"))

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	ExportData := mainPageData{}

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
	}
	ExportData.Categories = categories

	posts, err := src.GetRecentPosts()
	if err != nil {
		http.Error(w, "Unable to retrieve recent posts", http.StatusInternalServerError)
		return
	}
	ExportData.RecentPosts = posts

	tmpl.Execute(w, ExportData)
}
