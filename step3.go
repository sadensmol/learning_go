package main

import (
	"fmt"
)

func main() {
	x := 3

	if x > 5 {
		fmt.Println("x is bigger that 5!!!")
	} else {
		fmt.Println("x is not that big as 5")
	}

	a, b := 12.0, 22.0

	if frac := a / b; frac > 0.5 {
		fmt.Println("frac is bigger than 0.5")
	}

	n := 2
	switch n {
	case 1:
		fmt.Println("one")
	case 2:
		fmt.Println("two")
	default:
		fmt.Println("default")
	}
}
