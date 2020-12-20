package main

import (
	"fmt"
	"runtime"
)

func main() {
	//var ms1 runtime.BlockProfileRecord

	ms1 := runtime.NumGoroutine()

	fmt.Println(ms1)
	return
}