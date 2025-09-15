package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

type Employee struct {
	EmployeeID string
	Person
}

func (e Employee) PrintInfo() {
	fmt.Printf("Employee %s Info - Name:%s, Age:%d\n", e.EmployeeID, e.Name, e.Age)
}

func main() {
	p1 := Employee{
		EmployeeID: "YG001",
		Person: Person{
			Name: "James",
			Age:  32,
		},
	}
	p1.PrintInfo()

	p2 := Employee{
		EmployeeID: "YG002",
		Person: Person{
			Name: "张三",
			Age:  21,
		},
	}
	p2.PrintInfo()
}
