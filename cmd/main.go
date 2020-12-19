package main

import (
	"fmt"
	"time"
	bbmetric "github.com/buildboxapp/lib/metric"
)

func main() {
	metric := bbmetric.New(nil, nil, 10 * time.Second)

	metric.SetConnectionIncrement()

	fmt.Println(metric.Get())

	t :=  10 * time.Second
	f := t.Microseconds()
	fmt.Println(f)
	return
}