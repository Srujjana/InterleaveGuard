package main

var sharedValue int

func modifyValue() {
	sharedValue++
}

func helper() {
	modifyValue()
}

func main() {
	go helper()
	go func() {
		sharedValue = 10
	}()
}
