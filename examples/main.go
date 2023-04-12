package main

import (
	"fmt"
	"github.com/smartwalle/notification"
	"time"
)

func main() {
	notification.Default().Handle("test", handler1)
	notification.Default().Handle("test", handler2)

	for i := 0; i < 10; i++ {
		notification.Default().Dispatch("test", i)
	}
	time.Sleep(time.Second * 2)
}

func handler1(name string, value interface{}) {
	fmt.Println("handler1", name, value)
}

func handler2(name string, value interface{}) {
	fmt.Println("handler2", name, value)
}
