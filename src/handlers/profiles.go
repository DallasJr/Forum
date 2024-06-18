package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"html/template"
	"net/http"
	"strings"
)

type profilePageData struct {
	User         structs.User
	Posts        []structs.Post
	Answers      []structs.Answer
	LikedPosts   []structs.Post
	LikedAnswers []structs.Answer
}

func serveProfilePage(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/profile.html"))

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	ExportData := profilePageData{}

	if cookieExists(r, "sessionID") {
		sessionID := src.GetValidSession(r)
		if sessionID == "" {
			w, r = removeSession(w, r)
		}
		user, _ := src.GetUserFromSessionID(sessionID)
		ExportData.User = user
	}
	tmpl.Execute(w, ExportData)
}

func profilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/profile" || r.URL.Path == "/profiles/" {
		index(w, r)
		return
	} else {
		serveProfilePage(w, r)
	}
}
