package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Book struct {
	ID     int     `db:"id"`
	Title  string  `db:"title"`
	Author string  `db:"author"`
	Price  float64 `db:"price"`
}

// AuthorStats 用于存放作者分组统计结果
type AuthorStats struct {
	Author    string  `db:"author"`
	AvgPrice  float64 `db:"avg_price"`
	BookCount int     `db:"book_count"`
}

func DbPrepare() *sqlx.DB {
	// 1. 连接 SQLite 数据库
	db, err := sqlx.Open("sqlite3", "./books.db")
	if err != nil {
		log.Fatalln("数据库连接失败:", err)
	}

	// 2. 确保 books 表存在
	schema := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		author TEXT,
		price REAL
	);`
	db.MustExec(schema)

	// 3. 准备好数据，便于后续测试
	db.MustExec("DELETE FROM books")
	db.MustExec("DELETE FROM sqlite_sequence WHERE name='books'")
	db.MustExec(`INSERT INTO books (title, author, price) VALUES (?, ?, ?)`, "Go语言从入门到放弃", "Majt", 33.3)
	db.MustExec(`INSERT INTO books (title, author, price) VALUES (?, ?, ?)`, "数据库菜鸟教程", "Alan", 20.2)
	db.MustExec(`INSERT INTO books (title, author, price) VALUES (?, ?, ?)`, "许三观卖血记", "Anonymous", 35.5)
	db.MustExec(`INSERT INTO books (title, author, price) VALUES (?, ?, ?)`, "Golang是世界上最好的语言", "Majt", 77.6)
	db.MustExec(`INSERT INTO books (title, author, price) VALUES (?, ?, ?)`, "Go并发编程实战", "Alan", 36.0)
	db.MustExec(`INSERT INTO books (title, author, price) VALUES (?, ?, ?)`, "Mysql数据库系统概念", "Alan", 10.1)
	db.MustExec(`INSERT INTO books (title, author, price) VALUES (?, ?, ?)`, "小王子", "Anonymous", 22.0)
	db.MustExec(`INSERT INTO books (title, author, price) VALUES (?, ?, ?)`, "夜航", "Anonymous", 5.9)

	return db
}

// 价格大于 50 的书籍
func GetBooksByPrice(db *sqlx.DB, price float64) ([]Book, error) {
	var books []Book
	query := `SELECT id, title, author, price FROM books WHERE price > ? ORDER BY price DESC`
	err := db.Select(&books, query, 50)
	if err != nil {
		log.Fatalln("查询失败:", err)
	}

	// 输出查询结果
	fmt.Println("价格大于 50 元的书籍:")
	for _, b := range books {
		fmt.Printf("ID:%d, 标题:%s, 作者:%s, 价格:%.2f\n", b.ID, b.Title, b.Author, b.Price)
	}

	return books, err
}

func AuthorStat(db *sqlx.DB) ([]AuthorStats, error) {
	var stats []AuthorStats
	query := `
		SELECT author,
		       AVG(price) AS avg_price,
		       COUNT(*) AS book_count
		FROM books
		GROUP BY author
		ORDER BY avg_price DESC
	`
	err := db.Select(&stats, query)
	if err != nil {
		log.Fatalln("分组统计失败:", err)
	}

	// 输出结果
	fmt.Println("作者平均价格统计:")
	for _, s := range stats {
		fmt.Printf("作者:%s, 平均价格:%.2f, 书籍数量:%d\n", s.Author, s.AvgPrice, s.BookCount)
	}

	return stats, err
}

func main() {
	// 1.清空表并插入一些示例数据(先清空数据库数据)
	db := DbPrepare()
	defer db.Close()

	// 2. 复杂查询案例
	GetBooksByPrice(db, 50)
	AuthorStat(db)
}
