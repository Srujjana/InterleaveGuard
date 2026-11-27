package scenario

import "sync"

type Worker struct {
    mu sync.Mutex
}

func (w *Worker) Run() {
    w.mu.Lock()
    defer w.mu.Unlock()

    // Do the work while holding the lock.
}
