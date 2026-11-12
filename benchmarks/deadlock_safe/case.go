package scenario

import "sync"

type LockPair struct {
	A sync.Mutex
	B sync.Mutex
}

// Both workers acquire A before B.
// Consistent ordering prevents a circular lock dependency.
func (p *LockPair) AcquireForWorkerA() {
	p.A.Lock()
	defer p.A.Unlock()

	p.B.Lock()
	defer p.B.Unlock()
}

func (p *LockPair) AcquireForWorkerB() {
	p.A.Lock()
	defer p.A.Unlock()

	p.B.Lock()
	defer p.B.Unlock()
}
