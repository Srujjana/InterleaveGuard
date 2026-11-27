package scenario

import (
    "testing"
    "time"
)

func TestSendWithoutReceiverDeadlocks(t *testing.T) {
    ch := make(chan int)
    done := make(chan struct{})

    go func() {
        SendWithoutReceiver(ch)
        close(done)
    }()

    select {
    case <-done:
        t.Fatal("expected send to block without a receiver")
    case <-time.After(250 * time.Millisecond):
    }
}
