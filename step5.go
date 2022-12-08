package main

import (
	"fmt"
)

func main() {
	x := map[string]float32{"one": 1.0, "two": 2.0, "free": 3.0}
	fmt.Printf("x = %v (%T) %v\n", x, x, len(x))

	val, ok := x["four"]
	if !ok {
		fmt.Println("four is not found")
	} else {
		fmt.Printf("found value: %v", val)
	}
}
