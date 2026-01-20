package main

type reader interface {
	read()
}

type hello struct{}

func (*hello) read() {
	println("reading ...")
}

func set[T any, U interface {
	*T
	reader
}]() {
	var a U
	a.read()
}

func main() {

	set[hello]()
}
