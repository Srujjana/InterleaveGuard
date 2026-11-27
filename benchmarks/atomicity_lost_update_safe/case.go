package scenario

import "sync"

type SharedCounter struct {
    mu    sync.Mutex
    value int
}

// The mutex protects the complete read-modify-write operation.
func (c *SharedCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.value++
}

func (c *SharedCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()

    return c.value
}
