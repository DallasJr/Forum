package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"html/template"
	"net/http"
	"strings"
)

type mainPageData struct {
	User       structs.User
	Categories structs.Categories
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
	http.HandleFunc("/settings.html", serveSettingsPage)
	http.HandleFunc("/categories/", serveAllCategories)

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

	tmpl.Execute(w, ExportData)
}
func serveAllCategories(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/allcategories.html"))

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

	tmpl.Execute(w, ExportData)
}
