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
	counter := &Counter{} //shared counter

	const workers = 10
	const incrementsPerWorker = 1000

	start := make(chan struct{}) //start channel to sync the start of all routines
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() { //launching 10 goroutines that'll increment counter
			defer wg.Done()
			<-start //every goroutine waits for the start signal

			for j := 0; j < incrementsPerWorker; j++ {
				counter.Increment() //each routine increments the var a 1000 times
			}
		}()
	}

	// Release all workers together to encourage overlapping accesses.
	close(start) //signals all routines to start

	wg.Wait() //wait for all routines to finish

	// Read after all workers finish.
	// The race detector should still detect the concurrent writes.
	_ = counter.Read()
}
