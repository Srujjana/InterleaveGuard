package scenario

import (
    "sync"
    "testing"
)

func TestReproducesLostUpdate(t *testing.T) {
    counter := &SharedCounter{}

    readDone := make(chan struct{}, 2)
    writePermit := make(chan struct{})
    var wg sync.WaitGroup

    for i := 0; i < 2; i++ {
        wg.Add(1)

        go func() {
            defer wg.Done()
            counter.Increment(readDone, writePermit)
        }()
    }

    // Wait until both workers have read the initial value.
    <-readDone
    <-readDone

    // Allow both workers to write their updates.
    close(writePermit)
    wg.Wait()

    if got := counter.Value(); got != 1 {
        t.Fatalf("expected lost update (value 1), got %d", got)
    }
}
