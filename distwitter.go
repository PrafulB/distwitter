package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/tidwall/sjson"

	"github.com/tidwall/gjson"
)

const (
	usersFilePath  = "./users/"
	postsFilePath  = "./posts/"
	authsFilePath  = "./auths"
	EXPIRY_IN_SECS = 30 * 60
)

var (
	authsFile, _ = os.OpenFile(authsFilePath, os.O_RDWR|os.O_CREATE, 0644)
)
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func isLogin(authToken string) (*userData, string) {

	if authToken == "" {
		return nil, "No authentification token"
	}

	username := getUserIdFromAuth(authToken)
	if username == "" {

		return nil, "Wrong authentification token"
	}
	return loadUserInfo(username)
}

func loadUserInfo(username string) (*userData, string) {

	var userdata *userData
	temp, err := ioutil.ReadFile(usersFilePath + username)
	checkErr(err)
	err = json.Unmarshal(temp, &userdata)
	checkErr(err)
	return userdata, ""
}

func register(name, username, password, email string) (auth string, err string) {

	if !userExists(username) {

		authToken := randSeq(10)
		followers := []string{}
		following := []string{}
		posts := []string{}

		newUser := &userData{
			Name:      name,
			Username:  username,
			Password:  password,
			AuthToken: authToken,
			Followers: followers,
			Following: following,
			Posts:     posts,
		}

		go writeUserDataToFile(username, newUser)
		go writeAuthDataToFile(username, authToken)

		return authToken, ""

	}
	return "", "Sorry, that username/email has already been taken. Please try again!"
}

func login(username, password string) (string, *userData, string) {
	if !userExists(username) || !validUser(username, password) {
		return "", nil, "Wrong username or password!"
	}
	auth := randSeq(10)
	go startSession(username, auth)
	currentUserData, err := loadUserInfo(username)
	if err != "" {
		checkErr(errors.New(err))
	}
	return auth, currentUserData, ""
}

func logout(user *userData) {

	if nil == user {
		return
	}

	authToken := getAuthFromUserId(user.Username)
	go removeAuthToken(authToken)
}

func post(user *userData, status string) {
	newPost := &postData{
		PostId:     randSeq(15),
		Username:   user.Username,
		Content:    status,
		TimePosted: time.Now().String(),
	}

	go writePostDataToFile(user.Username, newPost)

}

func follow(username string, userToFollow string) {
	if userExists(userToFollow) {
		temp, err := ioutil.ReadFile(usersFilePath + username)
		checkErr(err)
		modifiedJson, err := sjson.Set(string(temp), "Following.-1", userToFollow)
		fmt.Println(modifiedJson)
		go ioutil.WriteFile(usersFilePath+username, []byte(modifiedJson), 0644)
	}
}

func delete(user *userData) {
	go logout(user)
	go os.Remove(usersFilePath + user.Username)
	go os.Remove(postsFilePath + user.Username)
}

func getUserPosts(username string) ([]postData, error) {
	user, err := loadUserInfo(username)
	allPosts := make([]postData, 0)
	checkErr(err)
	for i := range user.Following {
		values, err := ioutil.ReadFile(postsFilePath + user.Following[i])
		checkErr(err)
		postsList := gjson.Get(string(values), "posts").String()
		if postsList == "[]" {
			continue
		}
		postsByUser := make([]postData, 0)
		json.Unmarshal([]byte(postsList), &postsByUser)
		allPosts = append(allPosts, postsByUser...)
	}

	if len(allPosts) < 10 {
		users, err := getUsers()
		checkErr(err)
		userPostsToCapture := []string{}
		rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
		for i := 0; i < 5; i++ {
			if len(users) == 0 {
				break
			}
			idx := rand.Intn(len(users))
			newUser := users[idx]
			if newUser.Username != username {
				users[len(users)-1], users[idx] = users[idx], users[len(users)-1]
				users = users[:len(users)-1]
				userPostsToCapture = append(userPostsToCapture, newUser.Username)
			}
		}
		userPostsToCapture = append(userPostsToCapture, username)

		for i := range userPostsToCapture {
			values, err := ioutil.ReadFile(postsFilePath + userPostsToCapture[i])
			checkErr(err)
			postsList := gjson.Get(string(values), "posts").String()
			if postsList == "[]" {
				break
			}
			postsByUser := make([]postData, 0)
			json.Unmarshal([]byte(postsList), &postsByUser)
			allPosts = append(allPosts, postsByUser...)
		}

	}
	return allPosts, nil
}

func getUsers() ([]*userData, error) {

	users, err := ioutil.ReadDir(usersFilePath)
	checkErr(err)
	allUsernames := []*userData{}
	for _, user := range users {
		allUsernames = append(allUsernames, &userData{Username: user.Name()})
	}
	return allUsernames, nil
}
