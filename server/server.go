package server

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// hello world, the web server
func HelloServer(c *gin.Context) {
	var buffer bytes.Buffer
	buffer.WriteString("Hello world!\n")

	c.String(http.StatusOK, buffer.String())
}

func Run() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/", HelloServer)

	router.Run(":" + port)
}
