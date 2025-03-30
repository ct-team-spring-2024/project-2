package internal

type User struct {
	UserId   int64
	Username string
	Password string
}

func NewUser(userId int64, username string, password string) *User {
	return &User{
		UserId: userId,
		Username: username,
		Password: password,
	}
}
