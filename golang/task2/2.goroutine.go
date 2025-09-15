package main

import (
	"fmt"
	"sync"
	"time"
)

// Goroutine
func printOdd() {
	for i := 1; i < 10; i += 2 {
		fmt.Println("奇数协程:", i)
		time.Sleep(100 * time.Millisecond)
	}
}
func printEven() {
	for i := 2; i <= 10; i += 2 {
		fmt.Println("偶数协程:", i)
		time.Sleep(200 * time.Millisecond)
	}
}

type Task func()

type Result struct {
	ID       int
	Duration time.Duration
}

// 调度器：并发执行任务并统计时间
func runTasks(tasks []Task) []Result {
	results := make([]Result, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)

		// 启动协程
		go func(id int, t Task) {
			defer wg.Done()

			start := time.Now()
			t() // 执行任务
			duration := time.Since(start)

			fmt.Printf("任务 %d 完成，耗时: %v\n", id, duration)
			results[i] = Result{ID: i, Duration: time.Since(start)}
		}(i, task)
	}

	// 等待所有任务完成
	wg.Wait()

	return results
}

func main() {

	//fmt.Println("Goroutine - 题目1")
	//go printOdd()
	//go printEven()

	fmt.Println("Goroutine - 题目2")
	tasks := []Task{
		func() {
			time.Sleep(300 * time.Millisecond)
			fmt.Println("任务 1: 模拟 I/O 操作完成")
		},
		printOdd,
		printEven,
	}
	// 调用调度器
	runTasks(tasks)
	// 展示每个任务的执行时间
	fmt.Printf("Goroutine - 题目2 - ")
}
