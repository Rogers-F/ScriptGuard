package services

import (
	"sync"
)

// ConcurrencyLimiter 并发限制器
type ConcurrencyLimiter struct {
	mu      sync.Mutex
	running int
	max     int
	cond    *sync.Cond
}

// NewConcurrencyLimiter 创建并发限制器
func NewConcurrencyLimiter(max int) *ConcurrencyLimiter {
	if max < 1 {
		max = 1
	}
	l := &ConcurrencyLimiter{max: max}
	l.cond = sync.NewCond(&l.mu)
	return l
}

// TryAcquire 尝试获取执行权限（非阻塞）
// 返回 true 表示获取成功，false 表示已达上限
func (l *ConcurrencyLimiter) TryAcquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.running >= l.max {
		return false
	}
	l.running++
	return true
}

// Acquire 获取执行权限（阻塞等待）
func (l *ConcurrencyLimiter) Acquire() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for l.running >= l.max {
		l.cond.Wait()
	}
	l.running++
}

// Release 释放执行权限
func (l *ConcurrencyLimiter) Release() {
	l.mu.Lock()
	l.running--
	l.mu.Unlock()
	l.cond.Signal()
}

// SetMax 动态设置最大并发数
func (l *ConcurrencyLimiter) SetMax(max int) {
	if max < 1 {
		max = 1
	}
	l.mu.Lock()
	l.max = max
	l.mu.Unlock()
	l.cond.Broadcast()
}

// GetRunning 获取当前运行数
func (l *ConcurrencyLimiter) GetRunning() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}

// GetMax 获取最大并发数
func (l *ConcurrencyLimiter) GetMax() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.max
}
