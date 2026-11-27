package scenario

import (
    "testing"
    "time"
)

func TestChannelDeadlock(t *testing.T) {
    aToB := make(chan int)
    bToA := make(chan int)

    done := make(chan bool, 2)

    go func() {
        WorkerA(aToB, bToA)
        done <- true
    }()

    go func() {
        WorkerB(aToB, bToA)
        done <- true
    }()

    select {
    case <-done:
        t.Fatal("expected workers to deadlock, but one completed")
    case <-time.After(250 * time.Millisecond):
        // Both workers are blocked on their first send.
    }
}
