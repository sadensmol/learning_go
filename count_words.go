package main

import (
	"fmt"
	"strings"
)

func main() {
	words := "one big city lights turn lights on every one night"
	fmt.Printf("text = %v", words)

	wa := strings.Split(words, " ")

	wm := map[string]int{}

	for _, val := range wa {
		vvv, ok := wm[val]

		if ok {
			wm[val] = vvv + 1
		} else {
			wm[val] = 1
		}
	}

	fmt.Println("%v", wm)
}
