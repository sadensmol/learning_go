package main

import (
	"fmt"
	"os"
	"github.com/pkg/errors"
)



func main() {
	_,err:=os.Open("test.xts")
	if err!=nil{
		fmt.Printf( "error: %+v",errors.Wrap(err, "fadfasfas"))
	}
}
