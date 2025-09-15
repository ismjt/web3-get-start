package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var counter int64
	var wg sync.WaitGroup

	numGoroutines := 10
	numIncrements := 1000

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIncrements; j++ {
				atomic.AddInt64(&counter, 1) // 原子递增
			}
		}()
	}

	wg.Wait() // 等待所有协程完成
	fmt.Println("原子操作 - 最终计数器值:", counter)
}
