package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthGet(c *gin.Context) {
	authurl := "/signin"
	c.Redirect(http.StatusFound, authurl)
}
