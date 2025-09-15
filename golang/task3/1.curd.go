package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Student struct {
	ID    uint   `gorm:"primaryKey"` // 自增主键
	Name  string // 学生姓名
	Age   int    // 学生年龄
	Grade string // 学生年级
}

// 查询所有学生
func GetAllStudents(db *gorm.DB) []Student {
	var all []Student
	db.Find(&all)
	fmt.Println("查询验证 - 所有学生信息:")
	for _, s := range all {
		fmt.Printf("ID:%d, 姓名:%s, 年龄:%d, 年级:%s\n\n", s.ID, s.Name, s.Age, s.Grade)
	}
	return all
}

func main() {
	// 1. 连接 SQLite 数据库
	db, err := gorm.Open(sqlite.Open("students.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 2. 自动迁移表结构
	err = db.AutoMigrate(&Student{})
	if err != nil {
		log.Fatal("自动迁移失败:", err)
	}

	// 清空数据便于测试
	db.Where("1 = 1").Delete(&Student{})

	// 3. 插入单条记录
	db.Create(&Student{Name: "张三", Age: 20, Grade: "三年级"})

	// 4. 批量插入多条记录
	//students := []Student{
	//	{Name: "李四", Age: 19, Grade: "三年级"},
	//	{Name: "王五", Age: 18, Grade: "二年级"},
	//}
	//db.Create(&students)

	// 5. 查询所有年龄大于18岁的学生信息
	var students []Student
	if err := db.Where("age > ?", 18).Find(&students).Error; err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Println("年龄大于18岁的学生信息:")
	for _, s := range students {
		fmt.Printf("ID:%d, 姓名:%s, 年龄:%d, 年级:%s\n\n", s.ID, s.Name, s.Age, s.Grade)
	}

	// 6. 更新张三的年级
	db.Model(&Student{}).Where("name = ?", "张三").Update("Grade", "四年级")
	fmt.Println("张三年级已更新为四年级\n")
	GetAllStudents(db)

	// 7. 删除年龄小于15岁的学生
	db.Where("age < ?", 15).Delete(&Student{})
	fmt.Println("删除年龄小于15岁的学生完成\n")
	GetAllStudents(db)
}
