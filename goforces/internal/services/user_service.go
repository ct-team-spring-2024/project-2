package services

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"oj/goforces/internal/db"
	"oj/goforces/internal/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
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

func RegisterUser(u models.User) (int, error) {
	userMutex.Lock()
	defer userMutex.Unlock()

	user := findWithEmail(u.Email)
	if user != nil {
		logrus.Errorf("Cannot register userrr %v", u)
		return -1, errors.New("user already exists")
	}
	// TODO: hash the password
	id, err := db.DB.CreateUser(u)
	if err != nil {
		logrus.Errorf("Cannot register user %v", err)
	}
	return id, nil
}

// TODO: Move to auth package
func AuthenticateUser(email, password string) (string, error) {
	userMutex.Lock()
	user := findWithEmail(email)
	userMutex.Unlock()

	if user == nil || user.Password != password {
		return "", errors.New("invalid credentials")
	}

	tokenString, err := GenerateToken(*user)
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

	tokenString, err := GenerateToken(*user)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GenerateToken(user models.User) (string, error) {
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
		logrus.Infof("AHHHH %+v", user)
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
