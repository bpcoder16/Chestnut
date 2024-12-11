package local

import (
	"sync"
	"time"
)

const DelayExpireDuration = time.Minute * 15

type lock struct {
	m           *sync.Mutex
	expiredTime time.Time
}

func (l *lock) lock() {
	l.m.Lock()
}

func (l *lock) unlock() {
	l.m.Unlock()
}

func (l *lock) TryLock() bool {
	return l.m.TryLock()
}

func (l *lock) IsExpired() bool {
	return l.expiredTime.Add(DelayExpireDuration).Before(time.Now())
}

type LockManager struct {
	m  *sync.RWMutex
	lm map[string]*lock
}

func NewLocalLockManager() *LockManager {
	return &LockManager{
		lm: make(map[string]*lock, 10000),
		m:  new(sync.RWMutex),
	}
}

func (llm *LockManager) getLocalLock(key string) (lock *lock, isOK bool) {
	llm.m.RLock()
	defer llm.m.RUnlock()
	lock, isOK = llm.lm[key]
	return
}

func (llm *LockManager) setLocalLock(key string, lock *lock) {
	llm.m.Lock()
	defer llm.m.Unlock()
	llm.lm[key] = lock
}

func (llm *LockManager) Lock(key string, expiredTime time.Time) {
	localLock, isOK := llm.getLocalLock(key)
	if !isOK {
		localLock = &lock{
			m:           new(sync.Mutex),
			expiredTime: expiredTime,
		}
		llm.setLocalLock(key, localLock)
	}
	localLock.lock()
}

func (llm *LockManager) TryLock(key string, expiredTime time.Time) {
	localLock, isOK := llm.getLocalLock(key)
	if !isOK {
		localLock = &lock{
			m:           new(sync.Mutex),
			expiredTime: expiredTime,
		}
		llm.setLocalLock(key, localLock)
	}
	localLock.TryLock()
}

func (llm *LockManager) UnLock(key string) {
	if localLock, isOK := llm.getLocalLock(key); isOK {
		localLock.unlock()
	}
}

func (llm *LockManager) Len() int {
	llm.m.RLock()
	defer llm.m.RUnlock()
	return len(llm.lm)
}

func (llm *LockManager) Cleanup() {
	llm.m.Lock()
	defer llm.m.Unlock()
	for key, lm := range llm.lm {
		if lm.IsExpired() {
			delete(llm.lm, key)
		}
	}
}
