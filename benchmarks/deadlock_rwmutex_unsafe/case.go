package scenario

import "sync"

func UpgradeLock(mu *sync.RWMutex) {
    mu.RLock()
    defer mu.RUnlock()

    mu.Lock()
    mu.Unlock()
}
