package handlers

import (
	"Forum/src"
	"Forum/src/structs"
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

	tmpl := template.Must(template.ParseFiles("src/templates/administration.html"))
	tmpl.Execute(w, ExportData)
}
