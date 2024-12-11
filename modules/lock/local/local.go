package local

import (
	"sync"
	"time"
)

const DelayExpireDuration = time.Minute * 15

type localLock struct {
	m           *sync.Mutex
	expiredTime time.Time
}

func (l *localLock) lock() {
	l.m.Lock()
}

func (l *localLock) unlock() {
	l.m.Unlock()
}

func (l *localLock) TryLock() bool {
	return l.m.TryLock()
}

func (l *localLock) IsExpired() bool {
	return l.expiredTime.Add(DelayExpireDuration).Before(time.Now())
}

type LocalLockManager struct {
	m  *sync.RWMutex
	lm map[string]*localLock
}

func NewLocalLockManager() *LocalLockManager {
	return &LocalLockManager{
		lm: make(map[string]*localLock, 10000),
		m:  new(sync.RWMutex),
	}
}

func (llm *LocalLockManager) getLocalLock(key string) (lock *localLock, isOK bool) {
	llm.m.RLock()
	defer llm.m.RUnlock()
	lock, isOK = llm.lm[key]
	return
}

func (llm *LocalLockManager) setLocalLock(key string, lock *localLock) {
	llm.m.Lock()
	defer llm.m.Unlock()
	llm.lm[key] = lock
}

func (llm *LocalLockManager) Lock(key string, expiredTime time.Time) {
	lock, isOK := llm.getLocalLock(key)
	if !isOK {
		lock = &localLock{
			m:           new(sync.Mutex),
			expiredTime: expiredTime,
		}
		llm.setLocalLock(key, lock)
	}
	lock.lock()
}

func (llm *LocalLockManager) TryLock(key string, expiredTime time.Time) {
	lock, isOK := llm.getLocalLock(key)
	if !isOK {
		lock = &localLock{
			m:           new(sync.Mutex),
			expiredTime: expiredTime,
		}
		llm.setLocalLock(key, lock)
	}
	lock.TryLock()
}

func (llm *LocalLockManager) UnLock(key string) {
	if lock, isOK := llm.getLocalLock(key); isOK {
		lock.unlock()
	}
}

func (llm *LocalLockManager) Len() int {
	llm.m.RLock()
	defer llm.m.RUnlock()
	return len(llm.lm)
}

func (llm *LocalLockManager) Cleanup() {
	llm.m.Lock()
	defer llm.m.Unlock()
	for key, lm := range llm.lm {
		if lm.IsExpired() {
			delete(llm.lm, key)
		}
	}
}
