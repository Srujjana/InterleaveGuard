package scenario

import (
    "sync"
    "testing"
)

func TestConcurrentMapAccess(t *testing.T) {
    m := &SharedMap{Data: make(map[int]*Entry)}

    for i := 0; i < 10; i++ {
        m.Data[i] = &Entry{}
    }

    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(2)

        go func(key int) {
            defer wg.Done()
            m.Write(key, key)
        }(i)

        go func(key int) {
            defer wg.Done()
            _ = m.Read(key)
        }(i)
    }

    wg.Wait()
}
