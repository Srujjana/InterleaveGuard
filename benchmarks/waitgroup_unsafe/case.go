package scenario

import "sync"

func RunWorker() bool {
    var wg sync.WaitGroup
    started := make(chan struct{})
    release := make(chan struct{})
    finished := false

    go func() {
        close(started)
        <-release
        wg.Add(1)
        defer wg.Done()
        finished = true
    }()

    <-started
    wg.Wait()
    close(release)

    return finished
}
