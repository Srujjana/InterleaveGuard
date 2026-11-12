/*
Here the mutex protects both the increment and the read.
So we expect exactly 10,000.
*/
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

	close(start)
	wg.Wait()

	got := counter.Read()
	want := workers * incrementsPerWorker

	if got != want {
		t.Fatalf("counter = %d, want %d", got, want)
	}
}
