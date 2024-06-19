package handlers

import (
	"Forum/src"
	"Forum/src/structs"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type administrationPageData struct {
	User       structs.User
	Categories []structs.Category
	Posts      []structs.Post
	Answers    []structs.Answer
	Users      []structs.User
}

func serveAdministrationPage(w http.ResponseWriter, r *http.Request) {

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	ExportData := administrationPageData{}
	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)
	// Il faut minimum 1 de pouvoir pour avoir accès au panel admin, sinon page d'erreur
	if user.Power == 0 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}

	ExportData.User = user

	// Récupère toutes les catégories
	categories, err := src.GetAllCategories()
	if err != nil {
		http.Error(w, "Unable to retrieve categories", http.StatusInternalServerError)
	}
	ExportData.Categories = categories

	// Récupère tout les posts
	posts, err := src.GetAllPosts()
	if err != nil {
		http.Error(w, "Unable to retrieve posts", http.StatusInternalServerError)
	}
	ExportData.Posts = posts

	// Récupère tout les réponses/commentaires
	answers, err := src.GetAllAnswers()
	if err != nil {
		http.Error(w, "Unable to retrieve answers", http.StatusInternalServerError)
	}
	ExportData.Answers = answers

	// Récupère tout les utilisateurs
	users, err := src.GetAllUsers()
	if err != nil {
		http.Error(w, "Unable to retrieve users", http.StatusInternalServerError)
	}
	ExportData.Users = users

	tmpl := template.Must(template.ParseFiles("src/templates/administration.html"))
	tmpl.Execute(w, ExportData)
}

func addCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)

	if user.Power != 2 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}
	name := "Name"
	description := "Description"
	i := 1
	baseName := name
	message := "Couldn't add category."
	defaultImage := "/static/image/default-category-image.jpg"
	// Ajouts de la category dans la base de donnée
	_, err := src.Db.Exec("INSERT INTO categories (name, description, image) VALUES (?, ?, ?)", name, description, defaultImage)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintPrimaryKey) {
			name = fmt.Sprintf("%s%d", baseName, i)
			i++
		}
	} else {
		message = "Category added successfully."
	}
	http.Redirect(w, r, "/administration.html?message="+message, http.StatusSeeOther)
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)

	if user.Power != 2 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}
	categoryName := r.URL.Path[len("/delete-category/"):]
	// Supprime la catégory de la base de donnée
	category, _ := src.GetCategory(categoryName)
	_, err := src.Db.Exec("DELETE FROM categories WHERE name = ?", categoryName)
	if err != nil {
		http.Error(w, "Unable to delete category", http.StatusInternalServerError)
		return
	}
	//Supprime l'image de la catégory
	if category.Image != "/static/image/default-category-image.jpg" {
		oldImagePath := fmt.Sprintf("src%s", category.Image)
		err := os.Remove(oldImagePath)
		if err != nil {
			http.Error(w, "Failed to delete old image", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/administration.html?message=Category%20deleted%20successfully.", http.StatusSeeOther)
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	// Limit la taille de la ram utilisé à 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	user, _ := src.GetUserFromSessionID(sessionID)

	if user.Power != 2 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}
	categoryName := r.URL.Path[len("/update-category/"):]

	prevName := r.FormValue("prevName")
	category, _ := src.GetCategory(prevName)
	newName := r.FormValue("newName")
	newDescription := r.FormValue("newDescription")
	newImage := ""

	if newName == "" || newDescription == "" {
		http.Error(w, "New name and description are required", http.StatusInternalServerError)
		return
	}

	file, header, err := r.FormFile("newImage")
	// S'il y a une image
	if err == nil {
		defer file.Close()

		if category.Image != "/static/image/default-category-image.jpg" {
			oldImagePath := fmt.Sprintf("src%s", category.Image)
			// Supprime l'ancienne image (sauf si c'est l'image de référence)
			err := os.Remove(oldImagePath)
			if err != nil {
				http.Error(w, "Failed to delete old image", http.StatusInternalServerError)
				return
			}
		}
		imagePath := fmt.Sprintf("src/static/image/categories/%s-image%s", newName, filepath.Ext(header.Filename))
		// Sauvegarde la nouvelle image
		// Creation de l'image
		out, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Unable to upload image", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		// Transfère des données de l'image
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Unable to save image", http.StatusInternalServerError)
			return
		}
		newImage = fmt.Sprintf("/static/image/categories/%s-image%s", newName, filepath.Ext(header.Filename))
	} else {
		// Si le nom a été modifié
		if prevName != newName {
			// Rename de l'image
			if category.Image != "/static/image/default-category-image.jpg" {
				oldImagePath := fmt.Sprintf("src%s", category.Image)
				newImagePath := fmt.Sprintf("src/static/image/categories/%s-image%s", newName, filepath.Ext(category.Image))
				err := os.Rename(oldImagePath, newImagePath)
				if err != nil {
					http.Error(w, "Failed to rename image", http.StatusInternalServerError)
					return
				}
				newImage = fmt.Sprintf("/static/image/categories/%s-image%s", newName, filepath.Ext(category.Image))
			}
		}
	}
	// Update des données dans la bdd
	query := "UPDATE categories SET name = ?, description = ?"
	args := []interface{}{newName, newDescription}

	if newImage != "" {
		query += ", image = ?"
		args = append(args, newImage)
	}
	// Ajouts de l'image au Quey si modifié
	query += " WHERE name = ?"
	args = append(args, categoryName)
	_, err = src.Db.Exec(query, args...)
	if err != nil {
		http.Error(w, "Unable to update category", http.StatusInternalServerError)
		return
	}
	// On met à jour les posts aussi
	if prevName != newName {
		_, err = src.Db.Exec(`
			UPDATE posts
			SET category_name = ?
			WHERE category_name = ?
		`, newName, prevName)
		if err != nil {
			return
		}
	}

	// Récupère la position du scroll pour la restaurer
	scrollPos := r.FormValue("scrollPos")
	redirectURL := fmt.Sprintf("/administration.html?scrollPos=%s&message=%s", scrollPos, "Category updated successfully.")
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	postID := r.URL.Path[len("/delete-post/"):]

	user, _ := src.GetUserFromSessionID(sessionID)
	if user.Power == 0 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}
	// Supprime le post de la base de donnée
	post, _ := src.GetPost(postID)
	_, err := src.Db.Exec("DELETE FROM posts WHERE uuid = ?", postID)
	if err != nil {
		http.Error(w, "Unable to delete post", http.StatusInternalServerError)
		return
	}
	// S'il y a des images, on les supprimes tous
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
	// Récupère la position du scroll pour la restaurer
	scrollPos := r.FormValue("scrollPos")
	redirectURL := fmt.Sprintf("/administration.html?scrollPos=%s&section=%s", scrollPos, "posts-section")
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
func deleteAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	answerID := r.URL.Path[len("/delete-answer/"):]

	user, _ := src.GetUserFromSessionID(sessionID)
	if user.Power == 0 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}
	// Supprime la réponse/commentaire de la bdd
	_, err := src.Db.Exec("DELETE FROM answers WHERE uuid = ?", answerID)
	if err != nil {
		http.Error(w, "Unable to answer post", http.StatusInternalServerError)
		return
	}
	// Récupère la position du scroll pour la restaurer
	scrollPos := r.FormValue("scrollPos")
	redirectURL := fmt.Sprintf("/administration.html?scrollPos=%s&section=%s", scrollPos, "answers-section")
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
func deleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	userID := r.URL.Path[len("/delete-user/"):]

	user, _ := src.GetUserFromSessionID(sessionID)
	if user.Power != 2 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}
	// Supprime utilisateur de la bdd
	_, err := src.Db.Exec("DELETE FROM users WHERE uuid = ?", userID)
	if err != nil {
		http.Error(w, "Unable to user post", http.StatusInternalServerError)
		return
	}
	// Récupère la position du scroll pour la restaurer
	scrollPos := r.FormValue("scrollPos")
	redirectURL := fmt.Sprintf("/administration.html?scrollPos=%s&section=%s", scrollPos, "users-section")
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
func updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Empeche la création de cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// S'il n'y a pas de cookie de session, on le redirige vers la page de connexion
	if !cookieExists(r, "sessionID") {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	// Si le backend ne reconnait pas l'ID de la session, on retire le cookie et on le redirige vers la page de connexion
	sessionID := src.GetValidSession(r)
	if sessionID == "" {
		w, r = removeSession(w, r)
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	userID := r.URL.Path[len("/update-user/"):]

	user, _ := src.GetUserFromSessionID(sessionID)
	if user.Power != 2 {
		http.ServeFile(w, r, "src/templates/error.html")
		return
	}
	// Récupère l'utilisateur
	targetUUID, _ := uuid.Parse(userID)
	target, _ := src.GetUserFromUUID(targetUUID)
	// Change sa puissance 0 -> 1 -> 2 -> 0
	resultPower := 0
	if target.Power == 0 {
		resultPower = 1
	} else if target.Power == 1 {
		resultPower = 2
	} else if target.Power == 2 {
		resultPower = 0
	}
	_, err := src.Db.Exec("UPDATE users SET power = ? WHERE uuid = ?", resultPower, userID)
	if err != nil {
		http.Error(w, "Unable to update user", http.StatusInternalServerError)
		return
	}
	// Récupère la position du scroll pour la restaurer
	scrollPos := r.FormValue("scrollPos")
	redirectURL := fmt.Sprintf("/administration.html?scrollPos=%s&section=%s", scrollPos, "users-section")
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
