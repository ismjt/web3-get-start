package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	ch := make(chan int)
	var wg sync.WaitGroup

	// 启动生产者协程：生成 1 到 10 的整数
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i++ {
			ch <- i // 发送数据到通道
			time.Sleep(time.Duration(time.Millisecond * 500))
		}
		close(ch)
	}()

	// 启动消费者协程：接收并打印
	wg.Add(1)
	go func() {
		defer wg.Done()
		for num := range ch { // 从通道接收数据，直到通道关闭
			fmt.Println("接收到：", num)
		}
	}()

	// 为了避免主协程提前退出，这里简单阻塞一下
	// fmt.Scanln()

	// 等待两个协程完成
	wg.Wait()
	fmt.Println("所有数据发送和接收完毕，程序退出。")
}
