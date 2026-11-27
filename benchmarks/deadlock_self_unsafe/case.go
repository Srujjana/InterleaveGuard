package scenario

import "sync"

type Worker struct {
    mu sync.Mutex
}

func (w *Worker) Run() {
    w.mu.Lock()
    w.mu.Lock()

    w.mu.Unlock()
    w.mu.Unlock()
}
