package controllers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 获取文件大小的接口
type Size interface {
	Size() int64
}

// 获取文件信息的接口
type Stat interface {
	Stat() (os.FileInfo, error)
}

type LocalUploader struct {
	BasePath string   // 文件存储根路径，例如 "./static/upload"
	BaseURL  string   // 返回的文件访问 URL 前缀，例如 "/static/upload/"
	FileType []string // 允许上传的文件类型 ["image/*", "application/pdf"]
}

// MIME 类型匹配
func mimeMatch(mime string, allowed []string) bool {
	for _, t := range allowed {
		if strings.HasSuffix(t, "/*") {
			prefix := strings.TrimSuffix(t, "*")
			if strings.HasPrefix(mime, prefix) {
				return true
			}
		} else if mime == t {
			return true
		}
	}
	return false
}

func (u LocalUploader) upload(file multipart.File, _ *multipart.FileHeader) (url string, err error) {
	var size int64

	// 尝试通过 Stat 获取文件大小
	if statInterface, ok := file.(Stat); ok {
		fileInfo, _ := statInterface.Stat()
		size = fileInfo.Size()
	}

	// 尝试通过 Size 获取文件大小
	if sizeInterface, ok := file.(Size); ok {
		size = sizeInterface.Size()
	}

	fmt.Println("文件大小:", size)

	// 读取文件前 512 字节判断 MIME 类型
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])

	// 校验 MIME 类型
	if len(u.FileType) > 0 && !mimeMatch(mimeType, u.FileType) {
		return "", fmt.Errorf("不允许上传该类型文件: %s", mimeType)
	}

	// 重置文件读取指针
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	} else {
		return "", fmt.Errorf("无法重置文件指针")
	}

	// 确保存储目录存在
	if err = os.MkdirAll(u.BasePath, os.ModePerm); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 为文件生成唯一名字（避免覆盖）
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), size, filepath.Ext("file")) // 这里可以换成文件原始扩展名
	filepath := filepath.Join(u.BasePath, filename)

	// 创建目标文件
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	// 拷贝内容到目标文件
	if _, err = io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	// 返回 URL
	url = u.BaseURL + filename
	return url, nil
}
