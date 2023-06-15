package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var users []User

func main() {
	loadUsers()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/google-login", handleGoogleLogin)
	http.HandleFunc("/google-callback", handleGoogleCallback)

	fmt.Println("Le serveur est en cours d'exécution sur http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if isValidUser(username, password) {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		} else {
			fmt.Fprintf(w, "<p>Mauvais nom d'utilisateur ou mot de passe.</p>")
			return
		}
	}

	http.ServeFile(w, r, "index.html")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if !isExistingUser(username) {
			users = append(users, User{Username: username, Password: password})
			saveUsers()
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	http.ServeFile(w, r, "index.html")
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("dashboard.html"))
	tmpl.Execute(w, nil)
}

func isValidUser(username, password string) bool {
	for _, user := range users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

func isExistingUser(username string) bool {
	for _, user := range users {
		if user.Username == username {
			return true
		}
	}
	return false
}

func loadUsers() {
	file, err := ioutil.ReadFile("users.json")
	if err != nil {
		return
	}
	json.Unmarshal(file, &users)
}

func saveUsers() {
	file, _ := json.MarshalIndent(users, "", "  ")
	_ = ioutil.WriteFile("users.json", file, 0644)
}

// Configuration OAuth2 pour Google
var googleOauthConfig = &oauth2.Config{
	ClientID:     "20636667204-7ga8gi3k8p48uedp4lr8ceok33b3ht6s.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURL:  "http://localhost:8000/google-callback",
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
	},
	Endpoint: google.Endpoint,
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}

	client := googleOauthConfig.Client(oauth2.NoContext, token)
	response, err := client.Get("https://www.googleapis.com/userinfo/v2/me")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var userInfo struct {
		Email string `json:"email"`
	}

	err = json.Unmarshal(data, &userInfo)
	if err != nil {
		log.Fatal(err)
	}

	if isValidUser(userInfo.Email, "") {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	fmt.Fprintf(w, "<p>Mauvaises informations d'identification.</p>")
}
