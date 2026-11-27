package scenario

import (
    "sync"
    "testing"
)

func TestConcurrentAppend(t *testing.T) {
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
}
