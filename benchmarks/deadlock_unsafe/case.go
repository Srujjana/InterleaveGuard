package scenario

import "sync"

type LockPair struct {
	A sync.Mutex
	B sync.Mutex
}

// One worker acquires A, then B.
func (p *LockPair) AcquireAB(firstHeld chan<- struct{}, proceed <-chan struct{}) {
	p.A.Lock()
	defer p.A.Unlock()

	firstHeld <- struct{}{}
	<-proceed

	p.B.Lock()
	defer p.B.Unlock()
}

// Another worker acquires B, then A.
// If both hold their first lock simultaneously, they can deadlock.
func (p *LockPair) AcquireBA(firstHeld chan<- struct{}, proceed <-chan struct{}) {
	p.B.Lock()
	defer p.B.Unlock()

	firstHeld <- struct{}{}
	<-proceed

	p.A.Lock()
	defer p.A.Unlock()
}
