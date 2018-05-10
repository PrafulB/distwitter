package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	// "strconv"

	"github.com/gorilla/mux"
)

var cfg *configWA
var primaryID int
var servers int

func Index(w http.ResponseWriter, r *http.Request) {
	setupResponse(w, r)
	u, err1 := isLogin(getAuth(r))
	if "" != err1 {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User unauthenticated yet."))
		return
	} else {
		success := map[string]interface{}{}
		success["user"] = u
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(success)
		checkErr(err)
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	u, err1 := isLogin(getAuth(r))
	if "" != err1 {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User unauthenticated yet."))
		return
	}
	posts, err := getUserPosts(u.Username)
	checkErr(err)
	postsResp := map[string]interface{}{}

	postsResp["posts"] = posts
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(postsResp)
	checkErr(err)
}

func Register(w http.ResponseWriter, r *http.Request) {
	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	name := r.PostFormValue("name")
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	auth, err := register(name, username, password, email)
	if err != "" {
		goBack := map[string]string{
			"error": err,
		}
		fmt.Println(err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(goBack)
		return
	}

	setSession(auth, w)
	success := map[string]string{
		"name":     name,
		"username": username,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(success)
}

func Login(w http.ResponseWriter, r *http.Request) {
	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	auth, currentUserData, err := login(username, password)

	if err != "" {
		goBack := map[string]string{
			"error": err,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(goBack)
		return
	}
	setSession(auth, w)
	success := map[string]string{
		"name":     currentUserData.Name,
		"username": currentUserData.Username,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(success)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	u, err := isLogin(getAuth(r))
	if "" != err {
		goBack := map[string]string{
			"error": err,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(goBack)
		return
	}
	go logout(u)
	success := map[string]string{
		"message": "Logged Out",
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(success)
}

func Post(w http.ResponseWriter, r *http.Request) {
	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	u, err := isLogin(getAuth(r))
	if "" != err {
		goBack := map[string]string{
			"error": err,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(goBack)
		return
	}
	content := r.PostFormValue("status")
	go post(u, content)
	success := map[string]string{
		"message": "Post Successful",
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(success)
}

func Follow(w http.ResponseWriter, r *http.Request) {
	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	u, err := isLogin(getAuth(r))
	if "" != err {
		goBack := map[string]string{
			"error": err,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(goBack)
		return
	}
	userToFollow := r.PostFormValue("username")
	go follow(u.Username, userToFollow)
	success := map[string]string{
		"message": "Follower Added",
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(success)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	setupResponse(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	u, err := isLogin(getAuth(r))
	if "" != err {
		goBack := map[string]string{
			"error": err,
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(goBack)
		return
	}
	go delete(u)
	success := map[string]string{
		"message": "User Deleted",
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(success)
}

func setupResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func temp(*configWA) {

}

func main() {

	go os.MkdirAll(usersFilePath, 0755)
	go os.MkdirAll(postsFilePath, 0755)

	servers := 3
	cfg := make_config_write_auth(servers, false)
	temp(cfg)
	// if f.Size() == 0 {
	// 	userNamesFile.WriteString("[]")
	// }

	var router = mux.NewRouter()
	// router := httprouter.New()
	router.HandleFunc("/", Index)
	router.HandleFunc("/home", Home)
	router.HandleFunc("/register", Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/post", Post).Methods("POST", "OPTIONS")
	router.HandleFunc("/logout", Logout)
	router.HandleFunc("/delete", Delete).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/follow", Follow).Methods("POST", "OPTIONS")

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
