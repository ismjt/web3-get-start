package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Employee struct {
	ID         int     `db:"id"`
	Name       string  `db:"name"`
	Department string  `db:"department"`
	Salary     float64 `db:"salary"`
}

// 查询技术部的所有员工
func GetTechEmployees(db *sqlx.DB) ([]Employee, error) {
	var employees []Employee
	query := `SELECT id, name, department, salary FROM employees WHERE department = ?`
	err := db.Select(&employees, query, "技术部")
	return employees, err
}

// 查询工资最高的员工
func GetTopEmployee(db *sqlx.DB) (*Employee, error) {
	var emp Employee
	query := `SELECT id, name, department, salary FROM employees ORDER BY salary DESC LIMIT 1`
	err := db.Get(&emp, query)
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func main() {
	// 1. 连接 SQLite 数据库
	db, err := sqlx.Open("sqlite3", "./company.db")
	if err != nil {
		log.Fatalln("数据库连接失败:", err)
	}
	defer db.Close()

	// 2. 确保 employees 表存在
	schema := `
	CREATE TABLE IF NOT EXISTS employees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		department TEXT,
		salary REAL
	);`
	db.MustExec(schema)

	// 3. 插入一些示例数据(先清空数据库数据)
	db.MustExec("DELETE FROM employees")
	db.MustExec("DELETE FROM sqlite_sequence WHERE name='employees'")
	db.MustExec(`INSERT INTO employees (name, department, salary) VALUES (?, ?, ?)`, "张三", "技术部", 8000)
	db.MustExec(`INSERT INTO employees (name, department, salary) VALUES (?, ?, ?)`, "李四", "市场部", 6000)
	db.MustExec(`INSERT INTO employees (name, department, salary) VALUES (?, ?, ?)`, "王五", "技术部", 12000)

	// 4. 查询技术部员工
	techEmps, err := GetTechEmployees(db)
	if err != nil {
		log.Fatalln("查询技术部员工失败:", err)
	}
	fmt.Println("所有技术部员工:")
	for _, e := range techEmps {
		fmt.Printf("%+v\n", e)
	}

	// 2. 查询工资最高的员工
	topEmp, err := GetTopEmployee(db)
	if err != nil {
		log.Fatalln("查询最高工资员工失败:", err)
	}
	fmt.Println("工资最高的员工:")
	fmt.Printf("%+v\n", *topEmp)
}
