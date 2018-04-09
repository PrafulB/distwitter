package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	// "strconv"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

var (
	templateRender = render.New(render.Options{
		Layout:        "layout",
		IsDevelopment: true,
	})
)

func Index(w http.ResponseWriter, r *http.Request) {
	templateParams := map[string]interface{}{}
	u, err := isLogin(getAuth(r))
	fmt.Println(err, u)
	templateParams["user"] = u
	if nil != err {
		templateRender.HTML(w, http.StatusOK, "welcome", templateParams)
	} else {
		http.Redirect(w, r, "/home", http.StatusFound)
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	templateParams := map[string]interface{}{}
	u, err := isLogin(getAuth(r))
	if nil != err {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	templateParams["user"] = u

	posts, err := getUserPosts(u.Username)
	if err != nil {
		log.Fatal(err)
	}

	templateParams["posts"] = posts

	templateRender.HTML(w, http.StatusOK, "home", templateParams)
}

func Register(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	password2 := r.PostFormValue("password2")
	if username == "" || email == "" || password == "" || password2 == "" {
		GoBack(w, r, errors.New("Every field of the registration form is needed!"))
		return
	}
	if password != password2 {
		GoBack(w, r, errors.New("The two password fileds don't match!"))
		return
	}
	auth, err := register(username, password, email)
	if err != nil {
		GoBack(w, r, err)
		return
	}
	setSession(auth, w)
	templateParams := map[string]interface{}{}
	templateParams["username"] = username
	templateRender.HTML(w, http.StatusOK, "register", templateParams)
}

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if username == "" || password == "" {
		GoBack(w, r, errors.New("You need to enter both username and password to login."))
		return
	}
	auth, err := login(username, password)

	if err != nil {
		GoBack(w, r, err)
		return
	}
	setSession(auth, w)
	fmt.Println("Trying Redirect")
	http.Redirect(w, r, "/", http.StatusFound)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	u, err := isLogin(getAuth(r))
	if nil != err {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	logout(u)
	http.Redirect(w, r, "/", http.StatusFound)
}

func GoBack(w http.ResponseWriter, r *http.Request, err error) {
	templateParams := map[string]interface{}{}
	templateParams["error"] = err
	templateRender.HTML(w, http.StatusOK, "error", templateParams)
}

func main() {

	os.MkdirAll(usersFilePath, 0755)
	os.MkdirAll(postsFilePath, 0755)

	// if f.Size() == 0 {
	// 	userNamesFile.WriteString("[]")
	// }

	var router = mux.NewRouter()
	// router := httprouter.New()
	router.HandleFunc("/", Index)
	router.HandleFunc("/home", Home)
	router.HandleFunc("/register", Register).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")
	router.HandleFunc("/logout", Logout)

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
