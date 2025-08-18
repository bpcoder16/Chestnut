package localpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// RWLockPool 维护按 Key 分配的 *sync.RWMutex，并在空闲超过 ttl 后回收。
type RWLockPool[K comparable] struct {
	mu     sync.Mutex
	m      map[K]*Entry[K]
	ttl    time.Duration
	sweep  time.Duration
	closed chan struct{}
	wg     sync.WaitGroup
}

type Entry[K comparable] struct {
	key      K
	rw       sync.RWMutex
	refs     atomic.Int32
	lastUsed atomic.Int64
	p        *RWLockPool[K]
}

// NewRWLockPool 创建带定期清理的锁池。
// ttl: 锁在 无人持有(refs=0) 状态下允许存活的时间
// sweep: 清理周期；建议为 ttl 的 1/2~1/5
func NewRWLockPool[K comparable](ttl, sweep time.Duration) *RWLockPool[K] {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if sweep <= 0 {
		sweep = ttl / 2
		if sweep <= 0 {
			sweep = 5 * time.Second
		}
	}
	p := &RWLockPool[K]{
		m:      make(map[K]*Entry[K]),
		ttl:    ttl,
		sweep:  sweep,
		closed: make(chan struct{}),
	}
	p.wg.Add(1)
	go p.janitor()
	return p
}

// Close 停止清理器并清空池子（不会强制打断正在持有的锁）。
func (p *RWLockPool[K]) Close() {
	close(p.closed)
	p.wg.Wait()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m = make(map[K]*Entry[K]) // 让 GC 回收
}

func (p *RWLockPool[K]) Len() int {
	return len(p.m)
}

// ---- 内部辅助 ----
func (p *RWLockPool[K]) janitor() {
	defer p.wg.Done()
	t := time.NewTicker(p.sweep)
	defer t.Stop()
	for {
		select {
		case <-p.closed:
			return
		case <-t.C:
			now := time.Now().UnixNano()
			p.mu.Lock()
			for k, e := range p.m {
				if e.refs.Load() == 0 {
					idle := time.Duration(now - e.lastUsed.Load())
					if idle >= p.ttl {
						delete(p.m, k)
					}
				}
			}
			p.mu.Unlock()
		}
	}
}

func (p *RWLockPool[K]) Acquire(key K) *Entry[K] {
	p.mu.Lock()
	defer p.mu.Unlock()
	e := p.m[key]
	if e == nil {
		e = &Entry[K]{
			key: key,
			p:   p,
		}
		e.touch()
		p.m[key] = e
	}
	e.refs.Add(1)
	return e
}

func (e *Entry[K]) Done() {
	if e == nil {
		return
	}
	if e.refs.Add(-1) == 0 {
		e.touch()
	}
}

func (e *Entry[K]) RLock() {
	e.rw.RLock()
}

func (e *Entry[K]) TryRLock() bool {
	return e.rw.TryRLock()
}

func (e *Entry[K]) RUnlock() {
	e.rw.RUnlock()
	e.touch()
}

func (e *Entry[K]) Lock() {
	e.rw.Lock()
}

func (e *Entry[K]) TryLock() bool {
	return e.rw.TryLock()
}

func (e *Entry[K]) Unlock() {
	e.rw.Unlock()
	e.touch()
}

// touch 在成功解锁或活跃时刷新 lastUsed，避免被误清理。
func (e *Entry[K]) touch() {
	e.lastUsed.Store(time.Now().UnixNano())
}

// ---- 读写操作封装 ----

func (p *RWLockPool[K]) WithRLock(ctx context.Context, key K, fn func(ctx context.Context, dest any) error, dest any) error {
	e := p.Acquire(key)
	defer e.Done()
	e.RLock()
	defer e.RUnlock()
	return fn(ctx, dest)
}

func (p *RWLockPool[K]) WithTryRLock(ctx context.Context, key K, fn func(ctx context.Context, dest any) error, dest any) (result bool, err error) {
	e := p.Acquire(key)
	defer e.Done()
	if !e.TryRLock() {
		return
	}
	defer e.RUnlock()
	result = true
	err = fn(ctx, dest)
	return
}

func (p *RWLockPool[K]) WithLock(ctx context.Context, key K, fn func(ctx context.Context, dest any) error, dest any) error {
	e := p.Acquire(key)
	defer e.Done()
	e.Lock()
	defer e.Unlock()
	return fn(ctx, dest)
}

func (p *RWLockPool[K]) WithTryLock(ctx context.Context, key K, fn func(ctx context.Context, dest any) error, dest any) (result bool, err error) {
	e := p.Acquire(key)
	defer e.Done()
	if !e.TryLock() {
		return
	}
	defer e.Unlock()
	result = true
	err = fn(ctx, dest)
	return
}
