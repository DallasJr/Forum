package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"encoding/json"
	"github.com/google/uuid"
	"html/template"
	"net/http"
	"strconv"
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

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	profileID := strings.TrimPrefix(r.URL.Path, "/profile/")
	/*if id == "" {
		http.NotFound(w, r)
		return
	}*/

	ExportData := profilePageData{}

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
	posts, _ := src.GetPostsByUser(profileID, 0, 5)
	for i := range posts {
		posts[i].Title = structs.Shorten(posts[i].Title, 20)
		posts[i].Content = structs.Shorten(posts[i].Content, 20)
		if posts[i].Creator.Username != "Deleted User" {
			posts[i].Creator.Username = structs.Shorten(posts[i].Creator.Username, 12)
		}
	}
	ExportData.Posts = posts

	//Answers
	answers, _ := src.GetAnswersByUser(profileID, 0, 5)
	for i := range answers {
		answers[i].Content = structs.Shorten(answers[i].Content, 30)
		answers[i].PostTitle, _ = src.GetPostNameByPostID(answers[i].PostID.String())
		answers[i].PostTitle = structs.Shorten(answers[i].PostTitle, 20)
	}
	ExportData.Answers = answers

	//Liked Posts
	likedPosts, _ := src.GetLikedPostsByUser(profileID, 0, 5)
	for i := range likedPosts {
		likedPosts[i].Title = structs.Shorten(likedPosts[i].Title, 20)
		likedPosts[i].Content = structs.Shorten(likedPosts[i].Content, 20)
		if likedPosts[i].Creator.Username != "Deleted User" {
			likedPosts[i].Creator.Username = structs.Shorten(likedPosts[i].Creator.Username, 12)
		}
	}
	ExportData.LikedPosts = likedPosts

	//Liked Answers
	likedAnswers, _ := src.GetLikedAnswersByUser(profileID, 0, 5)
	for i := range likedAnswers {
		likedAnswers[i].Content = structs.Shorten(likedAnswers[i].Content, 30)
		likedAnswers[i].PostTitle, _ = src.GetPostNameByPostID(likedAnswers[i].PostID.String())
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

func showMorePostedAnswers(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("profile")
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}
	answers, err := src.GetAnswersByUser(userID, offset, 5)
	if err != nil {
		http.Error(w, "Unable to retrieve more answers", http.StatusInternalServerError)
		return
	}
	for i := range answers {
		answers[i].Content = structs.Shorten(answers[i].Content, 20)
		answers[i].PostTitle, _ = src.GetPostNameByPostID(answers[i].PostID.String())
		answers[i].PostTitle = structs.Shorten(answers[i].PostTitle, 20)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answers)
}

func showMoreCreatedPosts(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("profile")
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}
	posts, err := src.GetPostsByUser(userID, offset, 5)
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

func showMoreLikedAnswers(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("profile")
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}
	answers, err := src.GetLikedAnswersByUser(userID, offset, 5)
	if err != nil {
		http.Error(w, "Unable to retrieve more answers", http.StatusInternalServerError)
		return
	}
	for i := range answers {
		answers[i].Content = structs.Shorten(answers[i].Content, 20)
		answers[i].PostTitle, _ = src.GetPostNameByPostID(answers[i].PostID.String())
		answers[i].PostTitle = structs.Shorten(answers[i].PostTitle, 20)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answers)
}

func showMoreLikedPosts(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("profile")
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}
	posts, err := src.GetLikedPostsByUser(userID, offset, 5)
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
