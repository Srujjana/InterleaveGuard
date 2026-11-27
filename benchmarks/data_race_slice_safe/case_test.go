package scenario

import (
    "sync"
    "testing"
)

func TestConcurrentAppendSafe(t *testing.T) {
    shared := &SharedSlice{}
    var wg sync.WaitGroup

    for i := 0; i < 8; i++ {
        wg.Add(1)

        go func(worker int) {
            defer wg.Done()

            for j := 0; j < 1000; j++ {
                shared.Append(worker*1000 + j)
            }
        }(i)
    }

    wg.Wait()

    if got := shared.Length(); got != 8000 {
        t.Fatalf("expected 8000 elements, got %d", got)
    }
}
