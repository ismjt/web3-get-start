package main

import (
	"fmt"
	"sync"
)

func main() {
	var counter int       // 共享计数器
	var mutex sync.Mutex  // 互斥锁
	var wg sync.WaitGroup // 用于等待所有协程完成

	numGoroutines := 10
	numIncrements := 1000

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIncrements; j++ {
				mutex.Lock()   // 上锁
				counter++      // 安全访问共享计数器
				mutex.Unlock() // 解锁
			}
		}()
	}

	wg.Wait() // 等待所有协程完成
	fmt.Println("阻塞式互斥锁 - 最终计数器值:", counter)
}
