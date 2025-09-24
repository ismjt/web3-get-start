package controllers

import (
	"go-blog/helpers"
	"go-blog/models"
	"net/http"
	"time"

	"github.com/cihub/seelog"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type GithubUserInfo struct {
	Blog      string      `json:"blog"`
	CreatedAt string      `json:"created_at"`
	Email     interface{} `json:"email"`
	HTMLURL   string      `json:"html_url"`
	ID        int         `json:"id"`
	Login     string      `json:"login"`
	Name      interface{} `json:"name"`
	UpdatedAt string      `json:"updated_at"`
	URL       string      `json:"url"`
}

func SigninGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/signin.html", gin.H{
		"cfg": "",
	})
}

func SignupGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/signup.html", gin.H{
		"cfg": "",
	})
}

func LogoutGet(c *gin.Context) {
	s := sessions.Default(c)
	s.Delete(SessionJwtKey)
	s.Delete(ContextUserKey)
	s.Delete(SessionKey)
	s.Clear()
	err := s.Save()
	if err != nil {
		seelog.Error(err)
	}
	c.Redirect(http.StatusSeeOther, "/signin")
}

// 用户注册
func SignupPost(c *gin.Context) {
	var (
		err   error
		param *models.RegisterRequest
	)
	// 从请求体中解析 JSON 到 param
	if err = c.ShouldBind(&param); err != nil {
		seelog.Infof(err.Error())
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message": "param invalid",
		})
		return
	}

	if len(param.Password) == 0 || len(param.Username) == 0 {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message": "email or password cannot be null",
		})
		return
	}

	// 校验密码强度
	err = helpers.ValidatePasswordStrength(param.Password)
	if err != nil {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message": "password is simple",
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(param.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 随机设置一个邮箱号
	email, _ := helpers.RandomEmail()
	param.Email = email

	user := &models.User{
		Email:    param.Email,
		Username: param.Username,
		Password: string(hashedPassword),
	}
	err = user.Insert()
	if err != nil {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message": "register info already exists",
			"cfg":     "",
		})
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/signin")
}

// 用户登陆接口
func SigninPost(c *gin.Context) {
	var (
		err         error
		param       *models.LoginRequest
		user        *models.User
		tokenString string
		errTip      = "Invalid username or password"
	)

	if err := c.ShouldBind(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err = models.GetUserByUsername(param.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errTip})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(param.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errTip})
		return
	}

	// 生成 JWT
	exp := time.Now().Add(time.Hour * 24)
	tokenString, err = helpers.GenerateToken(*user, exp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	resp := models.DataResponse[models.LoginData]{
		BaseResponse: models.BaseResponse{Code: 200, Msg: "success"},
		Payload: models.LoginData{
			Token:     tokenString,
			ExpiresAt: exp.Unix(),
			UserID:    user.ID,
			Username:  user.Username,
		},
	}

	session := sessions.Default(c)
	session.Set("Token", tokenString)
	session.Set("ExpiresAt", exp.Unix())
	session.Set("UserID", user.ID)
	session.Set("Username", user.Username)
	err = session.Save()
	if err != nil {
		seelog.Error("SigninPost session.Save Error: " + err.Error())
	}

	c.JSON(http.StatusOK, resp)
}
