package syncx

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type spinlock uint32

const maxBackoff = 16

func (s *spinlock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(s), 0, 1) {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < maxBackoff {
			backoff <<= 1
		}
	}
}

func (s *spinlock) Unlock() {
	atomic.StoreUint32((*uint32)(s), 0)
}

func NewSpinLock() sync.Locker {
	return new(spinlock)
}
