package controllers

import (
	"go-blog/models"
	"math"
	"net/http"
	"strconv"

	"github.com/cihub/seelog"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
)

func IndexGet(c *gin.Context) {
	var (
		pageIndex int
		pageSize  = 10
		total     int
		page      string
		err       error
		posts     []*models.Post
	)

	session := sessions.Default(c)
	userId := session.Get(SessionKey)
	var loginUser *models.User
	if userId != nil {
		loginUser, err = models.GetUser(userId)
		if err != nil {
			c.Set(ContextUserKey, loginUser)
		}
	}

	page = c.Query("page")
	pageIndex, _ = strconv.Atoi(page)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	posts, err = models.ListAllPost()
	if err != nil {
		seelog.Errorf("models.ListAllPost err: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	total, err = models.CountPost()
	if err != nil {
		seelog.Errorf("models.CountPost err: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	for _, post := range posts {
		post.Content = string(blackfriday.MarkdownCommon([]byte(post.Content)))
	}
	postArchives, _ := models.ListPostArchives()
	maxCommentPost, _ := models.ListMaxCommentPost()
	c.HTML(http.StatusOK, "index/index.html", gin.H{
		"posts":           posts,
		"archives":        postArchives,
		"user":            loginUser,
		"pageIndex":       pageIndex,
		"totalPage":       int(math.Ceil(float64(total) / float64(pageSize))),
		"path":            c.Request.URL.Path,
		"maxCommentPosts": maxCommentPost,
	})
}
