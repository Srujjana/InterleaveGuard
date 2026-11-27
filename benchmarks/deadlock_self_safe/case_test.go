package scenario

import (
    "testing"
    "time"
)

func TestWorkerCompletes(t *testing.T) {
    worker := &Worker{}
    done := make(chan struct{})

    go func() {
        worker.Run()
        close(done)
    }()

    select {
    case <-done:
    case <-time.After(2 * time.Second):
        t.Fatal("worker did not complete")
    }
}
