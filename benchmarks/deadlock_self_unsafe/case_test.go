package scenario

import (
    "testing"
    "time"
)

func TestSelfDeadlock(t *testing.T) {
    worker := &Worker{}
    done := make(chan struct{})

    go func() {
        worker.Run()
        close(done)
    }()

    select {
    case <-done:
        t.Fatal("expected self-deadlock, but worker completed")
    case <-time.After(250 * time.Millisecond):
        // Expected: second Lock blocks forever.
    }
}
