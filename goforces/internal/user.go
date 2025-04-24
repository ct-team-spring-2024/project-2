package internal

type User struct {
	UserId   int
	Username string
	Password string
}

func NewUser(userId int, username string, password string) *User {
	return &User{
		UserId:   userId,
		Username: username,
		Password: password,
	}
}
