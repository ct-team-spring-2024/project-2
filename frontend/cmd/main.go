package main

import (
	"html/template"
	"net/http"
	"oj/frontend"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
	cnt := 0
	router.GET("/", func(c *gin.Context) {
		cnt = cnt + 1
		c.HTML(http.StatusOK, "index.html", gin.H{
			"counter": cnt,
		})
	})

	//Problems page

	router.GET("/problems.html", func(c *gin.Context) {
		pageNo := c.DefaultQuery("page", "1")
		limitNo := c.DefaultQuery("limit", "20")

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
		c.HTML(http.StatusOK, "problems.html", pageData)

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
