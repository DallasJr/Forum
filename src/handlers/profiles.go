package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"github.com/google/uuid"
	"html/template"
	"net/http"
	"strings"
)

type profilePageData struct {
	User         structs.User
	Profile      structs.User
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

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	profileID := strings.TrimPrefix(r.URL.Path, "/profile/")

	ExportData := profilePageData{}

	// S'il y a un cookie de session, s'il n'est pas valide, on le supprime,
	if cookieExists(r, "sessionID") {
		sessionID := src.GetValidSession(r)
		if sessionID == "" {
			w, r = removeSession(w, r)
		}
		user, _ := src.GetUserFromSessionID(sessionID)
		ExportData.User = user
	}

	//Profile user
	parsedUUID, _ := uuid.Parse(profileID)
	profile, _ := src.GetUserFromUUID(parsedUUID)
	profile.CreationDate, _ = profile.FormattedDate()
	ExportData.Profile = profile

	//Posts
	posts, _ := src.GetPostsByUser(profileID)
	// Rétréci le titre et contenu ainsi que le pseudo si trop long à l'affichage
	for i := range posts {
		posts[i].Title = structs.Shorten(posts[i].Title, 20)
		posts[i].Content = structs.Shorten(posts[i].Content, 20)
		if posts[i].Creator.Username != "Deleted User" {
			posts[i].Creator.Username = structs.Shorten(posts[i].Creator.Username, 12)
		}
	}
	ExportData.Posts = posts

	//Answers
	answers, _ := src.GetAnswersByUser(profileID)
	for i := range answers {
		answers[i].Content = structs.Shorten(answers[i].Content, 30)
		answers[i].PostTitle = structs.Shorten(answers[i].PostTitle, 20)
	}
	ExportData.Answers = answers

	//Liked Posts
	likedPosts, _ := src.GetLikedPostsByUser(profileID)
	for i := range likedPosts {
		likedPosts[i].Title = structs.Shorten(likedPosts[i].Title, 20)
		likedPosts[i].Content = structs.Shorten(likedPosts[i].Content, 20)
		if likedPosts[i].Creator.Username != "Deleted User" {
			likedPosts[i].Creator.Username = structs.Shorten(likedPosts[i].Creator.Username, 12)
		}
	}
	ExportData.LikedPosts = likedPosts

	//Liked Answers
	likedAnswers, _ := src.GetLikedAnswersByUser(profileID)
	for i := range likedAnswers {
		likedAnswers[i].Content = structs.Shorten(likedAnswers[i].Content, 30)
		likedAnswers[i].PostTitle = structs.Shorten(likedAnswers[i].PostTitle, 20)
		if likedAnswers[i].Creator.Username != "Deleted User" {
			likedAnswers[i].Creator.Username = structs.Shorten(likedAnswers[i].Creator.Username, 12)
		}
	}
	ExportData.LikedAnswers = likedAnswers

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
