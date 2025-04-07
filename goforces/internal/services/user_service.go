package services

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"oj/goforces/internal/models"

	"github.com/dgrijalva/jwt-go"
)

var (
	users     = make(map[string]models.User)
	userMutex = &sync.Mutex{}
	// TODO: use uuid
	userIDCounter = 1

	jwtKey = []byte("your_secret_key")
)

func RegisterUser(u models.User) (models.User, error) {
	userMutex.Lock()
	defer userMutex.Unlock()

	if _, exists := users[u.Email]; exists {
		return models.User{}, errors.New("user already exists")
	}

	u.UserId = userIDCounter
	userIDCounter++
	// TODO: hash the password
	users[u.Email] = u
	return u, nil
}

// TODO: Move to auth package
func AuthenticateUser(email, password string) (string, error) {
	userMutex.Lock()
	user, exists := users[email]
	userMutex.Unlock()

	if !exists || user.Password != password {
		return "", errors.New("invalid credentials")
	}

	tokenString, err := generateToken(user)
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
	for _, user := range users {
		if user.UserId == id {
			return user, nil
		}
	}
	return models.User{}, errors.New("user not found")
}

func UpdateUserProfile(id int, updated models.User) (models.User, error) {
	userMutex.Lock()
	defer userMutex.Unlock()
	for email, user := range users {
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
			users[email] = user
			return user, nil
		}
	}
	return models.User{}, errors.New("user not found")
}
