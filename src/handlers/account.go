package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"html/template"
	"net/http"
)

type accountPageData struct {
	User structs.User
}

func serveSettingsPage(w http.ResponseWriter, r *http.Request) {

	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if !cookieExists(r, "sessionID") {
		http.ServeFile(w, r, "src/templates/login.html")
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/settings.html"))

	ExportData := accountPageData{}

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
