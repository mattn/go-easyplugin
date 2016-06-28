package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("foo")
	time.Sleep(1 * time.Second)
	fmt.Println("bar")
	time.Sleep(1 * time.Second)
	fmt.Println("baz")
	time.Sleep(1 * time.Second)
}
