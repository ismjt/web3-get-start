package controllers

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	var (
		err      error
		res      = gin.H{}
		url      string
		uploader Uploader
		file     multipart.File
		fh       *multipart.FileHeader
	)
	defer writeJSON(c, res)
	file, fh, err = c.Request.FormFile("file")
	if err != nil {
		res["message"] = err.Error()
		return
	}

	uploader = LocalUploader{
		BasePath: "static/upload",
		BaseURL:  "/static/upload/",
		FileType: []string{"image/*"},
	}

	url, err = uploader.upload(file, fh)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
	res["url"] = url
}
