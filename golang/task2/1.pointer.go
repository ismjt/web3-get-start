package main

import (
	"fmt"
)

// 指针
func pointer1(param *int) {
	*param = *param + 10
}

func pointer2(param *[]int) {
	for i := range *param {
		(*param)[i] *= 2
	}
}

func main() {
	fmt.Println("指针")
	p1 := 12
	fmt.Printf("题目1-入参: %v", p1)
	pointer1(&p1)
	fmt.Printf(" 输出结果: %v\n", p1)

	p2 := []int{2, 7, 11, 15}
	fmt.Printf("题目2-入参: %v", p2)
	pointer2(&p2)
	fmt.Printf(" 输出结果: %v\n", p2)
}
