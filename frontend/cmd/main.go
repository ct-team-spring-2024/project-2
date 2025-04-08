package main

import (
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"net/http"
)


func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})

	logrus.Debug("This is a debug message")

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	// router.GET("/", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "index.html", nil)
	// })
	cnt := 0
	router.GET("/", func(c *gin.Context) {
		cnt = cnt + 1
		c.HTML(http.StatusOK, "index.html", gin.H{
			"counter": cnt,
		})
	})

	router.Run(":8080")
}
