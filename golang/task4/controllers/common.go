package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	SessionJwtKey  = "Token"
	SessionKey     = "UserID"      // session key
	ContextUserKey = "User"        // context user key
	SessionCaptcha = "GIN_CAPTCHA" // captcha session key
)

func Handle404(c *gin.Context) {
	HandleMessage(c, "Sorry,I lost myself!")
}

func HandleMessage(c *gin.Context, message string) {
	// user, _ := c.Get(ContextUserKey)
	c.HTML(http.StatusNotFound, "errors/error.html", gin.H{
		"message": message,
		"user":    "Test",
	})
}

func QueryUint(c *gin.Context, key string) (uint, error) {
	return parseUint(c.Query(key))
}

func ParamUint(c *gin.Context, key string) (uint, error) {
	return parseUint(c.Param(key))
}

func PostFormUint(c *gin.Context, key string) (uint, error) {
	return parseUint(c.PostForm(key))
}

func parseUint(value string) (uint, error) {
	val, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}

func writeJSON(ctx *gin.Context, h gin.H) {
	if _, ok := h["succeed"]; !ok {
		h["succeed"] = false
	}
	ctx.JSON(http.StatusOK, h)
}
