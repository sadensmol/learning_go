package main

import (
	"fmt"
)

func main() {
	x := []int{16,8,42,4,23,15}

	mmm:=0
	for _,m:= range x {
		if m>mmm {
			mmm=m
		} 
	}

	fmt.Printf("max = %v\n", mmm)

}
