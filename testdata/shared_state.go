package main

import "sync"

// Shared global state
var counter int
var mu sync.Mutex

// CASE 1: Direct read-modify-write
func incrementUnsafe() {
	counter++
}

// CASE 2: Read-only access
func readCounter() int {
	return counter
}

// CASE 3: Write-only access
func writeCounter(value int) {
	counter = value
}

// CASE 4: Access protected by a mutex
func incrementSafe() {
	mu.Lock()
	defer mu.Unlock()

	counter++
}

// CASE 5: Indirect access through a helper function
func helperUpdate() {
	incrementUnsafe()
}

// CASE 6: Pointer-based access (alias analysis comes later)
func updateThroughPointer(p *int) {
	*p = *p + 1
}
func shadowTest() {
	counter := 100
	counter++
}
func main() {
	go incrementUnsafe()
	go readCounter()
	go incrementSafe()
}
