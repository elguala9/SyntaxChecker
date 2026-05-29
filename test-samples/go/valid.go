package main

import "fmt"

func main() {
	msg := greet("world")
	fmt.Println(msg)
}

func greet(name string) string {
	return fmt.Sprintf("hello, %s", name)
}
