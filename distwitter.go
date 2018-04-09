package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gorilla/securecookie"
)

const (
	usersFilePath = "./users/"
	postsFilePath = "./posts/"
	authsFilePath = "./auths"
)

var (
	authsFile, _ = os.OpenFile(authsFilePath, os.O_RDWR|os.O_CREATE, 0644)
)

func isLogin(authToken string) (*userData, error) {

	if authToken == "" {
		return nil, errors.New("No authentification token")
	}

	username := getUserIdFromAuth(authToken)
	if username == "" {
		return nil, errors.New("Wrong authentification token")
	}
	return loadUserInfo(username)
}

func loadUserInfo(username string) (*userData, error) {

	temp, err := ioutil.ReadFile(usersFilePath + username)
	var userdata *userData

	json.Unmarshal(temp, userdata)
	if err != nil {
		log.Fatal(err)
	}
	return userdata, nil
}

func register(username, password, email string) (auth string, err error) {

	if !userExists(username) {

		authToken := string(securecookie.GenerateRandomKey(32))
		followers := []string{}
		following := []string{}
		posts := []string{}

		newUser := &userData{
			Username:  username,
			Password:  password,
			AuthToken: authToken,
			Followers: followers,
			Following: following,
			Posts:     posts,
		}

		_ = writeUserDataToFile(username, newUser)

		return authToken, err

	}
	return "", fmt.Errorf("Sorry, that username/email has already been taken. Please try again!")

}

func login(username, password string) (auth string, err error) {
	fmt.Println("IN LOGIN")
	if !userExists(username) || !validUser(username, password) {
		return "", errors.New("Wrong username or password!")
	}

	auth = string(securecookie.GenerateRandomKey(32))
	if startSession(username, auth) {
		return auth, nil
	}
	return "", errors.New("Some Error Occurred!")
}

func logout(user *userData) {

	if nil == user {
		return
	}

	// newAuth := string(securecookie.GenerateRandomKey(32))
	// oldAuth, _ := redis.String(conn.Do("HGET", "user:"+user.username, "auth"))

	// _, err := conn.Do("HSET", "user:"+user.Id, "auth", newAuth)
	// if err != nil {
	// 	log.Println(err)
	// }
	// _, err = conn.Do("HSET", "auths", newAuth, user.Id)
	// if err != nil {
	// 	log.Println(err)
	// }
	// _, err = conn.Do("HDEL", "auths", oldAuth)
	// if err != nil {
	// 	log.Println(err)
	// }
}

func post(user *userData, status string) error {

	return nil
}

func getUserPosts(username string) ([]*postData, error) {

	values, err := ioutil.ReadFile(postsFilePath + username)
	if err != nil {
		log.Fatal(err)
	}
	postsByUser := []*postData{}
	json.Unmarshal(values, postsByUser)
	return postsByUser, err
}

func getUsers() ([]*userData, error) {

	users, err := ioutil.ReadDir(usersFilePath)
	if err != nil {
		log.Fatal(err)
	}
	allUsernames := []*userData{}
	for _, user := range users {
		allUsernames = append(allUsernames, &userData{Username: user.Name()})
	}
	return allUsernames, nil
}
