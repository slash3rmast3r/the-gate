package authz

import (
	"sync"
	"time"
)

type attempt struct {
	fails    int
	lockedUntil time.Time
}

// BruteForce guards login endpoints in-memory (per process).
type BruteForce struct {
	mu       sync.Mutex
	byKey    map[string]*attempt
	maxFails int
	window   time.Duration
	lockFor  time.Duration
}

func NewBruteForce(maxFails int, window, lockFor time.Duration) *BruteForce {
	if maxFails < 1 {
		maxFails = 5
	}
	return &BruteForce{
		byKey:    make(map[string]*attempt),
		maxFails: maxFails,
		window:   window,
		lockFor:  lockFor,
	}
}

func (b *BruteForce) Allowed(key string) (bool, time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	a := b.byKey[key]
	if a == nil {
		return true, 0
	}
	if time.Now().Before(a.lockedUntil) {
		return false, time.Until(a.lockedUntil)
	}
	return true, 0
}

func (b *BruteForce) Fail(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	a := b.byKey[key]
	if a == nil {
		a = &attempt{}
		b.byKey[key] = a
	}
	a.fails++
	if a.fails >= b.maxFails {
		a.lockedUntil = time.Now().Add(b.lockFor)
		a.fails = 0
	}
}

func (b *BruteForce) Success(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.byKey, key)
}