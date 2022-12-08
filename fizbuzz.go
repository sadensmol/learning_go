package main

import (
	"fmt"
)

func main() {

	x := 20

	for i := 0; i < x; i++ {
		if i%3 == 0 {
			fmt.Print("fizz")
		} 
		if i%5 == 0 {
			fmt.Print("buzz")
		} 
		if i%3 != 0 && i%5 != 0 {
			fmt.Printf("%v\n", i)
		}

		if i%3 == 0 || i%5 == 0 {
			fmt.Print("\n")
		}
	}
}
