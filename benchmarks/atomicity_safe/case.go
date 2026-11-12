package scenario

import "sync"

type Job struct {
	mu      sync.Mutex
	claimed bool
}

// TryClaim performs the check and update as one atomic critical section.
func (j *Job) TryClaim() bool {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.claimed {
		return false
	}

	j.claimed = true
	return true
}
