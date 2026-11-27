package scenario

import (
    "sync"
    "testing"
)

func TestConcurrentIncrementSafe(t *testing.T) {
    counter := &SharedCounter{}
    var wg sync.WaitGroup

    for i := 0; i < 2; i++ {
        wg.Add(1)

        go func() {
            defer wg.Done()
            counter.Increment()
        }()
    }

    wg.Wait()

    if got := counter.Value(); got != 2 {
        t.Fatalf("expected value 2, got %d", got)
    }
}
