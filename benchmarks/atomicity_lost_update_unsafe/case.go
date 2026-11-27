package scenario

import "sync"

type SharedCounter struct {
    mu    sync.Mutex
    value int
}

// The read and write are protected separately,
// but the complete update is not protected.
func (c *SharedCounter) Increment(readDone chan<- struct{}, writePermit <-chan struct{}) {
    c.mu.Lock()
    current := c.value
    c.mu.Unlock()

    // Wait until both workers have read the old value.
    readDone <- struct{}{}
    <-writePermit

    c.mu.Lock()
    c.value = current + 1
    c.mu.Unlock()
}

func (c *SharedCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()

    return c.value
}
