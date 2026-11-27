package scenario

import "testing"

func TestWaitCanReturnBeforeWorkerFinishes(t *testing.T) {
    finished := RunWorker()

    if finished {
        t.Fatal("expected to reproduce premature WaitGroup completion")
    }
}
