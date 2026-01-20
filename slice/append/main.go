package main

import "fmt"

func main() {
	data := make([]int, 0, 2)
	appendOne(data)
	fmt.Println(data)     // what does it print?
	fmt.Println(data[:1]) // what does it print?
}

func appendOne(data []int) {
	data = append(data, 1)
}
