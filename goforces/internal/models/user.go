package models

type User struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` //either admin or user
}
