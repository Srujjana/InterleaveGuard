package scenario

import "sync"

type Job struct {
	mu      sync.Mutex
	claimed bool
}

// TryClaim deliberately separates checking from claiming.
// The channels let a test force two workers to pass the check first.
func (j *Job) TryClaim(afterCheck chan<- struct{}, resume <-chan struct{}) bool {
	j.mu.Lock()
	available := !j.claimed //checks availability
	j.mu.Unlock()

	if !available {
		return false
	}

	afterCheck <- struct{}{}
	<-resume

	j.mu.Lock()
	j.claimed = true //claims the job. two separate critical sections, so two workers can both claim the job.
	j.mu.Unlock()
	/*mu ensures only 1 worker claims the job at a time.
	  but it doesn't ensure that the entire operation of checking and claiming
	  happens atomically (the gap btw checking and claiming)*/
	return true
}

/*correct implementation:
func (j *Job) TryClaim() bool {
    j.mu.Lock()
    defer j.mu.Unlock()

    if j.claimed {
        return false
    }

    j.claimed = true
    return true
}
	mutex is held throughout the operation*/
