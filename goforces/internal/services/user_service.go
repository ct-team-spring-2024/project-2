package services

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"oj/goforces/internal/db"
	"oj/goforces/internal/models"

	"github.com/dgrijalva/jwt-go"
)

var (
	// users     = make(map[string]models.User)

	userMutex = &sync.Mutex{}
	// TODO: use uuid
	userIDCounter = 0

	jwtKey = []byte("your_secret_key")
)

func findWithEmail(email string) *models.User {
	for _, u := range db.DB.GetUsers() {
		if u.Email == email {
			return &u
		}
	}
	return nil
}

func findWithUsername(username string) *models.User {
	for _, u := range db.DB.GetUsers() {
		if u.Username == username {
			return &u
		}
	}
	return nil
}

func RegisterUser(u models.User) (models.User, error) {
	userMutex.Lock()
	defer userMutex.Unlock()

	user := findWithEmail(u.Email)
	if user != nil {
		return models.User{}, errors.New("user already exists")
	}
	//TODO : this must be synced
	//userIDCounter++
	// TODO: hash the password
	db.DB.CreateUser(u)
	return u, nil
}

// TODO: Move to auth package
func AuthenticateUser(email, password string) (string, error) {
	userMutex.Lock()
	user := findWithEmail(email)
	userMutex.Unlock()

	if user == nil || user.Password != password {
		return "", errors.New("invalid credentials")
	}

	tokenString, err := generateToken(*user)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func AuthenticateUserWithUsername(username, password string) (string, error) {
	userMutex.Lock()
	user := findWithUsername(username)
	userMutex.Unlock()

	if user == nil || user.Password != password {
		return "", errors.New("invalid credentials")
	}

	tokenString, err := generateToken(*user)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func generateToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		Subject:   strconv.Itoa(user.UserId),
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GetUserByID(id int) (models.User, error) {
	userMutex.Lock()
	defer userMutex.Unlock()
	for _, user := range db.DB.GetUsers() {
		if user.UserId == id {
			return user, nil
		}
	}
	return models.User{}, errors.New("user not found")
}

func GetUserByUsername(username string) (models.User, error) {
	userMutex.Lock()
	defer userMutex.Unlock()
	for _, user := range db.DB.GetUsers() {
		if user.Username == username {
			return user, nil
		}
	}
	return models.User{}, errors.New("user not found")
}

func UpdateUserProfile(id int, updated models.User) (models.User, error) {
	userMutex.Lock()
	defer userMutex.Unlock()
	for _, user := range db.DB.GetUsers() {
		if user.UserId == id {

			if updated.Username != "" {
				user.Username = updated.Username
			}
			if updated.Password != "" {
				// TODO: hash the password
				user.Password = updated.Password
			}
			if updated.Role != "" {
				user.Role = updated.Role
			}
			db.DB.UpdateUsers(user.UserId, user)
			return user, nil
		}
	}
	return models.User{}, errors.New("user not found")
}
