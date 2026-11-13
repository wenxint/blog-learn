// Package ratelimit 提供「滑动窗口」限流。
//
// 为什么需要：评论接口是公开的、不需要登录的，任何人都能脚本刷。
// 一个 IP 一分钟发几千条，数据库和审核列表立刻被灌满。
//
// 为什么用滑动窗口而不是「固定窗口计数器」：
// 固定窗口（每分钟清零一次）在窗口边界会允许 2 倍突发——
// 00:59 打满一次，01:00 又打满一次，实际 2 秒内 2 倍流量。
// 滑动窗口记录每次命中的时间点，只看「最近 N 秒内有多少次」。
package ratelimit

import (
	"sync"
	"time"
)

// Limiter 按 key（这里用 IP）限制单位时间内的次数。
// 数据只在内存里，进程重启后清零；多实例部署时不共享
// （那时候应该换 Redis：key = ip，INCR + EXPIRE）。
type Limiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string][]time.Time
	lastGC time.Time
}

func New(limit int, window time.Duration) *Limiter {
	if limit <= 0 {
		limit = 3
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Limiter{
		limit:  limit,
		window: window,
		hits:   make(map[string][]time.Time),
		lastGC: time.Now(),
	}
}

// Allow 判断这次请求是否放行。
// 放行时记录一次命中；拒绝时返回还需要等多久。
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	// 顺手做一次过期清理，避免 map 无限增长
	//（只在每次调用时顺手清当前 key，全量 GC 见 gcLocked）
	l.gcLocked(now)

	// 只保留窗口内的记录
	list := l.hits[key]
	kept := list[:0]
	for _, t := range list {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= l.limit {
		// 被拒绝也要回写清理后的切片，否则旧记录永远不清
		l.hits[key] = kept
		// 最早的一次命中过期后就能再发，算出还要等多久
		retryAfter := kept[0].Add(l.window).Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter
	}

	l.hits[key] = append(kept, now)
	return true, 0
}

// gcLocked 每分钟全量清理一次过期 key，防止大量不同 IP 把内存撑爆。
func (l *Limiter) gcLocked(now time.Time) {
	if now.Sub(l.lastGC) < l.window {
		return
	}
	l.lastGC = now
	cutoff := now.Add(-l.window)
	for k, list := range l.hits {
		alive := 0
		for _, t := range list {
			if t.After(cutoff) {
				alive++
			}
		}
		if alive == 0 {
			delete(l.hits, k)
		}
	}
}

// Reset 清空所有记录（测试用）
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hits = make(map[string][]time.Time)
}

// Size 当前记录了多少个 key（测试用）
func (l *Limiter) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.hits)
}
