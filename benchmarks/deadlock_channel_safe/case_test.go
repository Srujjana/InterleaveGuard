package scenario

import (
    "testing"
    "time"
)

func TestChannelCoordinationCompletes(t *testing.T) {
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

    for i := 0; i < 2; i++ {
        select {
        case <-done:
        case <-time.After(2 * time.Second):
            t.Fatal("expected both workers to complete")
        }
    }
}
