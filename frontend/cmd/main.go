package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	ppath "path"
	"strconv"
	"time"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"

	"oj/frontend"
)

var store = sessions.NewCookieStore([]byte("a-very-secret-key"))

// var path = "C:/Users/Asus/Documents/GitHub/project-2"
var path = "/home/mbroughani81/Documents/test/computer-technology-project-2"

func logStringError(body []byte) {
	logrus.Warnf("ERROR => %s", string(body))
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})

	logrus.Debug("This is a debug message")

	router := gin.Default()
	router.HTMLRender = ginview.Default()

	staticPath := fmt.Sprintf("%s/static", ppath.Base("."))
	router.Static("/static", staticPath)

	// router.SetFuncMap(template.FuncMap{
	//	"add":       func(a, b int) int { return a + b },
	//	"minus":     func(a, b int) int { return a - b },
	//	"pageRange": pageRange,
	// })

	backendUrl := "http://localhost:8080"
	//Log in page
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.POST("/login", func(c *gin.Context) {
		// request
		var loginData struct {
			Username string `form:"username"`
			Password string `form:"password"`
		}

		if err := c.ShouldBind(&loginData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}
		payloadBytes, err := json.Marshal(loginData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
			return
		}

		// send to backend
		loginUrl := fmt.Sprintf("%s/login", backendUrl)
		resp, err := http.Post(loginUrl, "application/json", bytes.NewReader(payloadBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact backend"})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read backend response"})
			return
		}

		var result struct {
			Token string `json:"token"`
		}
		// TODO: the response should have a structure. A simple string is not good.
		if err := json.Unmarshal(body, &result); err != nil || result.Token == "" {
			logrus.Infof("Result => %+v", result)
			logStringError(body)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid response from backend"})
			return
		}

		session, err := store.Get(c.Request, "session-name")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
			return
		}
		session.Values["username"] = loginData.Username
		session.Values["jwt"] = result.Token
		if err := session.Save(c.Request, c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}

		c.Redirect(http.StatusFound, fmt.Sprintf("/profile/%s", loginData.Username))
	})

	router.GET("/profile/:username", func(c *gin.Context) {
		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)
		username := c.Param("username")
		logrus.Infof("clientUsername => %s", clientUsername)
		logrus.Infof("username => %s", username)

		submissions := []frontend.Submission{
			{
				ProblemName:    "Sorting Algorithm",
				Status:         "Accepted",
				SubmissionDate: "2023-09-01",
			},
			{
				ProblemName:    "Binary Search",
				Status:         "Wrong Answer",
				SubmissionDate: "2023-09-02",
			},
			{
				ProblemName:    "Linked List Manipulation",
				Status:         "Accepted",
				SubmissionDate: "2023-09-03",
			},
		}

		pageData := frontend.ProfilePageData{
			Page:             "profile",
			ClientUsername:   clientUsername,
			IsClientAdmin:    clientUsername == "admin",
			IsUserAdmin:      username == "admin",
			Submissions:      submissions,
			CurrentPage:      1,
			Limit:            10,
			HasNextPage:      true,
			TotalPages:       5,
			Username:         username,
			Email:            "johndoe@example.com",
			MemberSince:      "January 2023",
			TotalSubmissions: 50,
			SolvedProblems:   30,
			SolveRate:        60,
		}
		c.HTML(http.StatusOK, "profile", pageData)
	})

	router.GET("/problems", func(c *gin.Context) {
		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)
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
		logrus.Infof("paged %+v", paged)
		pageData := frontend.ProblemsPageData{
			Page:           "problems",
			ClientUsername: clientUsername,
			IsClientAdmin:  clientUsername == "admin",
			Problems:       paged,
			CurrentPage:    page,
			Limit:          limit,
			HasNextPage:    page < totalPages,
			TotalPages:     totalPages,
		}

		fmt.Println("came here")
		c.HTML(http.StatusOK, "problems", pageData)
	})

	router.GET("/problem/:id", func(c *gin.Context) {
		// Extract the problem ID from the URL
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
			return
		}

		// Fetch the problem details
		problem, err := getProblemByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
			return
		}

		// Render the problem.html template with the problem data
		c.HTML(http.StatusOK, "problem.html", problem)
	})

	router.GET("/add-problem", func(c *gin.Context) {
		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)
		pageData := frontend.AddProblemPageData{
			Page:           "add-problem",
			ClientUsername: clientUsername,
			IsClientAdmin:  clientUsername == "admin",
		}
		c.HTML(http.StatusOK, "add-problem", pageData)
	})
	router.POST("/add-problem", func(c *gin.Context) {
		// Parse the multipart form (with a max memory of 32MB for file uploads)
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse form data"})
			return
		}

		// Extract form fields
		problemTitle := c.PostForm("problemTitle")
		timeLimit := c.PostForm("timeLimit")
		memoryLimit := c.PostForm("memoryLimit")
		problemDescription := c.PostForm("problemDescription")

		// Extract the uploaded test case file
		file, fileHeader, err := c.Request.FormFile("testCaseFile")
		var testCaseContent string
		if err == nil {
			defer file.Close()
			// Read the file content
			content, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to read test case file"})
				return
			}
			testCaseContent = string(content)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to retrieve test case file"})
			return
		}

		// Print the extracted data
		fmt.Printf("Problem Title: %s\n", problemTitle)
		fmt.Printf("Time Limit: %s ms\n", timeLimit)
		fmt.Printf("Memory Limit: %s MB\n", memoryLimit)
		fmt.Printf("Problem Description: %s\n", problemDescription)
		fmt.Printf("Test Case File Name: %s\n", fileHeader.Filename)
		fmt.Printf("Test Case Content: %s\n", testCaseContent)

		// Respond with success
		c.JSON(http.StatusOK, gin.H{"message": "Problem data received and printed"})
	})
	router.GET("/my-problems", func(c *gin.Context) {
		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)
		problems := []frontend.ProblemSummary{
			{
				Id:     "1",
				Title:  "Two Sum",
				Status: "published",
			},
			{
				Id:     "2",
				Title:  "Fibonacci Sequence",
				Status: "draft",
			},
			{
				Id:     "3",
				Title:  "Binary Search",
				Status: "published",
			},
		}
		pageData := frontend.MyProblemsPageData{
			Page:           "my-problems",
			ClientUsername: clientUsername,
			IsClientAdmin:  clientUsername == "admin",
			Problems:       problems,
		}

		c.HTML(http.StatusOK, "my-problems", pageData)
	})
	router.GET("/edit/:id", func(c *gin.Context) {
		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)

		// Extract the problem ID from the URL
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
			return
		}

		// Fetch the problem details
		problem, err := getProblemByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
			return
		}

		editProblemPageData := frontend.EditProblemPageData{
			Page:           "edit-problem",
			ClientUsername: clientUsername,
			IsClientAdmin:  clientUsername == "admin",
			Problem:        problem,
		}
		logrus.Infof("Problem => %+v", editProblemPageData.Problem)

		// Render the problem.html template with the problem data
		c.HTML(http.StatusOK, "edit-problem", editProblemPageData)
	})
	router.GET("/manage-problems", func(c *gin.Context) {
		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)
		problems := []frontend.ProblemSummary{
			{
				Id:     "1",
				Title:  "Two Sum",
				Status: "published",
			},
			{
				Id:     "2",
				Title:  "Fibonacci Sequence",
				Status: "draft",
			},
			{
				Id:     "3",
				Title:  "Binary Search",
				Status: "published",
			},
		}
		pageData := frontend.ManageProblemsPageData{
			Page:           "manage-problems",
			ClientUsername: clientUsername,
			IsClientAdmin:  clientUsername == "admin",
			Problems:       problems,
		}

		c.HTML(http.StatusOK, "manage-problems", pageData)
	})

	router.Run(":8081")
}

func getProblemByID(id int) (frontend.Problem, error) {
	// Simulate a database lookup
	mockProblems := map[int]frontend.Problem{
		1000: {
			Id:          1000,
			Title:       "Sorting Algorithm",
			ProblemName: "Sort the Array",
			Statement:   "Given an array of integers, sort the array in ascending order.",
			TimeLimit:   2,
			MemoryLimit: 256,
		},
		1001: {
			Id:          1001,
			Title:       "Binary Search",
			ProblemName: "Find the Element",
			Statement:   "Given a sorted array and a target value, find the index of the target using binary search.",
			TimeLimit:   2,
			MemoryLimit: 128,
		},
	}

	problem, exists := mockProblems[id]
	if !exists {
		problem = mockProblems[1000]
		// return nil, fmt.Errorf("problem not found")
	}
	return problem, nil
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
		Role:   "admin",
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
