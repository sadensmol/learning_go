package main

import (
	"fmt"
)

func main() {
	ch := make(chan int)

	go func() {
		ch <- 111
		ch <- 222
		close(ch)
	}()

	//read all
	for val := range ch {
		fmt.Printf("%v\n", val)
	}

	ch2:= make (chan int, 2)
	ch2<-123
	ch2<-223
	val2:=<-ch2
	val3:=<-ch2
	fmt.Println(val2)
	fmt.Println(val3)
}

func send() {

}

func receive() {
}
