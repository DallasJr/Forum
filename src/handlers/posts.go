package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
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

type postPageData struct {
	User    structs.User
	Post    structs.Post
	Answers []structs.Answer
}

func servePostCreatePage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/creation-post.html"))

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	ExportData := postCreationPageData{}
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
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
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
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
	if len(title) > 100 || len(title) < 10 {
		http.Error(w, "Title incorrect length", http.StatusBadRequest)
		return
	}
	if len(content) > 2500 || len(content) < 50 {
		http.Error(w, "Content incorrect length", http.StatusBadRequest)
		return
	}

	post := structs.Post{
		Uuid:         uuid.New(),
		Title:        title,
		Content:      content,
		CreatorUUID:  user.Uuid,
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
	_, err = stmt.Exec(post.Uuid, post.Title, post.Content, post.CreatorUUID, post.Category, post.CreationDate, likesJSON, dislikesJSON, imagesJSON)
	if err != nil {
		http.Error(w, "Failed to execute SQL statement", http.StatusInternalServerError)
		return
	}
	//http.Redirect(w, r, "/post/"+post.Uuid.String()+"?p-message=Post%20posted%20successfully%20!", http.StatusSeeOther)
	json.NewEncoder(w).Encode(map[string]string{"redirect": "/post/" + post.Uuid.String() + "?p-message=Post%20posted%20successfully%20!"})
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/post" || r.URL.Path == "/post/" {
		index(w, r)
		return
	} else {
		servePostPage(w, r)
	}
}

func servePostPage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	id := strings.TrimPrefix(r.URL.Path, "/post/")
	/*if id == "" {
		http.NotFound(w, r)
		return
	}*/

	ExportData := postPageData{}

	if cookieExists(r, "sessionID") {
		sessionID := src.GetValidSession(r)
		if sessionID == "" {
			w, r = removeSession(w, r)
		}
		user, _ := src.GetUserFromSessionID(sessionID)
		ExportData.User = user
	}

	post, _ := src.GetPost(id)
	ExportData.Post = post

	//Answers
	answers, _ := src.GetAnswersByPosts(id)

	ExportData.Answers = answers

	funcMap := template.FuncMap{
		"userHasLiked":    userHasLiked,
		"userHasDisliked": userHasDisliked,
	}
	t := template.Must(template.New("post.html").Funcs(funcMap).ParseFiles("src/templates/post.html"))
	t.Execute(w, ExportData)
}

func userHasLiked(userUUID uuid.UUID, likes []uuid.UUID) bool {
	for _, like := range likes {
		if like == userUUID {
			return true
		}
	}
	return false
}

func userHasDisliked(userUUID uuid.UUID, dislikes []uuid.UUID) bool {
	for _, dislike := range dislikes {
		if dislike == userUUID {
			return true
		}
	}
	return false
}

func handleAnswerSubmission(w http.ResponseWriter, r *http.Request) {
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
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)

	postString := r.FormValue("post")
	post, err := src.GetPost(postString)
	if err != nil {
		http.Redirect(w, r, "/error.html", http.StatusFound)
		return
	}
	content := r.FormValue("answer-content")

	// Validate title and content
	if content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}
	if len(content) > 1000 || len(content) < 2 {
		http.Error(w, "Content incorrect length", http.StatusBadRequest)
		return
	}

	answer := structs.Answer{
		Uuid:         uuid.New(),
		Content:      content,
		CreatorUUID:  user.Uuid,
		PostID:       post.Uuid,
		CreationDate: time.Now().Format("2006-01-02 15:04:05"),
	}

	likeStrings := make([]string, len(answer.Likes))
	for i, like := range answer.Likes {
		likeStrings[i] = like.String()
	}

	dislikeStrings := make([]string, len(answer.Dislikes))
	for i, dislike := range answer.Dislikes {
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

	stmt, err := src.Db.Prepare(`INSERT INTO answers (uuid, content, owner_id, post_id, created_at, likes, dislikes)
                             VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(answer.Uuid, answer.Content, answer.CreatorUUID, answer.PostID, answer.CreationDate, likesJSON, dislikesJSON)
	if err != nil {
		http.Error(w, "Failed to execute SQL statement", http.StatusInternalServerError)
		return
	}
	scrollPos := r.FormValue("scrollPos")
	redirectURL := fmt.Sprintf("/post/%s?a-message=Answer%%20posted%%20successfully%%20!&scrollPos=%s", post.Uuid.String(), scrollPos)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func handlePostDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	postID := r.URL.Path[len("/delete-own-post/"):]
	post, _ := src.GetPost(postID)

	user, _ := src.GetUserFromSessionID(sessionID)
	if user.Uuid != post.CreatorUUID {
		http.Error(w, "Unable to delete post", http.StatusInternalServerError)
		return
	}
	// Delete the post from the database
	_, err := src.Db.Exec("DELETE FROM posts WHERE uuid = ?", postID)
	if err != nil {
		http.Error(w, "Unable to delete post", http.StatusInternalServerError)
		return
	}
	if len(post.Images) != 0 {
		for _, j := range post.Images {
			oldImagePath := fmt.Sprintf("src%s", j)
			err := os.Remove(oldImagePath)
			if err != nil {
				http.Error(w, "Failed to delete old image", http.StatusInternalServerError)
				return
			}
		}
	}
	redirectURL := fmt.Sprintf("/categories/%s", post.Category)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func handleAnswerDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	answerID := r.URL.Path[len("/delete-own-answer/"):]
	answer, _ := src.GetAnswer(answerID)

	user, _ := src.GetUserFromSessionID(sessionID)
	if user.Uuid != answer.CreatorUUID {
		http.Error(w, "Unable to delete answer", http.StatusInternalServerError)
		return
	}
	// Delete the answer from the database
	_, err := src.Db.Exec("DELETE FROM answers WHERE uuid = ?", answerID)
	if err != nil {
		http.Error(w, "Unable to answer post", http.StatusInternalServerError)
		return
	}
	scrollPos := r.FormValue("scrollPos")
	profile := r.FormValue("profile")

	var redirectURL string

	if profile == "true" {
		// Redirect to profile page or another appropriate location
		redirectURL = "/profile/" + user.Uuid.String()
		if scrollPos != "" {
			redirectURL += "?scrollPos=" + scrollPos
		}
	} else {
		// Default redirect to the post page with scroll position
		redirectURL = fmt.Sprintf("/post/%s", answer.PostID)
		if scrollPos != "" {
			redirectURL += "?scrollPos=" + scrollPos
		}
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func handleLikeDislike(w http.ResponseWriter, r *http.Request) {
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
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	user, err := src.GetUserFromSessionID(sessionID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}
	postID := r.Form.Get("postID")
	answerID := r.Form.Get("answerID")
	action := r.Form.Get("action")

	var likes []string
	var likesJSON string
	if postID != "" {
		err = src.Db.QueryRow("SELECT likes FROM posts WHERE uuid = ?", postID).Scan(&likesJSON)
		if err != nil {
			http.Error(w, "Database error1", http.StatusInternalServerError)
			return
		}
	} else {
		err = src.Db.QueryRow("SELECT likes FROM answers WHERE uuid = ?", answerID).Scan(&likesJSON)
		if err != nil {
			http.Error(w, "Database error2", http.StatusInternalServerError)
			return
		}
	}
	err = json.Unmarshal([]byte(likesJSON), &likes)
	if err != nil {
		http.Error(w, "Failed to parse likes", http.StatusInternalServerError)
		return
	}

	var dislikes []string
	var dislikesJSON string
	if postID != "" {
		err = src.Db.QueryRow("SELECT dislikes FROM posts WHERE uuid = ?", postID).Scan(&dislikesJSON)
		if err != nil {
			http.Error(w, "Database error3", http.StatusInternalServerError)
			return
		}
	} else {
		err = src.Db.QueryRow("SELECT dislikes FROM answers WHERE uuid = ?", answerID).Scan(&dislikesJSON)
		if err != nil {
			http.Error(w, "Database error4", http.StatusInternalServerError)
			return
		}
	}
	err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
	if err != nil {
		http.Error(w, "Failed to parse likes", http.StatusInternalServerError)
		return
	}

	if action == "like" {
		index := indexOf(likes, user.Uuid.String())
		if index == -1 {
			likes = append(likes, user.Uuid.String())
			index2 := indexOf(dislikes, user.Uuid.String())
			if index2 != -1 {
				dislikes = append(dislikes[:index2], dislikes[index2+1:]...)
			}
		} else {
			likes = append(likes[:index], likes[index+1:]...)
		}
	} else if action == "dislike" {
		index := indexOf(dislikes, user.Uuid.String())
		if index == -1 {
			dislikes = append(dislikes, user.Uuid.String())
			index2 := indexOf(likes, user.Uuid.String())
			if index2 != -1 {
				likes = append(likes[:index2], likes[index2+1:]...)
			}
		} else {
			dislikes = append(dislikes[:index], dislikes[index+1:]...)
		}
	} else {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Convert likes, dislikes, and images to JSON
	updatedLikesJSON, err := json.Marshal(likes)
	if err != nil {
		http.Error(w, "Unable to marshal likes", http.StatusInternalServerError)
		return
	}

	updatedDislikesJSON, err := json.Marshal(dislikes)
	if err != nil {
		http.Error(w, "Unable to marshal dislikes", http.StatusInternalServerError)
		return
	}

	// Update likes in the database
	if postID != "" {
		_, err = src.Db.Exec("UPDATE posts SET likes = ? WHERE uuid = ?", updatedLikesJSON, postID)
		if err != nil {
			http.Error(w, "Database error5", http.StatusInternalServerError)
			return
		}
		_, err = src.Db.Exec("UPDATE posts SET dislikes = ? WHERE uuid = ?", updatedDislikesJSON, postID)
		if err != nil {
			http.Error(w, "Database error6", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = src.Db.Exec("UPDATE answers SET likes = ? WHERE uuid = ?", updatedLikesJSON, answerID)
		if err != nil {
			http.Error(w, "Database error7", http.StatusInternalServerError)
			return
		}
		_, err = src.Db.Exec("UPDATE answers SET dislikes = ? WHERE uuid = ?", updatedDislikesJSON, answerID)
		if err != nil {
			http.Error(w, "Database error8", http.StatusInternalServerError)
			return
		}
	}

	scrollPos := r.FormValue("scrollPos")
	var id string
	if postID != "" {
		id = postID
	} else {
		answer, _ := src.GetAnswer(answerID)
		id = answer.PostID.String()
	}
	redirectURL := fmt.Sprintf("/post/%s?scrollPos=%s", id, scrollPos)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func indexOf(slice []string, val string) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return -1
}
