package main

import (
	"fmt"
)

type Circle struct {
	Radius float32
}
type Reactangle struct {
	Length float32
}

type Shape interface {
	Area() float32
}

func (c Circle) Area() float32 {
	return 33.3
}
func (c Reactangle) Area() float32 {
	return 33.3
}
func main() {
	b1 := []Shape{Circle{11.1}, Reactangle{12.3}}

	for _, v := range b1 {
		fmt.Println(v.Area())
	}

}
