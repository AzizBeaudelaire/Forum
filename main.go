package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/script.js", func(w http.ResponseWriter, r *http.Request) {
		js, err := ioutil.ReadFile("script.js")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/javascript")
		fmt.Fprint(w, string(js))
	})

	r.HandleFunc("/changement_pdp.js", func(w http.ResponseWriter, r *http.Request) {
		js, err := ioutil.ReadFile("changement_pdp.js")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/javascript")
		fmt.Fprint(w, string(js))
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			html, err := ioutil.ReadFile("./templates/html/login_page.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			css, err := ioutil.ReadFile("templates/css/style.css")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			image, err := ioutil.ReadFile("./static/images/BG.jpg")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<html><head><title>Login Page</title><style>%s</style></head><body style=\"background-image: url('data:image/png;base64,%s')\">%s</body></html>", string(css), base64.StdEncoding.EncodeToString(image), string(html))
		} else if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			username := r.Form.Get("username")
			password := r.Form.Get("password")
			email := r.Form.Get("email")

			// Create a new User instance
			user := User{
				Username: username,
				Email:    email,
				Password: password,
			}

			// Read existing user data from JSON file
			users, err := readUsersFromFile("users.json")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Add the new user to the existing users slice
			users = append(users, user)

			// Write the updated users data to the JSON file
			err = writeUsersToFile(users, "users.json")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Redirect to a success page
			http.Redirect(w, r, "/success", http.StatusSeeOther)
		}
	})

	r.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		html, err := ioutil.ReadFile("./templates/html/success.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<html><head><title>Success Page</title></head>%s</body></html>", string(html))
	})

	r.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		users, err := readUsersFromFile("users.json")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, user := range users {
			fmt.Fprintf(w, "Username: %s, Email: %s, Password: %s\n", user.Username, user.Email, user.Password)
		}
	})

	fs := http.FileServer(http.Dir("./templates/css/"))
	r.PathPrefix("/templates/css/").Handler(http.StripPrefix("/templates/css/", fs))

	fmt.Printf("Server is running on http://localhost:8080/\n")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func readUsersFromFile(filename string) ([]User, error) {
	filePath := filepath.Join("users.json")
	file, err := os.Open(filePath)
	if err != nil {
		// If the file doesn't exist, return an empty slice of users
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}
	defer file.Close()

	// Decode the JSON data into a slice of User
	var users []User
	err = json.NewDecoder(file).Decode(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func writeUsersToFile(users []User, filename string) error {
	filePath := filepath.Join("users.json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the users slice as JSON and write it to the file
	err = json.NewEncoder(file).Encode(users)
	if err != nil {
		return err
	}

	return nil
}
