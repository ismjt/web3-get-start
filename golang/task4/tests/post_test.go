package tests

import (
	"go-blog/models"
	"testing"

	"gorm.io/gorm"
)

func TestInsertPost(t *testing.T) {
	db := setupTestDB()

	// 重置数据表中的记录
	db.Exec("DELETE FROM posts")                              // 删除所有记录
	db.Exec("DELETE FROM sqlite_sequence WHERE name='posts'") // 重置自增 ID

	var u models.User
	db.First(&u, 1)
	post := models.Post{
		Model:   gorm.Model{ID: 1},
		Title:   "Go Concurrency Patterns",
		Content: "In this article, we explore Go concurrency patterns using goroutines and channels...",
		View:    0,
		User:    u,
		Comments: []models.Comment{
			{Content: "Great article!"},
			{Content: "Very helpful, thanks!"},
		},
		CommentTotal: 2,
	}

	result := db.Create(&post)
	if result.Error != nil {
		t.Fatalf("Failed to insert user: %v", result.Error)
	}

	if post.ID == 0 {
		t.Fatal("Expected user.ID to be set after insert")
	}

	// 查询验证
	var p models.Post
	db.First(&p, post.ID)
	if p.ID != 1 {
		t.Errorf("Expected title 'Go Concurrency Patterns', got '%s'", p.Title)
	}
}
