package scenario

import (
	"sync"
	"testing"
)

func TestReproducesDoubleClaim(t *testing.T) {
	job := &Job{}

	// Both workers signal after checking availability.
	afterCheck := make(chan struct{}, 2)

	// They cannot proceed to claim until we release them.
	resume := make(chan struct{})

	results := make(chan bool, 2)

	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			results <- job.TryClaim(afterCheck, resume)
		}()
	}

	// Wait until BOTH workers have checked the job.
	<-afterCheck
	<-afterCheck

	// Both workers have observed the job as available.
	// Now allow both to perform the claim.
	close(resume)

	wg.Wait()
	close(results)

	successes := 0

	for result := range results {
		if result {
			successes++
		}
	}

	// The intended behaviour is exactly one successful claim.
	// Two successes demonstrate the atomicity violation.
	if successes != 2 {
		t.Fatalf("expected to reproduce double claim (2 successes), got %d", successes)
	}

	t.Logf("BUG REPRODUCED: both workers successfully claimed the same job")
}
