package handlers

import (
	"html/template"
	"net/http"
)

type exportData struct{}

var ExportData exportData

func serveAccountPage(w http.ResponseWriter, r *http.Request) {

	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if !cookieExists(r, "sessionID") {
		http.ServeFile(w, r, "src/templates/login.html")
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/account.html"))
	tmpl.Execute(w, ExportData)
}
