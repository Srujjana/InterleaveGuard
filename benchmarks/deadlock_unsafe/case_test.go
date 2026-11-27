package scenario

import (
    "testing"
    "time"
)

func TestOppositeLockOrderDeadlocks(t *testing.T) {
    pair := &LockPair{}

    firstHeld := make(chan struct{}, 2)
    proceed := make(chan struct{})
    done := make(chan struct{}, 2)

    go func() {
        pair.AcquireAB(firstHeld, proceed)
        done <- struct{}{}
    }()

    go func() {
        pair.AcquireBA(firstHeld, proceed)
        done <- struct{}{}
    }()

    // Ensure both goroutines hold their first lock.
    for i := 0; i < 2; i++ {
        select {
        case <-firstHeld:
        case <-time.After(2 * time.Second):
            t.Fatal("workers did not acquire their first locks")
        }
    }

    // Both workers now attempt to acquire the other worker's lock.
    close(proceed)

    // Neither worker should finish because they are deadlocked.
    select {
    case <-done:
        t.Fatal("expected deadlock, but a worker completed")
    case <-time.After(250 * time.Millisecond):
        // Expected: neither worker completed.
    }
}
