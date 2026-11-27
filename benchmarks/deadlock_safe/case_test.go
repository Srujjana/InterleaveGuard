package scenario

import (
    "testing"
    "time"
)

func TestConsistentLockOrderCompletes(t *testing.T) {
    pair := &LockPair{}
    done := make(chan struct{}, 2)

    go func() {
        pair.AcquireForWorkerA()
        done <- struct{}{}
    }()

    go func() {
        pair.AcquireForWorkerB()
        done <- struct{}{}
    }()

    // Both workers should complete within the timeout.
    for i := 0; i < 2; i++ {
        select {
        case <-done:
        case <-time.After(2 * time.Second):
            t.Fatal("worker failed to complete; possible deadlock")
        }
    }
}
