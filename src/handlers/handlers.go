package handlers

import (
	"html/template"
	"net/http"
	"strings"
)

type exportData struct{}

var ExportData exportData

func SetupHandlers() {
	http.Handle("/src/templates/", http.StripPrefix("/src/templates/", http.FileServer(http.Dir("src/templates"))))
	http.HandleFunc("/", index)

	//login register page
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/check-username", checkUsernameAvailability)
	http.HandleFunc("/check-email", checkEmailAvailability)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/login.html", serveLoginPage)
	http.HandleFunc("/register.html", serveRegisterPage)

	http.HandleFunc("/error.html", serveErrorPage)
}

func index(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/index.html"))
	tmpl.Execute(w, ExportData)
}
