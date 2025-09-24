package tests

import (
	"go-blog/models"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestInsertUser(t *testing.T) {
	db := setupTestDB()

	// 重置数据表中的记录
	db.Exec("DELETE FROM users")                              // 删除所有记录
	db.Exec("DELETE FROM sqlite_sequence WHERE name='users'") // 重置自增 ID

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := models.User{
		Model:    gorm.Model{ID: 1},
		Username: "Alice",
		Password: string(hashedPassword),
		Email:    "alice@example.com",
	}

	result := db.Create(&user)
	if result.Error != nil {
		t.Fatalf("Failed to insert user: %v", result.Error)
	}

	if user.ID == 0 {
		t.Fatal("Expected user.ID to be set after insert")
	}

	// 查询验证
	var u models.User
	db.First(&u, user.ID)
	if u.ID != 1 {
		t.Errorf("Expected username 'Alice', got '%s'", u.Username)
	}
}
