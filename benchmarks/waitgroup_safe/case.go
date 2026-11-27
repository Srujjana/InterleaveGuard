package scenario

import "sync"

func RunWorker() bool {
    var wg sync.WaitGroup
    finished := false

    wg.Add(1)

    go func() {
        defer wg.Done()
        finished = true
    }()

    wg.Wait()
    return finished
}
