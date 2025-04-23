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
		type formDataType struct {
			Username string `form:"username"`
			Password string `form:"password"`
		}
		type apiRequestDataType struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		type apiResponseDataType struct {
			Token string `json:"token"`
		}

		// request
		var formData formDataType
		if err := c.ShouldBind(&formData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		var apiRequestData apiRequestDataType
		apiRequestData = apiRequestDataType{
			Username: formData.Username,
			Password: formData.Password,
		}
		payloadBytes, err := json.Marshal(apiRequestData)
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

		var result apiResponseDataType
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
		session.Values["username"] = formData.Username
		session.Values["jwt"] = result.Token
		if err := session.Save(c.Request, c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}

		c.Redirect(http.StatusFound, fmt.Sprintf("/profile/%s", formData.Username))
	})

	router.GET("/profile/:username", func(c *gin.Context) {
		type apiResponseDataType struct {
			Profile struct {
				UserId   int    `json:"userId"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Password string `json:"password"`
				Role     string `json:"role"`
			} `json:"profile"`
			SubmissionStats struct {
				Total          int     `json:"total"`
				SuccessCount   int     `json:"successCount"`
				SuccessPercent float64 `json:"successPercent"`
				FailCount      int     `json:"failCount"`
				FailPercent    float64 `json:"failPercent"`
				ErrorCount     int     `json:"errorCount"`
				ErrorPercent   float64 `json:"errorPercent"`
			} `json:"submissionStats"`
		}

		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)
		token := session.Values["jwt"].(string)
		username := c.Param("username")
		logrus.Infof("clientUsername => %s", clientUsername)
		logrus.Infof("token => %s", token)
		logrus.Infof("username => %s", username)

		profileUrl := fmt.Sprintf("%s/profile/%s", backendUrl, username)
		req, err := http.NewRequest("POST", profileUrl, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
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

		var result apiResponseDataType
		if err := json.Unmarshal(body, &result); err != nil {
			logrus.Infof("Resultttt => %+v", result)
			logrus.Infof("Error => %+v", err)
			logStringError(body)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid response from backend"})
			return
		}

		logrus.Infof("result => %+v", result)
		pageData := frontend.ProfilePageData{
			Page:             "profile",
			ClientUsername:   clientUsername,
			IsClientAdmin:    clientUsername == "admin",
			IsUserAdmin:      username == "admin",
			Username:         username,
			Submissions:      make([]frontend.Submission, 0),
			Email:            result.Profile.Email,
			MemberSince:      "January 2023",
			TotalSubmissions: result.SubmissionStats.Total,
			SolvedProblems:   result.SubmissionStats.SuccessCount,
			SolveRate:        100,
		}
		c.HTML(http.StatusOK, "profile", pageData)
	})

	router.GET("/problems", func(c *gin.Context) {
		type apiRequestDataType struct {
			PageNo   int `json:"pageno"`
			PageSize int `json:"pagesize"`
		}
		type apiResponseDataType struct {
			ProblemId    int       `json:"problemId"`
			OwnerId      int       `json:"ownerId"`
			Title        string    `json:"title"`
			Statement    string    `json:"statement"`
			TimeLimit    int       `json:"timeLimit"`
			MemoryLimit  int       `json:"memoryLimit"`
			Input        string    `json:"input"`
			Output       string    `json:"output"`
			Status       string    `json:"status"`
			Feedback     string    `json:"feedback"`
			PublishDate  string    `json:"publishDate"`
		}

		session, _ := store.Get(c.Request, "session-name")
		token := session.Values["jwt"].(string)
		clientUsername := session.Values["username"].(string)
		pageNo, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))


		problemsUrl := fmt.Sprintf("%s/problems", backendUrl)
		apiRequestData := apiRequestDataType{
			PageNo: pageNo,
			PageSize: pageSize,
		}
		apiRequestBytes, err := json.Marshal(apiRequestData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
			return
		}

		req, err := http.NewRequest("GET", problemsUrl, bytes.NewReader(apiRequestBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
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

		var result []apiResponseDataType
		if err := json.Unmarshal(body, &result); err != nil {
			logrus.Infof("Result => %+v", result)
			logrus.Infof("Error => %+v", err)
			logStringError(body)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid response from backend"})
			return
		}

		problems := make([]frontend.ProblemPageProblemSummary, 0, 0)
		for _, p := range result {
			problems = append(problems, frontend.ProblemPageProblemSummary{
				Id: p.ProblemId,
				Title: p.Title,
			})
		}
		pageData := frontend.ProblemsPageData{
			Page:           "problems",
			ClientUsername: clientUsername,
			IsClientAdmin:  clientUsername == "admin",
			Problems:       problems,
			CurrentPage:    pageNo,
			Limit:          20,
			HasNextPage:    pageNo < 100,
			TotalPages:     1000,
		}
		c.HTML(http.StatusOK, "problems", pageData)
	})

	router.GET("/problem/:problemid", func(c *gin.Context) {
		type apiResponseDataType struct {
			ProblemId    int       `json:"problemId"`
			OwnerId      int       `json:"ownerId"`
			Title        string    `json:"title"`
			Statement    string    `json:"statement"`
			TimeLimit    int       `json:"timeLimit"`
			MemoryLimit  int       `json:"memoryLimit"`
			Input        string    `json:"input"`
			Output       string    `json:"output"`
			Status       string    `json:"status"`
			Feedback     string    `json:"feedback"`
			PublishDate  string    `json:"publishDate"`
		}

		session, _ := store.Get(c.Request, "session-name")
		token := session.Values["jwt"].(string)
		clientUsername := session.Values["username"].(string)
		problemIdStr := c.Param("problemid")
		problemId, err := strconv.Atoi(problemIdStr)
		if err != nil || problemId <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
			return
		}

		problemsUrl := fmt.Sprintf("%s/problems/%d", backendUrl, problemId)
		req, err := http.NewRequest("GET", problemsUrl, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
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

		var result apiResponseDataType
		if err := json.Unmarshal(body, &result); err != nil {
			logrus.Infof("Result => %+v", result)
			logrus.Infof("Error => %+v", err)
			logStringError(body)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid response from backend"})
			return
		}

		pageData := frontend.SingleProblemPageData{
			Page:               "problem",
			ClientUsername:     clientUsername,
			IsClientAdmin:      clientUsername == "admin",
			ProblemId:          problemId,
			ProblemStatement:   result.Statement,
			ProblemTitle:       result.Title,
			ProblemTimeLimit:   result.TimeLimit,
			ProblemMemoryLimit: result.MemoryLimit,
		}
		c.HTML(http.StatusOK, "problem", pageData)
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
