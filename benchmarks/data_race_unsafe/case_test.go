//Why don't we assert that the final value is 10,000?

/*Because this test is designed to expose a data race, not to
depend on a particular lost-update outcome. The race detector can identify
conflicting unsynchronized accesses even if the final count happens to be correct.*/

package scenario

import (
	"sync"
	"testing"
)

func TestConcurrentIncrement(t *testing.T) {
	counter := &Counter{}

	const workers = 10
	const incrementsPerWorker = 1000

	start := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			<-start

			for j := 0; j < incrementsPerWorker; j++ {
				counter.Increment()
			}
		}()
	}

	// Release all workers together to encourage overlapping accesses.
	close(start)

	wg.Wait()

	// Read after all workers finish.
	// The race detector should still detect the concurrent writes.
	_ = counter.Read()
}
