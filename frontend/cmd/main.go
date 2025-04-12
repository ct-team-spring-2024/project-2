package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
	//"github.com/golang-jwt/jwt/v5"
	"html/template"
	"net/http"
	"oj/frontend"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

var store = sessions.NewCookieStore([]byte("a-very-secret-key"))

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})

	logrus.Debug("This is a debug message")

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//router.LoadHTMLGlob("templates/*")
	router.SetFuncMap(template.FuncMap{
		"add":       func(a, b int) int { return a + b },
		"minus":     func(a, b int) int { return a - b },
		"pageRange": pageRange,
	})

	router.LoadHTMLGlob("C:/Users/Asus/Documents/GitHub/project-2/frontend/templates/*")
	router.Static("/static", "C:/Users/Asus/Documents/GitHub/project-2/frontend/static")

	// router.GET("/", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "index.html", nil)
	// })
	//Log in page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	//Problems page

	router.GET("/problems.html", func(c *gin.Context) {
		pageNo := c.DefaultQuery("page", "1")
		limitNo := c.DefaultQuery("limit", "20")
		authenticate(c)
		fmt.Println("canme here")
		tokensring, _ := store.Get(c.Request, "session-name")
		fmt.Println(tokensring)

		page, _ := strconv.Atoi(pageNo)
		limit, _ := strconv.Atoi(limitNo)
		if page < 1 {
			page = 1
		}
		allProblems := getProblems(1, 1000)
		total := len(allProblems)
		totalPages := (total + limit - 1) / limit // round up

		if page > totalPages {
			page = totalPages
		}
		start := (page - 1) * limit
		end := start + limit
		if end > total {
			end = total
		}
		paged := allProblems[start:end]
		pageData := frontend.PageData{
			Problems:    paged,
			CurrentPage: page,
			Limit:       limit,
			HasNextPage: page < totalPages,
			TotalPages:  totalPages,
		}
		fmt.Println("came here")
		c.HTML(http.StatusOK, "problems.html", pageData)

	})
	router.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		if username == "" {
			logrus.Error("Error : username is empty")
		}
		if username == "username" && password == "password" {
			session, _ := store.Get(c.Request, "session-name")
			session.Values["username"] = username
			//TODO : add JWT to tokens for better security

			tokenString := createToken()

			session.Values["jwt"] = tokenString

			session.Save(c.Request, c.Writer)

			c.HTML(http.StatusOK, "user_page.html", nil)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		}

	})
	router.Run(":8080")
}
func getProblems(pageNumber int, limit int) []frontend.Problem {
	//Complete after DB
	var problems = make([]frontend.Problem, 0)
	for i := pageNumber * limit; i < (pageNumber+1)*limit; i++ {
		problems = append(problems, frontend.Problem{Title: "first problem", Id: i})
	}
	return problems
}
func pageRange(current, total int) []int {
	var pages []int
	start := current - 1
	if start < 1 {
		start = 1
	}
	end := start + 2
	if end > total {
		end = total
		start = end - 2
		if start < 1 {
			start = 1
		}
	}
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}
func createToken() string {
	claims := MyCustomClaims{
		UserID: 123,

		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "123", // typically user ID
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("secret-key"))
	return tokenString
}

func authenticate(c *gin.Context) bool {
	session, _ := store.Get(c.Request, "session-name")
	jwtToken, ok := session.Values["jwt"].(string)

	if !ok || jwtToken == "" {
		// http.Redirect(w, r, "/login", http.StatusFound)
		fmt.Println("came here1")
		return false
	}

	// optionally validate the JWT if needed
	claims := &MyCustomClaims{}
	_, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret-key"), nil

	})
	if err != nil {
		http.Redirect(c.Writer, c.Request, "/index.html", http.StatusFound)

		return false

	}
	return true
}

type MyCustomClaims struct {
	UserID               int    `json:"user_id"`
	Email                string `json:"email"`
	Role                 string `json:"role"`
	jwt.RegisteredClaims        // include standard claims like exp, iat, sub
}
