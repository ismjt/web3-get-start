package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// 设置50个用来展示缓冲如何影响生产消费的节奏
	ch := make(chan int, 50)
	var wg sync.WaitGroup

	// 启动生产者协程：生成 1 到 10 的整数
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("生产者协程 开始")
		for i := 1; i <= 100; i++ {
			ch <- i // 发送数据到通道
			// time.Sleep(time.Duration(time.Millisecond * 100)) // 模拟生产延迟
		}
		fmt.Println("生产者协程 结束")
		close(ch)
	}()

	// 启动消费者协程：接收并打印
	wg.Add(1)
	go func() {
		defer wg.Done()
		for num := range ch { // 从通道接收数据，直到通道关闭
			time.Sleep(time.Duration(time.Millisecond * 100)) // 模拟消费延迟
			fmt.Println("接收到：", num)
		}
	}()

	// 为了避免主协程提前退出，这里简单阻塞一下
	// fmt.Scanln()

	// 等待两个协程完成
	wg.Wait()
	fmt.Println("带有缓冲的通道 - 所有数据发送和接收完毕，程序退出。")
}
