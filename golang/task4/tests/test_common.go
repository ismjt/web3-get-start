package tests

import (
	"go-blog/models"
	"go-blog/system"
	"log"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	cfg = system.GetConfiguration()
)

func setupTestDB() *gorm.DB {
	testDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal("Failed to get absolute path:", err)
	}
	db, err := gorm.Open(sqlite.Open(filepath.Join(testDir, "..", "db", "personal_blog.db")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.User{})
	return db
}
