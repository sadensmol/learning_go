package main

import (
	"fmt"
)

func main() {

	fmt.Printf("x = %v\n", doSum(4, 5))
	t, t2 := doDivMod(4, 5)

	fmt.Printf("x = %v,%v\n", t, t2)
	x := 5
	mulBy2(&x)
	fmt.Printf("x = %v\n", x)

	//error always right! 
	_, err := returnError(11)
	fmt.Printf("this is error = %v\n", err)

}

func doSum(a int, b int) int {
	return a + b
}
func doDivMod(a int, b int) (int, int) {
	return a / b, a % b
}

func mulBy2(a *int) {
	*a *= 2
}

func returnError(a int) (int, error) {
	return 0, fmt.Errorf("fasfasdfas error!!!")
}
