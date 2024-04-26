package handlers

import (
	"Forum/src"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	surname := r.FormValue("surname")
	name := r.FormValue("name")
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	gender := r.FormValue("gender")
	uniqueId := uuid.Must(uuid.NewV4())
	userID := uniqueId.String()

	hashedPassword, er := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if er != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	_, err := src.Db.Exec("INSERT INTO users (id, name, surname, username, email, password, admin, gender) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userID, name, surname, username, email, hashedPassword, false, gender)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	/*g := gender == "male"
	user := structs.User{uniqueId, name, surname, username, email, password, g, time.Now(), false}
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Couldn't save data")
		return
	}
	file, err := os.Create("data.json")
	if err != nil {
		fmt.Println("Couldn't save data")
		return
	}
	defer file.Close()
	_, err = file.Write(jsonBytes)
	if err != nil {
		fmt.Println("Couldn't save data")
	}*/

	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}

func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "src/templates/login.html")
}

func checkUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	var count int
	err := src.Db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		http.Error(w, "Failed to check username availability", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		fmt.Fprintf(w, "available")
	} else {
		fmt.Fprintf(w, "not available")
	}
}

func checkEmailAvailability(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var count int
	err := src.Db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		http.Error(w, "Failed to check email availability", http.StatusInternalServerError)
		return
	}
	if count == 0 {
		fmt.Fprintf(w, "available")
	} else {
		fmt.Fprintf(w, "not available")
	}
}
