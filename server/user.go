package server

import (
	//"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heroku/stocksignals/model"
	"github.com/heroku/stocksignals/store"
)

// GetSignals retrieves the signals from the user
func GetUsers(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer store.Disconnect()

	users, err := store.GetUsers()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, users)
}

// RegisterUser registers the given user on the database
func RegisterUser(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var user model.User
	if err := c.BindJSON(&user); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := store.RegisterUser(user); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "user is registered"})
}

// GetUserByEmail gets the given user info from the database via email
func GetUserByEmail(c *gin.Context) {
	if err := store.Connect(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	email := c.Param("email")
	user, err := store.GetUser(email)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}
