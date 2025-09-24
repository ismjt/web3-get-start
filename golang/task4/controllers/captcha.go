package controllers

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

func CaptchaGet(c *gin.Context) {
	captchaId := captcha.NewLen(4)
	// 可把 captchaId 返回给前端，前端用作请求参数
	c.JSON(http.StatusOK, gin.H{
		"captchaId": captchaId,
		"imageUrl":  "/captcha/image/" + captchaId, // 前端可访问
	})
}

// 直接输出图片
func CaptchaImage(c *gin.Context) {
	captchaId := c.Param("captchaId")
	captcha.WriteImage(c.Writer, captchaId, 100, 40)
}
