package goforces

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

// TODO
func LogUserIn(username, password string) bool {
	//get username row in DB and its hased password and then compare the two hashed passwords
	user := GetUserFromDB(username)
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		slog.Error(fmt.Sprintf("Error logging in  => %v", err))
		return false
	}
	return true
}
func SignUpUser(username, password string) bool {
	//TODO : check if the username exists
	pass, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		slog.Error(fmt.Sprintf("Error encypting the pass word =>   %v", err))
		return false
	}
	user := User{Username: username, Password: string(pass)}
	Error := SaveUserInDB(user)
	if Error != nil {
		slog.Error(fmt.Sprintf("Error in signing user up => %v", err))
		return false
	}
	return true
}

// TODO : Complete this function when the DB is ready
func GetUserFromDB(username string) User {
	return User{Username: "", Password: ""}
}
func SaveUserInDB(user User) error {
	return nil

}
