package scenario

import "testing"

func TestWaitReturnsAfterWorkerFinishes(t *testing.T) {
    finished := RunWorker()

    if !finished {
        t.Fatal("expected worker to finish before Wait returns")
    }
}
