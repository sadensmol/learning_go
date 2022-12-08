package main

import (
	"fmt"
)

func main() {
	x := []string{"one", "two", "free"}
	fmt.Printf("x = %v (%T) %v\n", x, x, len(x))

	for i := range x {
		fmt.Println(i)
	}
	for i, val := range x {
		fmt.Printf("%d %s", i, val)
	}
}
