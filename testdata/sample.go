package main

import "fmt"

func greet() {
	fmt.Println("Hello!")
}

func processData() {
	fmt.Println("Processing...")
}

func main() {
	go greet()
	go processData()
}
