package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	// "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

/////////////////////////////////
type userData struct {
	Username  string
	Email     string
	Password  string
	AuthToken string
	Followers []string
	Following []string
	Posts     []string
}

type postData struct {
	Username   string
	PostId     string
	Content    string
	TimePosted string
}

type usersFileData struct {
	Username userData
}

func userExists(username string) bool {
	if _, err := os.Stat(usersFilePath + username); os.IsNotExist(err) {
		return false
	}
	return true

}

func validUser(username, password string) bool {
	usersData, err := ioutil.ReadFile(usersFilePath + username)
	if err != nil {
		log.Fatal(err)
	}
	usersDataString := string(usersData)
	storedPassword := gjson.Get(usersDataString, "Password")
	// dataOfUser := []byte(usersDataString[i : j+2])
	// fmt.Println(j, usersDataString[i:j+2], storedPassword.String())
	// var jsonData userData
	// err := json.Unmarshal(dataOfUser, &jsonData)
	// if err != nil {
	// }
	// fmt.Println(jsonData.Password, password)
	if storedPassword.String() == password {
		return true
	}
	return false
}

func startSession(username, authToken string) bool {
	fmt.Println("IN STARTSESSION")
	userFile, err := os.OpenFile(usersFilePath+username, os.O_RDWR, 0644)
	usersData, err := ioutil.ReadFile(usersFilePath + username)
	if err != nil {
		log.Fatal(err)
	}
	usersDataString := string(usersData)
	earlierAuth := gjson.Get(usersDataString, "AuthToken")

	modifiedJson, err := sjson.Set(usersDataString, "AuthToken", authToken)
	if err != nil {
		log.Fatal(err)
	}
	userFile.WriteString(modifiedJson)
	authsData, err := ioutil.ReadFile(authsFilePath)
	authsDataString := string(authsData)
	modifiedAuthsData := strings.Replace(authsDataString, earlierAuth.String(), authToken, -1)
	fmt.Println(earlierAuth.String(), "\n-----------\n", authToken, "\n-----------\n", modifiedAuthsData)
	authsFile.WriteString(modifiedAuthsData)
	return true

}

func writeUserDataToFile(username string, newUser *userData) bool {
	userFile, err := os.OpenFile(usersFilePath+username, os.O_RDWR|os.O_CREATE, 0644)
	postsFile, err := os.OpenFile(postsFilePath+username, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatal(err)
	}

	// x := &userData{
	// 	Username:  newUser.Username,
	// 	Email:     newUser.Email,
	// 	Password:  newUser.Password,
	// 	AuthToken: newUser.AuthToken,
	// 	Followers: newUser.Followers,
	// 	Following: newUser.Following,
	// 	Posts:     newUser.Posts,
	// }

	out, err := json.Marshal(newUser)
	if err != nil {
		log.Fatal(err)
	}

	userFile.WriteString(string(out))
	postsFile.WriteString("[]")

	// authsFile.WriteString("{\"auth\":\"" + newUser.AuthToken + "\",\"username\":\"" + username + "\"}\n")

	return true
}

func getUserIdFromAuth(authToken string) string {
	authsData, err := ioutil.ReadFile(authsFilePath)
	if err != nil {
		log.Fatal(err)
	}
	authsDataString := string(authsData)
	username := gjson.Get(authsDataString, "..#[auth ==\""+authToken+"\"]")

	// fmt.Println(authsDataString, authToken, username.String())
	if username.Exists() {
		return username.String()
	}
	// dataOfUser := []byte(usersDataString[i : j+2])
	// fmt.Println(j, usersDataString[i:j+2], storedPassword.String())
	// var jsonData userData
	// err := json.Unmarshal(dataOfUser, &jsonData)
	// if err != nil {
	// }
	// fmt.Println(jsonData.Password, password)f
	return ""

}
