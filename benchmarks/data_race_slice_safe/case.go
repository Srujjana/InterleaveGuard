package scenario

import "sync"

type SharedSlice struct {
    mu     sync.Mutex
    Values []int
}

func (s *SharedSlice) Append(value int) {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.Values = append(s.Values, value)
}

func (s *SharedSlice) Length() int {
    s.mu.Lock()
    defer s.mu.Unlock()

    return len(s.Values)
}
