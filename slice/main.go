package main

import "fmt"

func growSlice(s []int) {
	t := append(s, 3)
	fmt.Printf("Inside: %v\n", t)
}

func main() {
	s := make([]int, 0, 10)
	s = append(s, 1, 2)
	fmt.Printf("Before: %v\n", s)
	growSlice(s)

	fmt.Printf("After: %v\n", s)
	s2 := s[:3]
	fmt.Printf("subb: %v\n", s2)
}
