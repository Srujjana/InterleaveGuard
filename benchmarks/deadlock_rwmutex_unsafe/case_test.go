package scenario

import (
    "sync"
    "testing"
    "time"
)

func TestRWMutexUpgradeDeadlocks(t *testing.T) {
    var mu sync.RWMutex
    done := make(chan struct{})

    go func() {
        UpgradeLock(&mu)
        close(done)
    }()

    select {
    case <-done:
        t.Fatal("expected RWMutex upgrade to deadlock")
    case <-time.After(250 * time.Millisecond):
    }
}
