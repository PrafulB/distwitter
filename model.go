package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	// "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

/////////////////////////////////
type userData struct {
	Name      string
	Username  string
	Email     string
	Password  string
	AuthToken string
	Followers []string
	Following []string
	Posts     []string
}

type postData struct {
	PostId     string
	Username   string
	Content    string
	TimePosted string
}

type usersFileData struct {
	Username userData
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func userExists(username string) bool {
	if _, err := os.Stat(usersFilePath + username); os.IsNotExist(err) {
		return false
	}
	return true
}

func validUser(username, password string) bool {
	usersData, err := ioutil.ReadFile(usersFilePath + username)
	checkErr(err)
	usersDataString := string(usersData)
	storedPassword := gjson.Get(usersDataString, "Password")
	if storedPassword.String() == password {
		return true
	}
	return false
}

func startSession(username, authToken string) {
	usersData, err := ioutil.ReadFile(usersFilePath + username)
	checkErr(err)
	usersDataString := string(usersData)

	modifiedJson, err := sjson.Set(usersDataString, "AuthToken", authToken)
	checkErr(err)
	go ioutil.WriteFile(usersFilePath+username, []byte(modifiedJson), 0644)

	earlierAuth := gjson.Get(usersDataString, "AuthToken").String()
	authsData, err := ioutil.ReadFile(authsFilePath)
	checkErr(err)

	modifiedJson, err = sjson.Delete(string(authsData), earlierAuth)
	checkErr(err)

	modifiedJson, err = sjson.Set(modifiedJson, authToken, username)
	checkErr(err)

	go ioutil.WriteFile(authsFilePath, []byte(modifiedJson), 0644)

}

func writeUserDataToFile(username string, newUser *userData) {
	out, err := json.Marshal(newUser)
	checkErr(err)

	err = ioutil.WriteFile(usersFilePath+username, []byte(string(out)), 0644)
	checkErr(err)
	go ioutil.WriteFile(postsFilePath+username, []byte(`{"posts":[]}`), 0644)
}

func writeAuthDataToFile(username string, authToken string) {
	// servers := 3
	// v0Primary := GetPrimary(1, servers)
	// cfg.replicateWriteAuth(v0Primary, username, authToken, servers)
	authsData, err := ioutil.ReadFile(authsFilePath)
	modifiedJson, err := sjson.Set(string(authsData), authToken, username)
	checkErr(err)
	ioutil.WriteFile(authsFilePath, []byte(modifiedJson), 0644)
	go ioutil.WriteFile(authsFilePath, []byte(modifiedJson), 0644)
}

func writePostDataToFile(username string, newPost *postData) {
	previousPosts, err := ioutil.ReadFile(postsFilePath + username)
	checkErr(err)

	modifiedJson, err := sjson.Set(string(previousPosts), "posts.-1", newPost)
	checkErr(err)

	go ioutil.WriteFile(postsFilePath+username, []byte(modifiedJson), 0644)
}

func getUserIdFromAuth(authToken string) string {
	authsData, err := ioutil.ReadFile(authsFilePath)
	checkErr(err)
	authsDataString := string(authsData)
	username := gjson.Get(authsDataString, authToken)

	if username.Exists() {
		return username.String()
	}

	return ""
}

func getAuthFromUserId(username string) string {
	userData, err := ioutil.ReadFile(usersFilePath + username)
	checkErr(err)

	return gjson.Get(string(userData), "AuthToken").String()
}

func removeAuthToken(authToken string) {
	authsData, err := ioutil.ReadFile(authsFilePath)
	if err != nil {
		panic(err)
	}
	modifiedJson, err := sjson.Delete(string(authsData), authToken)
	go ioutil.WriteFile(authsFilePath, []byte(modifiedJson), 0644)

}
