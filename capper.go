package main

import (
	"fmt"
	"io"
	"os"
)

type Capper struct {
	PPP io.Writer
}

func (c *Capper) dWrite(p []byte) (n int, err error) {
	c.PPP.Write(p)
	return len(p), nil
}

func main() {
	c := &Capper{os.Stdout}
	fmt.Fprintln(c, "Hello there!")

}
