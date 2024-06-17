package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type postCreationPageData struct {
	User     structs.User
	Category structs.Category
}

func servePostCreatePage(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/creation-post.html"))

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	ExportData := postCreationPageData{}
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		logoutHandler(w, r)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)
	ExportData.User = user

	categoryString := r.URL.Query().Get("category")
	if categoryString == "" {
		http.Redirect(w, r, "/error.html", http.StatusFound)
		return
	}

	category, err := src.GetCategory(categoryString)
	if err != nil {
		http.Redirect(w, r, "/error.html", http.StatusFound)
		return
	}
	ExportData.Category = category

	tmpl.Execute(w, ExportData)
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		logoutHandler(w, r)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Redirect(w, r, "/error.html", http.StatusFound)
		return
	} // limit upload size to 32MB

	categoryString := r.FormValue("category")
	category, err := src.GetCategory(categoryString)
	if err != nil {
		http.Redirect(w, r, "/error.html", http.StatusFound)
		return
	}
	title := r.FormValue("post-title")
	content := r.FormValue("post-content")

	// Validate title and content
	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}
	post := structs.Post{
		Uuid:         uuid.Must(uuid.NewV4()).String(),
		Title:        title,
		Content:      content,
		Creator:      user.Username,
		Category:     category.Name,
		CreationDate: time.Now().Format("2006-01-02 15:04:05"),
	}
	// Handle file uploads
	files := r.MultipartForm.File["post-images"]
	totalSize := int64(0)
	uploadedFiles := make([]string, 0)
	for _, fileHeader := range files {
		if fileHeader.Size > 20*1024*1024 {
			sendErrorResponse(w, "Each image should not exceed 20MB", http.StatusBadRequest)
			return
		}
		totalSize += fileHeader.Size
		if totalSize > 20*1024*1024 {
			sendErrorResponse(w, "Total size of images should not exceed 20MB", http.StatusInternalServerError)
			return
		}
	}
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Unable to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		filename := fmt.Sprintf("%s-%s", post.Uuid, fileHeader.Filename)
		filepath := fmt.Sprintf("src/static/image/posts/%s", filename)
		outFile, err := os.Create(filepath)
		filepath = fmt.Sprintf("/static/image/posts/%s", filename)
		if err != nil {
			http.Error(w, "Unable to upload file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, file)
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		uploadedFiles = append(uploadedFiles, filepath)
	}

	post.Images = uploadedFiles

	// Save the post in the database
	likeStrings := make([]string, len(post.Likes))
	for i, like := range post.Likes {
		likeStrings[i] = like.String()
	}

	dislikeStrings := make([]string, len(post.Dislikes))
	for i, dislike := range post.Dislikes {
		dislikeStrings[i] = dislike.String()
	}

	// Convert likes, dislikes, and images to JSON
	likesJSON, err := json.Marshal(likeStrings)
	if err != nil {
		http.Error(w, "Unable to post post", http.StatusInternalServerError)
		return
	}

	dislikesJSON, err := json.Marshal(dislikeStrings)
	if err != nil {
		http.Error(w, "Unable to post post", http.StatusInternalServerError)
		return
	}

	imagesJSON, err := json.Marshal(post.Images)
	if err != nil {
		http.Error(w, "Unable to post post", http.StatusInternalServerError)
		return
	}
	stmt, err := src.Db.Prepare(`INSERT INTO posts (uuid, title, content, owner_id, category_name, created_at, likes, dislikes, images)
                             VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(post.Uuid, post.Title, post.Content, post.Creator, post.Category, post.CreationDate, likesJSON, dislikesJSON, imagesJSON)
	if err != nil {
		http.Error(w, "Failed to execute SQL statement", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"redirect": "/categories/" + category.Name + "?message=Post%20posted%20successfully%20!"})
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
