package scenario

import (
	"sync"
	"testing"
)

func TestOnlyOneWorkerCanClaim(t *testing.T) {
	job := &Job{}

	const workers = 2

	start := make(chan struct{})
	results := make(chan bool, workers)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start
			results <- job.TryClaim()
		}()
	}

	// Release both workers together.
	close(start)

	wg.Wait()
	close(results)

	successes := 0

	for result := range results {
		if result {
			successes++
		}
	}

	if successes != 1 {
		t.Fatalf("expected exactly one successful claim, got %d", successes)
	}

	t.Log("SAFE: exactly one worker claimed the job")
}
