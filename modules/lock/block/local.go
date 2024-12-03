package block

import (
	"sync"
	"time"
)

const DelayExpireDuration = time.Minute * 15

type LocalLock struct {
	m           *sync.Mutex
	expiredTime time.Time
}

func (l *LocalLock) Lock() {
	l.m.Lock()
}

func (l *LocalLock) Unlock() {
	l.m.Unlock()
}

func (l *LocalLock) TryLock() bool {
	return l.m.TryLock()
}

func (l *LocalLock) IsExpired() bool {
	return l.expiredTime.Add(DelayExpireDuration).Before(time.Now())
}

type LocalLockManager struct {
	m  *sync.RWMutex
	lm map[string]*LocalLock
}

func NewLocalLockManager() *LocalLockManager {
	return &LocalLockManager{
		lm: make(map[string]*LocalLock),
		m:  new(sync.RWMutex),
	}
}

func (llm *LocalLockManager) getLocalLock(key string) (lock *LocalLock, isOK bool) {
	llm.m.RLock()
	defer llm.m.RUnlock()
	lock, isOK = llm.lm[key]
	return
}

func (llm *LocalLockManager) setLocalLock(key string, lock *LocalLock) {
	llm.m.Lock()
	defer llm.m.Unlock()
	llm.lm[key] = lock
}

func (llm *LocalLockManager) Lock(key string, expiredTime time.Time) {
	lock, isOK := llm.getLocalLock(key)
	if !isOK {
		lock = &LocalLock{
			m:           new(sync.Mutex),
			expiredTime: expiredTime,
		}
		llm.setLocalLock(key, lock)
	}
	lock.Lock()
}

func (llm *LocalLockManager) UnLock(key string) {
	if lock, isOK := llm.getLocalLock(key); isOK {
		lock.Unlock()
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
