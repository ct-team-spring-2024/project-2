package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	ppath "path"
	"strconv"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"

	"oj/frontend"
)

var store = sessions.NewCookieStore([]byte("a-very-secret-key"))

var path = "C:/Users/Asus/Documents/GitHub/project-2"
var backendUrl string

//var path = "/home/mbroughani81/Documents/test/computer-technology-project-2"

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

	backendUrl = "http://localhost:8080"
	//Log in page
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})
	router.POST("/signup", func(c *gin.Context) {
		type formDataType struct {
			Username string `form:"username"`
			Password string `form:"password"`
			Email    string `form:"email"`
		}
		type apiRequestDataType struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		}
		type apiResponseDataType struct {
			Token string `json:"token"`
		}
		var formData formDataType
		if err := c.ShouldBind(&formData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}
		var apiRequestData apiRequestDataType
		apiRequestData = apiRequestDataType{
			Username: formData.Username,
			Password: formData.Password,
			Email:    formData.Email,
			Role:     "user",
		}
		payloadBytes, err := json.Marshal(apiRequestData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
			return
		}

		// send to backend
		signUpUrl := fmt.Sprintf("%s/register", backendUrl)
		resp, err := http.Post(signUpUrl, "application/json", bytes.NewReader(payloadBytes))
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
		user, err := getClientByUsername(clientUsername, c)
		if err != nil {
			logrus.Error("Error fetching user from backend")

		}
		profileUser, err := getClientByUsername(username, c)
		if err != nil {
			logrus.Error("Error fetching user from backend")

		}

		pageData := frontend.ProfilePageData{
			Page:             "profile",
			ClientUsername:   clientUsername,
			IsClientAdmin:    user.Role == "admin",
			IsUserAdmin:      profileUser.Role == "admin",
			Username:         username,
			Submissions:      make([]frontend.Submission, 0),
			Email:            result.Profile.Email,
			MemberSince:      "January 2023",
			TotalSubmissions: result.SubmissionStats.Total,
			SolvedProblems:   result.SubmissionStats.SuccessCount,
			SolveRate:        100,
		}
		logrus.Infof("pageData => %+v", pageData)
		logrus.Infof("pageData => %+v", user)
		logrus.Infof("pageData => %+v", profileUser)
		c.HTML(http.StatusOK, "profile", pageData)
	})

	router.GET("/problems", func(c *gin.Context) {
		type apiRequestDataType struct {
			PageNo   int `json:"pageno"`
			PageSize int `json:"pagesize"`
		}
		type apiResponseDataType struct {
			ProblemId   int    `json:"problemId"`
			OwnerId     int    `json:"ownerId"`
			Title       string `json:"title"`
			Statement   string `json:"statement"`
			TimeLimit   int    `json:"timeLimit"`
			MemoryLimit int    `json:"memoryLimit"`
			Input       string `json:"input"`
			Output      string `json:"output"`
			Status      string `json:"status"`
			Feedback    string `json:"feedback"`
			PublishDate string `json:"publishDate"`
		}

		session, _ := store.Get(c.Request, "session-name")
		token := session.Values["jwt"].(string)
		clientUsername := session.Values["username"].(string)
		pageNo, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		problemsUrl := fmt.Sprintf("%s/problems", backendUrl)
		apiRequestData := apiRequestDataType{
			PageNo:   pageNo,
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
		logrus.Infof("result => %+v", result)

		problems := make([]frontend.ProblemPageProblemSummary, 0)
		for _, p := range result {
			problems = append(problems, frontend.ProblemPageProblemSummary{
				Id:    p.ProblemId,
				Title: p.Title,
			})
		}
		user, err := getClientByUsername(clientUsername, c)
		if err != nil {
			logrus.Error("Error fetching the user from backend")
		}
		pageData := frontend.ProblemsPageData{
			Page:           "problems",
			ClientUsername: clientUsername,
			IsClientAdmin:  user.Role == "admin",
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
			ProblemId   int    `json:"problemId"`
			OwnerId     int    `json:"ownerId"`
			Title       string `json:"title"`
			Statement   string `json:"statement"`
			TimeLimit   int    `json:"timeLimit"`
			MemoryLimit int    `json:"memoryLimit"`
			Input       string `json:"input"`
			Output      string `json:"output"`
			Status      string `json:"status"`
			Feedback    string `json:"feedback"`
			PublishDate string `json:"publishDate"`
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
		user, err := getClientByUsername(clientUsername, c)
		if err != nil {
			logrus.Infof("Error fetching the user from the backend => %v", err)
		}
		logrus.Info("Client is", user)

		pageData := frontend.SingleProblemPageData{
			Page:               "problem",
			ClientUsername:     clientUsername,
			IsClientAdmin:      user.Role == "admin",
			ProblemId:          problemId,
			ProblemStatement:   result.Statement,
			ProblemTitle:       result.Title,
			ProblemTimeLimit:   result.TimeLimit,
			ProblemMemoryLimit: result.MemoryLimit,
		}
		c.HTML(http.StatusOK, "problem", pageData)
	})

	router.GET("/problem/:problemid/submit", func(c *gin.Context) {
		session, _ := store.Get(c.Request, "session-name")
		// token := session.Values["jwt"].(string)
		clientUsername := session.Values["username"].(string)
		problemIdStr := c.Param("problemid")
		problemId, _ := strconv.Atoi(problemIdStr)

		pageData := frontend.SubmitPageData{
			Page:           "submit",
			ClientUsername: clientUsername,
			IsClientAdmin:  clientUsername == "admin",
			ProblemId:      problemId,
		}

		c.HTML(http.StatusOK, "submit", pageData)
	})

	router.POST("/problem/:problemid/submit", func(c *gin.Context) {
		type formDataType struct {
			Code string `form:"code"`
		}
		type apiRequestDataType struct {
			Code      string `json:"code"`
			ProblemId int    `json:"problemId"`
			Language  string `json:"language"`
		}

		session, _ := store.Get(c.Request, "session-name")
		token := session.Values["jwt"].(string)
		clientUsername := session.Values["username"].(string)
		problemIdStr := c.Param("problemid")
		problemId, _ := strconv.Atoi(problemIdStr)
		submitUrl := fmt.Sprintf("%s/submit", backendUrl)

		var formData formDataType
		if err := c.ShouldBind(&formData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}
		var apiRequestData apiRequestDataType
		apiRequestData = apiRequestDataType{
			Code:      formData.Code,
			ProblemId: problemId,
			Language:  "go",
		}
		jsonData, err := json.Marshal(apiRequestData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request data"})
			return
		}
		req, _ := http.NewRequest("POST", submitUrl, bytes.NewBuffer(jsonData))
		logrus.Infof("SSSS => %s", jsonData)
		logrus.Infof("SSSS => %s", submitUrl)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact backend"})
			return
		}
		defer resp.Body.Close()

		c.Redirect(http.StatusFound, fmt.Sprintf("/profile/%s", clientUsername))
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
		type apiRequestDataType struct {
			ProblemId   int           `json:"problemId"`
			OwnerId     int           `json:"ownerId"`
			Title       string        `json:"title"`
			Statement   string        `json:"statement"`
			TimeLimit   int           `json:"timeLimit"`   // in seconds
			MemoryLimit int           `json:"memoryLimit"` // in MB
			Inputs      []string      `json:"inputs"`
			Outputs     []string      `json:"outputs"`
			Status      string        `json:"status"`
		}
		session, _ := store.Get(c.Request, "session-name")
		token := session.Values["jwt"].(string)
		clientUsername := session.Values["username"].(string)
		problemIdStr := c.Param("problemid")
		problemId, err := strconv.Atoi(problemIdStr)

		// Parse the multipart form (with a max memory of 32MB for file uploads)
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse form data"})
			return
		}
		problemTitle := c.PostForm("problemTitle")
		timeLimitStr := c.PostForm("timeLimit")
		timeLimit, _ := strconv.Atoi(timeLimitStr)
		memoryLimitStr := c.PostForm("memoryLimit")
		memoryLimit, _ := strconv.Atoi(memoryLimitStr)
		problemDescription := c.PostForm("problemDescription")
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
		fmt.Printf("Problem Title: %s\n", problemTitle)
		fmt.Printf("Time Limit: %d ms\n", timeLimit)
		fmt.Printf("Memory Limit: %d MB\n", memoryLimit)
		fmt.Printf("Problem Description: %s\n", problemDescription)
		fmt.Printf("Test Case File Name: %s\n", fileHeader.Filename)
		fmt.Printf("Test Case Content: %s\n", testCaseContent)

		// Parse Test Case Content
		var testCases map[string]struct {
			Inputs string `json:"inputs"`
			Output string `json:"output"`
		}
		err = json.Unmarshal([]byte(testCaseContent), &testCases)
		if err != nil {
			logrus.Errorf("Error parsing JSON: %v\n", err)
			return
		}
		var inputs []string
		var outputs []string
		for _, testCase := range testCases {
			inputs = append(inputs, testCase.Inputs)
			outputs = append(outputs, testCase.Output)
		}
		//
		user, err := getClientByUsername(clientUsername, c)
		if err != nil {
			logrus.Error("Error fetching user from backend")

		}
		var apiRequestData apiRequestDataType
		apiRequestData = apiRequestDataType{
			ProblemId: problemId,
			OwnerId:   user.ID,
			Title:     problemTitle,
			Statement: problemDescription,
			TimeLimit: timeLimit,
			MemoryLimit: memoryLimit,
			Inputs:      inputs,
			Outputs:     outputs,
			Status:      "Draft",
		}
		jsonData, err := json.Marshal(apiRequestData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request data"})
			return
		}
		addProblemUrl := fmt.Sprintf("%s/problems", backendUrl)
		req, _ := http.NewRequest("POST", addProblemUrl, bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact backend"})
			return
		}
		defer resp.Body.Close()

		c.Redirect(http.StatusFound, fmt.Sprintf("/profile/%s", clientUsername))
	})
	router.GET("/my-problems", func(c *gin.Context) {
		type apiResponseDataType struct {
			ProblemId   int    `json:"problemId"`
			OwnerId     int    `json:"ownerId"`
			Title       string `json:"title"`
			TimeLimit   int    `json:"timeLimit"`
			Status      string `json:"status"`
			MemoryLimit int    `json:"memoryLimit"`
			PublishDate string `json:"publishDate"`
		}
		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)

		token := session.Values["jwt"].(string)
		myProblemsurl := fmt.Sprintf("%v/problems/mine", backendUrl)

		req, err := http.NewRequest("GET", myProblemsurl, nil)

		if err != nil {
			logrus.Error("Error contacing the backend")
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
		problems := make([]frontend.ProblemSummary, 0)

		n := len(result)
		for i := 0; i < n; i++ {
			p := frontend.ProblemSummary{
				Id:     strconv.Itoa(result[i].ProblemId),
				Title:  result[i].Title,
				Status: result[i].Status,
			}
			problems = append(problems, p)
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
		logrus.Info("id is ", id)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
			return
		}

		// Fetch the problem details
		problem, err := getProblemByID(id, c)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
			logrus.Error(err)
			return
		}
		user, err := getClientByUsername(clientUsername, c)
		if err != nil {
			logrus.Error("Error fetching client from backend")
		}

		editProblemPageData := frontend.EditProblemPageData{
			Page:           "edit-problem",
			ClientUsername: clientUsername,
			IsClientAdmin:  user.Role == "admin",
			Problem:        problem,
		}
		logrus.Infof("Problem => %+v", editProblemPageData.Problem)

		// Render the problem.html template with the problem data
		c.HTML(http.StatusOK, "edit-problem", editProblemPageData)
	})
	router.POST("/edit/:id", func(c *gin.Context) {
		// session, _ := store.Get(c.Request, "session-name")
		// clientUsername := session.Values["username"].(string)

		// // Extract the problem ID from the URL
		// idStr := c.Param("id")
		// id, err := strconv.Atoi(idStr)
		// logrus.Info("id is ", id)
		// if err != nil || id <= 0 {
		//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		//	return
		// }

		// Fetch the problem details
		// problem, err := getProblemByID(id, c)
		// if err != nil {
		//	c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
		//	logrus.Error(err)
		//	return
		// }
		// user, err := getClientByUsername(clientUsername, c)
		// if err != nil {
		//	logrus.Error("Error fetching client from backend")
		// }

	})
	router.GET("/manage-problems", func(c *gin.Context) {
		type apiResponseDataType struct {
			ProblemId   int    `json:"problemId"`
			OwnerId     int    `json:"ownerId"`
			Title       string `json:"title"`
			TimeLimit   int    `json:"timeLimit"`
			Status      string `json:"status"`
			MemoryLimit int    `json:"memoryLimit"`
			PublishDate string `json:"publishDate"`
		}
		type apiRequestDataType struct {
		}

		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)
		token := session.Values["jwt"].(string)
		user, err := getClientByUsername(clientUsername, c)
		if err != nil {
			logrus.Error("Error fetching user from backend")

		}

		allProblemsUrl := fmt.Sprintf("%v/admin/problems", backendUrl)
		req, err := http.NewRequest("GET", allProblemsUrl, nil)
		if err != nil {
			logrus.Error("Error Creating the request")
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
		logrus.Info(string(body))

		var result []apiResponseDataType
		if err := json.Unmarshal(body, &result); err != nil {
			logrus.Infof("Result => %+v", result)
			logrus.Infof("Error => %+v", err)
			logStringError(body)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid response from backend"})
			return
		}
		logrus.Info(result)
		problems := make([]frontend.ProblemSummary, 0)

		n := len(result)
		for i := 0; i < n; i++ {
			p := frontend.ProblemSummary{
				Id:     strconv.Itoa(result[i].ProblemId),
				Title:  result[i].Title,
				Status: result[i].Status,
			}
			problems = append(problems, p)
		}
		logrus.Infof("OK %s %+v", clientUsername, user)
		pageData := frontend.ManageProblemsPageData{
			Page:           "manage-problems",
			ClientUsername: clientUsername,
			IsClientAdmin:  user.Role == "admin",
			Problems:       problems,
		}

		c.HTML(http.StatusOK, "manage-problems", pageData)
	})
	router.POST("/manage-problems/update", func(c *gin.Context) {
		type formDataType struct {
			ID     string `form:"id"`
			Status string `form:"status"`
		}
		type apiRequestDataType struct {
			ProblemId int    `json:"problemId"`
			NewStatus string `json:"newStatus"`
			Feedback  string `json:"feedback"`
		}

		session, _ := store.Get(c.Request, "session-name")
		token := session.Values["jwt"].(string)
		var formData formDataType
		if err := c.ShouldBind(&formData); err != nil {
			logrus.Errorf("Error binding form data: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
			return
		}

		updateProblemStatusUrl := fmt.Sprintf("%v/admin/problems/status", backendUrl)
		problemId, _ := strconv.Atoi(formData.ID)
		var apiRequestData apiRequestDataType
		apiRequestData = apiRequestDataType{
			ProblemId: problemId,
			Feedback: "",
			NewStatus: formData.Status,
		}
		apiRequestBytes, err := json.Marshal(apiRequestData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
			return
		}
		req, _ := http.NewRequest("POST", updateProblemStatusUrl, bytes.NewReader(apiRequestBytes))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact backend"})
			return
		}
		defer resp.Body.Close()

		c.Redirect(http.StatusFound, "/manage-problems")
	})
	router.GET("/submissions", func(c *gin.Context) {
		type apiResponseDataType struct {
			ID               int                    `json:"id"`
			UserID           int                    `json:"userId"`
			ProblemID        int                    `json:"problemId"`
			Code             string                 `json:"code"`
			TestsStatus      map[string]interface{} `json:"testsstatus"` // Empty object in JSON
			SubmissionStatus string                 `json:"submissionstatus"`
		}

		session, _ := store.Get(c.Request, "session-name")
		clientUsername := session.Values["username"].(string)
		token := session.Values["jwt"].(string)
		username := c.Param("username")
		logrus.Infof("clientUsername => %s", clientUsername)
		logrus.Infof("token => %s", token)
		logrus.Infof("username => %s", username)

		submissionsUrl := fmt.Sprintf("%s/submissions", backendUrl)
		req, err := http.NewRequest("GET", submissionsUrl, nil)
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
			logrus.Infof("Resultttt => %+v", result)
			logrus.Infof("Error => %+v", err)
			logStringError(body)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid response from backend"})
			return
		}

		logrus.Infof("result => %+v", result)
		submissions := make([]frontend.SubmissionsPageEntryData, 0, 0)

		var parseTestsStatus = func(m map[string]interface{}) frontend.TestsStatus {
			result := make(frontend.TestsStatus)
			for key, value := range m {
				if status, ok := value.(string); ok {
					result[key] = struct {
						Status string
					}{
						Status: status,
					}
				} else {
					result[key] = struct {
						Status string
					}{
						Status: "Unknown",
					}
				}
			}
			return result
		}
		var calScore = func(t frontend.TestsStatus) int {
			totalTests := len(t)
			okCount := 0
			for _, testResult := range t {
				if testResult.Status == "OK" {
					okCount++
				}
			}
			if totalTests == 0 {
				return 0
			}
			percentage := (okCount * 100) / totalTests

			return percentage
		}

		for _, d := range result {
			testsStatus := parseTestsStatus(d.TestsStatus)
			score := calScore(testsStatus)
			submissions = append(submissions, frontend.SubmissionsPageEntryData{
				Id:               d.ID,
				ProblemId:        d.ProblemID,
				SubmissionStatus: d.SubmissionStatus,
				Score:            score,
				TestsStatus:      testsStatus,
			})
		}
		pageData := frontend.SubmissionsPageData{
			Page:           "submissions",
			ClientUsername: clientUsername,
			IsClientAdmin:  clientUsername == "admin",
			Submissions:    submissions,
		}
		logrus.Infof("pd => \n %+v", pageData)
		c.HTML(http.StatusOK, "submissions", pageData)
	})

	router.Run(":8081")
}

func getProblemByID(id int, c *gin.Context) (frontend.Problem, error) {

	type apiResponseDataType struct {
		ProblemId   int    `json:"problemId"`
		OwnerId     int    `json:"ownerId"`
		Title       string `json:"title"`
		Statement   string `json:"statement"`
		TimeLimit   int    `json:"timeLimit"`
		MemoryLimit int    `json:"memoryLimit"`
		Input       string `json:"input"`
		Output      string `json:"output"`
		Status      string `json:"status"`
		Feedback    string `json:"feedback"`
		PublishDate string `json:"publishDate"`
	}

	session, _ := store.Get(c.Request, "session-name")

	token := session.Values["jwt"].(string)
	problemUrl := fmt.Sprintf("%v/problems/%v", backendUrl, id)

	req, err := http.NewRequest("GET", problemUrl, nil)

	if err != nil {
		logrus.Error("Error Creating the request")
		return frontend.Problem{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact backend"})
		return frontend.Problem{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read backend response"})
		return frontend.Problem{}, err
	}
	logrus.Info(string(body))

	var result apiResponseDataType
	if err := json.Unmarshal(body, &result); err != nil {
		logrus.Infof("Result => %+v", result)
		logrus.Infof("Error => %+v", err)
		logStringError(body)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid response from backend"})
		return frontend.Problem{}, err
	}
	problem := frontend.Problem{
		Id:          result.ProblemId,
		Statement:   result.Statement,
		Title:       result.Title,
		TimeLimit:   result.TimeLimit,
		MemoryLimit: result.MemoryLimit,
		Status:      result.Status,
	}

	return problem, nil

}
func getProblemByIdMock(id int) (frontend.Problem, error) {
	//	Simulate a database lookup
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
		return frontend.Problem{}, fmt.Errorf("problem not found")

	}
	return problem, nil
}
func getClientByUsername(username string, c *gin.Context) (frontend.User, error) {
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
	logrus.Infof("clientUsername => %s", clientUsername)
	logrus.Infof("token => %s", token)
	logrus.Infof("username => %s", username)

	profileUrl := fmt.Sprintf("%s/profile/%s", backendUrl, username)
	req, err := http.NewRequest("GET", profileUrl, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return frontend.User{}, nil
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact backend"})
		return frontend.User{}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read backend response"})
		return frontend.User{}, nil
	}
	var user apiResponseDataType
	if err := json.Unmarshal(body, &user); err != nil {
		logrus.Infof("Resultttt => %+v", user)
		logrus.Infof("Error => %+v", err)
		logStringError(body)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid response from backend"})
		return frontend.User{}, nil
	}
	UserC := frontend.User{
		ID:       user.Profile.UserId,
		Username: user.Profile.Username,
		Email:    user.Profile.Email,
		Role:     user.Profile.Role,
	}

	return UserC, nil

}
