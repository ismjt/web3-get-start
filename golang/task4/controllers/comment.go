package controllers

import (
	"go-blog/models"

	"github.com/cihub/seelog"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

// 根据文章ID查询评论列表和对应的用户数据
func CommentPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
		// post *models.Post
	)
	defer writeJSON(c, res)

	// 检查验证码
	verifyCode := c.PostForm("verifyCode")
	captchaId := c.PostForm("captchaId")
	flag := captcha.VerifyString(captchaId, verifyCode)
	if !flag {
		res["message"] = "verify code incorrect"
		return
	}

	// 读取用户信息
	userInterface, exists := c.Get(ContextUserKey)
	if !exists {
		res["message"] = "please login first"
		return
	}
	user, _ := userInterface.(*models.User)

	content := c.PostForm("content")
	if len(content) == 0 {
		res["message"] = "content cannot be empty."
		return
	}
	pid, err := PostFormUint(c, "postId")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	_, err = models.GetPostById(pid)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := &models.Comment{
		PostID:  pid,
		Content: content,
		UserID:  user.ID,
	}
	err = comment.Insert()
	if err != nil {
		seelog.Errorf("comment insert err: %v", err)
		res["message"] = err.Error()
		return
	}

	seelog.Infof("User[ID:%v] Save Post[ID:%v] comment: %s ", user.ID, pid, content)

	res["succeed"] = true
}

// TODO 根据ID删除评论
func CommentDelete(c *gin.Context) {
	var (
		err error
		res = gin.H{}
		cid uint
	)
	defer writeJSON(c, res)

	userInterface, exists := c.Get(ContextUserKey)
	if !exists {
		res["message"] = "please login first"
		return
	}
	user, _ := userInterface.(*models.User)

	// 验证是否为自己的评论

	cid, err = ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := &models.Comment{
		UserID: user.ID,
	}
	comment.ID = cid
	err = comment.Delete()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

// TODO
func CommentRead(c *gin.Context) {
	var (
		id  uint
		err error
		res = gin.H{}
	)
	defer writeJSON(c, res)
	id, err = ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := new(models.Comment)
	comment.ID = id
	err = comment.Update()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
