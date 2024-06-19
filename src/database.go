package src

import (
	"Forum/src/structs"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sort"
	"sync"
)

var (
	Sessions = make(map[string]string)
	Mutex    sync.Mutex
)

var Db *sql.DB

func SetupDatabase() *sql.DB {
	var err error
	Db, err = sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}

	// Users
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			surname TEXT NOT NULL,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			gender TEXT NOT NULL,
			power INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Posts
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			uuid TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			category_name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			likes TEXT,
			dislikes TEXT,
			images TEXT 
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Answers
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS answers (
			uuid TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			post_id TEXT NOT NULL,
			created_at TEXT NOT NULL,
			likes TEXT,
			dislikes TEXT,
			FOREIGN KEY(owner_id) REFERENCES users(uuid),
			FOREIGN KEY(post_id) REFERENCES posts(uuid)
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Categories
	_, err = Db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
    		name TEXT PRIMARY KEY,
    		description TEXT,
			image TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	return Db
}

func GetUserFromSessionID(sessionID string) (structs.User, error) {
	Mutex.Lock()
	userID, exists := Sessions[sessionID]
	Mutex.Unlock()
	if !exists {
		return structs.User{}, nil
	}
	var user structs.User
	query := `SELECT uuid, name, surname, username, email, gender, created_at, power FROM users WHERE uuid = ?`
	err := Db.QueryRow(query, userID).Scan(
		&user.Uuid, &user.Name, &user.Surname, &user.Username,
		&user.Email, &user.Gender, &user.CreationDate, &user.Power,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return structs.User{}, nil
		}
		fmt.Println("query error:", err)
		return structs.User{}, err
	}
	return user, nil
}

func GetUserFromUUID(userID uuid.UUID) (structs.User, error) {
	var user structs.User
	query := `SELECT uuid, name, surname, username, email, gender, created_at, power FROM users WHERE uuid = ?`
	err := Db.QueryRow(query, userID).Scan(
		&user.Uuid, &user.Name, &user.Surname, &user.Username,
		&user.Email, &user.Gender, &user.CreationDate, &user.Power,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return structs.User{}, nil
		}
		fmt.Println("query error:", err)
		return structs.User{}, err
	}
	return user, nil
}

func GetValidSession(r *http.Request) string {
	c, _ := r.Cookie("sessionID")
	Mutex.Lock()
	_, exists := Sessions[c.Value]
	Mutex.Unlock()
	if !exists {
		return ""
	} else {
		return c.Value
	}
}

func GetCategory(name string) (structs.Category, error) {
	var category structs.Category

	query := `SELECT name, description, image FROM categories WHERE name = ?`
	err := Db.QueryRow(query, name).Scan(
		&category.Name, &category.Description, &category.Image,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return structs.Category{}, nil
		}
		fmt.Println("query error:", err)
		return structs.Category{}, err
	}
	return category, nil
}

func GetPostsByCategory(categoryName string) ([]structs.Post, error) {
	rows, err := Db.Query("SELECT uuid, title, content, owner_id, category_name, created_at, likes, dislikes FROM posts WHERE category_name = ? ORDER BY created_at DESC", categoryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		var likesJSON, dislikesJSON string
		if err := rows.Scan(&post.Uuid, &post.Title, &post.Content, &post.CreatorUUID, &post.Category, &post.CreationDate, &likesJSON, &dislikesJSON); err != nil {
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		post.Likes = likes
		post.Dislikes = dislikes
		answersCount, _ := GetAnswersCountByPost(post.Uuid)
		post.AnswersCount = answersCount

		formattedDate, err := post.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}

		post.Creator, _ = GetUserFromUUID(post.CreatorUUID)
		if post.Creator.Username == "" {
			post.Creator = structs.User{Username: "Deleted User"}
		}

		post.CreationDate = formattedDate
		posts = append(posts, post)
	}
	return posts, nil
}

func GetPostsCountByCategory(categoryName string) (int, error) {
	rows, err := Db.Query("SELECT COUNT(*) FROM posts WHERE category_name = ?", categoryName)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func GetRecentPosts() ([]structs.Post, error) {
	var posts []structs.Post
	rows, err := Db.Query(`
	        SELECT uuid, title, content, owner_id, category_name, created_at, likes, dislikes
	        FROM posts
	        ORDER BY created_at DESC
	        LIMIT 5
	    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var post structs.Post
		var likesJSON, dislikesJSON string
		err := rows.Scan(&post.Uuid, &post.Title, &post.Content, &post.CreatorUUID, &post.Category, &post.CreationDate, &likesJSON, &dislikesJSON)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		post.Likes = likes
		post.Dislikes = dislikes

		formattedDate, err := post.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}

		post.Creator, _ = GetUserFromUUID(post.CreatorUUID)
		if post.Creator.Username == "" {
			post.Creator = structs.User{Username: "Deleted User"}
		}

		post.CreationDate = formattedDate
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetPost(id string) (structs.Post, error) {
	var post structs.Post
	var likesJSON, dislikesJSON, imagesJSON string

	query := `
				SELECT uuid, title, content, owner_id, category_name, created_at, likes, dislikes, images
				FROM posts
				WHERE uuid = ?`
	err := Db.QueryRow(query, id).Scan(
		&post.Uuid, &post.Title, &post.Content, &post.CreatorUUID, &post.Category, &post.CreationDate, &likesJSON, &dislikesJSON, &imagesJSON,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return structs.Post{}, nil
		}
		fmt.Println("query error:", err)
		return structs.Post{}, err
	}

	var likes []uuid.UUID
	var dislikes []uuid.UUID
	var images []string

	err = json.Unmarshal([]byte(likesJSON), &likes)
	if err != nil {
		fmt.Println("Error unmarshaling likes:", err)
		return structs.Post{}, err
	}

	err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
	if err != nil {
		fmt.Println("Error unmarshaling dislikes:", err)
		return structs.Post{}, err
	}

	err = json.Unmarshal([]byte(imagesJSON), &images)
	if err != nil {
		fmt.Println("Error unmarshaling images:", err)
		return structs.Post{}, err
	}

	post.Likes = likes
	post.Dislikes = dislikes
	post.Images = images

	formattedDate, err := post.FormattedDate()
	if err != nil {
		fmt.Println("Error formatting date:", err)
	}

	post.Creator, _ = GetUserFromUUID(post.CreatorUUID)
	if post.Creator.Username == "" {
		post.Creator = structs.User{Username: "Deleted User"}
	}

	post.CreationDate = formattedDate
	return post, nil
}

func GetAnswersByPosts(post string) ([]structs.Answer, error) {
	rows, err := Db.Query("SELECT uuid, content, owner_id, post_id, created_at, likes, dislikes FROM answers WHERE post_id = ? ORDER BY created_at", post)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []structs.Answer
	for rows.Next() {
		var answer structs.Answer
		var likesJSON, dislikesJSON string
		if err := rows.Scan(&answer.Uuid, &answer.Content, &answer.CreatorUUID, &answer.PostID, &answer.CreationDate, &likesJSON, &dislikesJSON); err != nil {
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		answer.Likes = likes
		answer.Dislikes = dislikes
		answer.PostTitle, _ = GetPostNameByPostID(answer.PostID.String())
		answer.Creator, _ = GetUserFromUUID(answer.CreatorUUID)
		if answer.Creator.Username == "" {
			answer.Creator = structs.User{Username: "Deleted User"}
		}
		answer.CreationDate, err = answer.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}
		answers = append(answers, answer)
	}
	return answers, nil
}

func GetAnswersCountByPost(post uuid.UUID) (int, error) {
	rows, err := Db.Query("SELECT COUNT(*) FROM answers WHERE post_id = ?", post)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func GetPostsByUser(userID string) ([]structs.Post, error) {
	rows, err := Db.Query("SELECT uuid, title, content, owner_id, category_name, created_at, likes, dislikes FROM posts WHERE owner_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		var likesJSON, dislikesJSON string
		if err := rows.Scan(&post.Uuid, &post.Title, &post.Content, &post.CreatorUUID, &post.Category, &post.CreationDate, &likesJSON, &dislikesJSON); err != nil {
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		post.Likes = likes
		post.Dislikes = dislikes
		answersCount, _ := GetAnswersCountByPost(post.Uuid)
		post.AnswersCount = answersCount

		formattedDate, err := post.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}

		post.Creator, _ = GetUserFromUUID(post.CreatorUUID)
		if post.Creator.Username == "" {
			post.Creator = structs.User{Username: "Deleted User"}
		}

		post.CreationDate = formattedDate
		posts = append(posts, post)
	}
	return posts, nil
}

func GetLikedPostsByUser(userID string) ([]structs.Post, error) {
	rows, err := Db.Query("SELECT uuid, title, content, owner_id, category_name, created_at, likes, dislikes FROM posts WHERE likes LIKE CONCAT('%', ?, '%') ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		var likesJSON, dislikesJSON string
		if err := rows.Scan(&post.Uuid, &post.Title, &post.Content, &post.CreatorUUID, &post.Category, &post.CreationDate, &likesJSON, &dislikesJSON); err != nil {
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		post.Likes = likes
		post.Dislikes = dislikes
		answersCount, _ := GetAnswersCountByPost(post.Uuid)
		post.AnswersCount = answersCount

		formattedDate, err := post.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}

		post.Creator, _ = GetUserFromUUID(post.CreatorUUID)
		if post.Creator.Username == "" {
			post.Creator = structs.User{Username: "Deleted User"}
		}

		post.CreationDate = formattedDate
		posts = append(posts, post)
	}
	return posts, nil
}

func GetPostNameByPostID(postID string) (string, error) {
	rows, err := Db.Query("SELECT title FROM posts WHERE uuid = ? ", postID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var title string
	if rows.Next() {
		if err := rows.Scan(&title); err != nil {
			return "", err
		}
	}
	return title, nil
}

func GetAnswersByUser(userID string) ([]structs.Answer, error) {
	rows, err := Db.Query("SELECT uuid, content, owner_id, post_id, created_at, likes, dislikes FROM answers WHERE owner_id = ? ORDER BY created_at", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []structs.Answer
	for rows.Next() {
		var answer structs.Answer
		var likesJSON, dislikesJSON string
		if err := rows.Scan(&answer.Uuid, &answer.Content, &answer.CreatorUUID, &answer.PostID, &answer.CreationDate, &likesJSON, &dislikesJSON); err != nil {
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		answer.Likes = likes
		answer.Dislikes = dislikes
		answer.PostTitle, _ = GetPostNameByPostID(answer.PostID.String())

		formattedDate, err := answer.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}

		answer.Creator, _ = GetUserFromUUID(answer.CreatorUUID)
		if answer.Creator.Username == "" {
			answer.Creator = structs.User{Username: "Deleted User"}
		}

		answer.CreationDate = formattedDate
		answers = append(answers, answer)
	}
	return answers, nil
}

func GetLikedAnswersByUser(userID string) ([]structs.Answer, error) {
	rows, err := Db.Query("SELECT uuid, content, owner_id, post_id, created_at, likes, dislikes FROM answers WHERE likes LIKE CONCAT('%', ?, '%') ORDER BY created_at", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []structs.Answer
	for rows.Next() {
		var answer structs.Answer
		var likesJSON, dislikesJSON string
		if err := rows.Scan(&answer.Uuid, &answer.Content, &answer.CreatorUUID, &answer.PostID, &answer.CreationDate, &likesJSON, &dislikesJSON); err != nil {
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		answer.Likes = likes
		answer.Dislikes = dislikes

		formattedDate, err := answer.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}

		answer.Creator, _ = GetUserFromUUID(answer.CreatorUUID)
		if answer.Creator.Username == "" {
			answer.Creator = structs.User{Username: "Deleted User"}
		}

		answer.CreationDate = formattedDate
		answers = append(answers, answer)
	}
	return answers, nil
}

func GetAllCategories() ([]structs.Category, error) {
	rows, err := Db.Query("SELECT name, description, image FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []structs.Category
	for rows.Next() {
		var category structs.Category
		if err := rows.Scan(&category.Name, &category.Description, &category.Image); err != nil {
			return nil, err
		}

		postCount, err := GetPostsCountByCategory(category.Name)
		if err != nil {
			return nil, err
		}
		category.PostsCount = postCount

		categories = append(categories, category)
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].PostsCount > categories[j].PostsCount
	})
	return categories, nil
}

func GetAllAnswers() ([]structs.Answer, error) {
	rows, err := Db.Query("SELECT uuid, content, owner_id, post_id, created_at, likes, dislikes FROM answers ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []structs.Answer
	for rows.Next() {
		var answer structs.Answer
		var likesJSON, dislikesJSON string
		if err := rows.Scan(&answer.Uuid, &answer.Content, &answer.CreatorUUID, &answer.PostID, &answer.CreationDate, &likesJSON, &dislikesJSON); err != nil {
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		answer.Likes = likes
		answer.Dislikes = dislikes
		answer.PostTitle, _ = GetPostNameByPostID(answer.PostID.String())

		formattedDate, err := answer.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}

		answer.Creator, _ = GetUserFromUUID(answer.CreatorUUID)
		if answer.Creator.Username == "" {
			answer.Creator = structs.User{Username: "Deleted User"}
		}

		answer.CreationDate = formattedDate
		answers = append(answers, answer)
	}
	return answers, nil
}

func GetAllPosts() ([]structs.Post, error) {
	rows, err := Db.Query("SELECT uuid, title, content, owner_id, category_name, created_at, likes, dislikes, images FROM posts ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []structs.Post
	for rows.Next() {
		var post structs.Post
		var likesJSON, dislikesJSON, imagesJSON string
		if err := rows.Scan(&post.Uuid, &post.Title, &post.Content, &post.CreatorUUID, &post.Category, &post.CreationDate, &likesJSON, &dislikesJSON, &imagesJSON); err != nil {
			return nil, err
		}
		var likes []uuid.UUID
		var dislikes []uuid.UUID
		var images []string

		err = json.Unmarshal([]byte(likesJSON), &likes)
		if err != nil {
			fmt.Println("Error unmarshaling likes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
		if err != nil {
			fmt.Println("Error unmarshaling dislikes:", err)
			return nil, err
		}

		err = json.Unmarshal([]byte(imagesJSON), &images)
		if err != nil {
			fmt.Println("Error unmarshaling images:", err)
			return nil, err
		}

		post.Likes = likes
		post.Dislikes = dislikes
		post.Images = images
		answersCount, _ := GetAnswersCountByPost(post.Uuid)
		post.AnswersCount = answersCount

		formattedDate, err := post.FormattedDate()
		if err != nil {
			fmt.Println("Error formatting date:", err)
		}

		post.Creator, _ = GetUserFromUUID(post.CreatorUUID)
		if post.Creator.Username == "" {
			post.Creator = structs.User{Username: "Deleted User"}
		}

		post.CreationDate = formattedDate
		posts = append(posts, post)
	}
	return posts, nil
}

func GetAllUsers() ([]structs.User, error) {
	rows, err := Db.Query("SELECT uuid, name, surname, username, email, gender, power, created_at FROM users ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []structs.User
	for rows.Next() {
		var user structs.User
		if err := rows.Scan(&user.Uuid, &user.Name, &user.Surname, &user.Username, &user.Email, &user.Gender, &user.Power, &user.CreationDate); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetAnswer(id string) (structs.Answer, error) {
	var answer structs.Answer
	var likesJSON, dislikesJSON string
	query := `SELECT uuid, content, owner_id, post_id, created_at, likes, dislikes FROM answers WHERE uuid = ?`
	err := Db.QueryRow(query, id).Scan(
		&answer.Uuid, &answer.Content, &answer.CreatorUUID, &answer.PostID, &answer.CreationDate, &likesJSON, &dislikesJSON,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return structs.Answer{}, nil
		}
		fmt.Println("query error:", err)
		return structs.Answer{}, err
	}

	var likes []uuid.UUID
	var dislikes []uuid.UUID

	err = json.Unmarshal([]byte(likesJSON), &likes)
	if err != nil {
		fmt.Println("Error unmarshaling likes:", err)
		return structs.Answer{}, err
	}

	err = json.Unmarshal([]byte(dislikesJSON), &dislikes)
	if err != nil {
		fmt.Println("Error unmarshaling dislikes:", err)
		return structs.Answer{}, err
	}

	answer.Likes = likes
	answer.Dislikes = dislikes
	answer.PostTitle, _ = GetPostNameByPostID(answer.PostID.String())

	formattedDate, err := answer.FormattedDate()
	if err != nil {
		fmt.Println("Error formatting date:", err)
	}

	answer.Creator, _ = GetUserFromUUID(answer.CreatorUUID)
	if answer.Creator.Username == "" {
		answer.Creator = structs.User{Username: "Deleted User"}
	}

	answer.CreationDate = formattedDate
	return answer, nil
}
