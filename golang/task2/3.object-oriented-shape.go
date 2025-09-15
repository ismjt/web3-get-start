package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

// 矩形结构体
type Rectangle struct {
	Width, Height float64
}

// 实现 Shape 接口的方法
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// 圆形结构体
type Circle struct {
	Radius float64
}

// 实现 Shape 接口的方法
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func main() {
	// 创建 Rectangle 和 Circle 实例
	r := Rectangle{Width: 5, Height: 3}
	c := Circle{Radius: 4}
	var s Shape
	s = r
	fmt.Printf("矩形: 面积=%.2f, 周长=%.2f\n", s.Area(), s.Perimeter())
	s = c
	fmt.Printf("圆形: 面积=%.2f, 周长=%.2f\n", s.Area(), s.Perimeter())
}
