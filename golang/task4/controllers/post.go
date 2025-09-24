package controllers

import (
	"go-blog/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func PostGet(c *gin.Context) {
	id, err := ParamUint(c, "id")
	if err != nil {
		HandleMessage(c, err.Error())
		return
	}
	post, err := models.GetPostById(id)
	if err != nil {
		Handle404(c)
		return
	}
	post.View++
	post.Comments, _ = models.ListCommentByPostID(id)
	userInterface, exists := c.Get(ContextUserKey)
	if exists {
		user, _ := userInterface.(*models.User)
		c.HTML(http.StatusOK, "post/display.html", gin.H{
			"post": post,
			"user": user,
		})
	} else {
		c.HTML(http.StatusOK, "post/display.html", gin.H{
			"post": post,
			"user": nil,
		})
	}
}

func PostNew(c *gin.Context) {
	c.HTML(http.StatusOK, "post/new.html", gin.H{
		"user": c.MustGet(ContextUserKey),
	})
}

func PostCreate(c *gin.Context) {
	title := c.PostForm("title")
	content := c.PostForm("body")

	userInterface := c.MustGet(ContextUserKey)
	user, _ := userInterface.(*models.User)

	post := &models.Post{
		Title:   title,
		Content: content,
		UserID:  user.ID,
		View:    0,
	}
	err := post.Insert()
	if err != nil {
		c.HTML(http.StatusOK, "post/new.html", gin.H{
			"post":    post,
			"message": err.Error(),
			"user":    user,
		})
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/admin/post")
}

func PostEdit(c *gin.Context) {
	userInterface := c.MustGet(ContextUserKey)
	user, _ := userInterface.(*models.User)
	id, err := ParamUint(c, "id")
	if err != nil {
		HandleMessage(c, err.Error())
		return
	}
	post, err := models.GetPostById(id)
	if err != nil {
		Handle404(c)
		return
	}

	// 首先验证是否具备编辑权限：只有文章的作者才能更新自己的文章
	if post.UserID == user.ID {
		c.HTML(http.StatusOK, "post/modify.html", gin.H{
			"post": post,
			"user": user,
		})
	} else {
		c.HTML(http.StatusOK, "errors/error.html", gin.H{
			"user":    user,
			"message": "《" + post.Title + "》只有文章的作者才能更新自己的文章",
		})
	}
}

func PostUpdate(c *gin.Context) {
	userInterface := c.MustGet(ContextUserKey)
	user, _ := userInterface.(*models.User)

	title := c.PostForm("title")
	content := c.PostForm("content")

	id, err := ParamUint(c, "id")
	if err != nil {
		HandleMessage(c, err.Error())
		return
	}

	exist, err := models.GetPostById(id)
	if err != nil {
		Handle404(c)
	}
	if exist.UserID == user.ID {
		post := &models.Post{
			Title:   title,
			Content: content,
		}
		post.ID = id
		err = post.Update()
		if err != nil {
			c.HTML(http.StatusOK, "post/modify.html", gin.H{
				"post":    post,
				"message": err.Error(),
				"user":    user,
			})
		}
	} else {
		c.HTML(http.StatusOK, "errors/error.html", gin.H{
			"user":    user,
			"message": "《" + exist.Title + "》只有文章的作者才能更新自己的文章",
		})
	}

	c.Redirect(http.StatusMovedPermanently, "/admin/post/"+strconv.Itoa(int(id))+"/edit")
}

func PostPublish(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		post *models.Post
	)
	defer writeJSON(c, res)
	id, err := ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	post, err = models.GetPostById(id)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	err = post.Update()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func PostDelete(c *gin.Context) {
	userInterface := c.MustGet(ContextUserKey)
	user, _ := userInterface.(*models.User)

	var (
		err error
		res = gin.H{}
	)
	defer writeJSON(c, res)
	id, err := ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}

	exist, err := models.GetPostById(id)
	if err != nil {
		Handle404(c)
	}
	if exist.UserID == user.ID {
		post := &models.Post{}
		post.ID = id
		err = post.LogicDelete()
		if err != nil {
			res["message"] = err.Error()
			return
		}
		res["succeed"] = true
	} else {
		c.HTML(http.StatusOK, "errors/error.html", gin.H{
			"user":    user,
			"message": "《" + exist.Title + "》只有文章的作者才能删除自己的文章",
		})
	}
}

func PostIndex(c *gin.Context) {
	posts, _ := models.ListAllPost()
	comments, _ := models.ListAllComment()
	userInterface := c.MustGet(ContextUserKey)
	user, _ := userInterface.(*models.User)
	c.HTML(http.StatusOK, "admin/post.html", gin.H{
		"posts":    posts,
		"Active":   "posts",
		"user":     user,
		"comments": comments,
	})
}
