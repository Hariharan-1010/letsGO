package main

import "fmt"

// func main() {
// 	announce("hi")
// 	fmt.Println("Hello, world!")
// }

func announce(message string) {
	go func() {
		fmt.Println(message)
	}()
}
