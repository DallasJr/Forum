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
	http.HandleFunc("/change-username", usernameHandler)
	http.HandleFunc("/change-names", namesHandler)

	//categories page
	http.HandleFunc("/categories/", categoriesHandler)

	//profile page
	http.HandleFunc("/profile/", profilesHandler)

	//post page
	http.HandleFunc("/create-post/", servePostCreatePage)
	http.HandleFunc("/create-post/submit", handleCreatePost)
	http.HandleFunc("/post/", postsHandler)
	http.HandleFunc("/submit-answer", handleAnswerSubmission)

	//administration page
	http.HandleFunc("/administration.html", serveAdministrationPage)
	http.HandleFunc("/add-category", addCategory)
	http.HandleFunc("/delete-category/", deleteCategory)
	http.HandleFunc("/update-category/", updateCategory)
	http.HandleFunc("/delete-post/", deletePost)
	http.HandleFunc("/delete-answer/", deleteAnswer)
	http.HandleFunc("/delete-user/", deleteUser)
	http.HandleFunc("/update-user/", updateUser)

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
			w, r = removeSession(w, r)
		}
		user, _ := src.GetUserFromSessionID(sessionID)
		ExportData.User = user
	}

	categories, err := src.GetAllCategories()
	if err != nil {
		http.Error(w, "Unable to retrieve categories", http.StatusInternalServerError)
	}
	ExportData.Categories = categories

	posts, err := src.GetRecentPosts()
	if err != nil {
		http.Error(w, "Unable to retrieve recent posts", http.StatusInternalServerError)
		return
	}
	for i := range posts {
		posts[i].Title = structs.Shorten(posts[i].Title, 20)
		posts[i].Content = structs.Shorten(posts[i].Content, 20)
		if posts[i].Creator.Username != "Deleted User" {
			posts[i].Creator.Username = structs.Shorten(posts[i].Creator.Username, 12)
		}
	}
	ExportData.RecentPosts = posts

	tmpl.Execute(w, ExportData)
}
