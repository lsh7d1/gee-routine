package main

import (
	"fmt"
	"fucker"
	"sync"
	"time"
)

func DoSth() {
	time.Sleep(time.Millisecond * 10)
	fmt.Println("Hello World!")
}

func main() {
	var wg sync.WaitGroup
	t := 0
	syncDoSth := func() {
		DoSth()
		fmt.Println(t)
		t++
		wg.Done()
	}

	start := time.Now()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		fucker.Submit(syncDoSth)
	}
	time.Sleep(time.Second)
	fmt.Println(time.Since(start))
}
