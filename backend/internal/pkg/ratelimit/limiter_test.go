package ratelimit

import (
	"testing"
	"time"
)

func TestAllowWithinLimit(t *testing.T) {
	l := New(3, time.Minute)

	for i := 1; i <= 3; i++ {
		ok, _ := l.Allow("1.1.1.1")
		if !ok {
			t.Fatalf("第 %d 次应当放行", i)
		}
	}
}

func TestBlockWhenExceeded(t *testing.T) {
	l := New(3, time.Minute)

	for i := 0; i < 3; i++ {
		l.Allow("1.1.1.1")
	}

	ok, retryAfter := l.Allow("1.1.1.1")
	if ok {
		t.Error("超过限制应当拒绝")
	}
	if retryAfter <= 0 || retryAfter > time.Minute {
		t.Errorf("重试等待时间不合理: %v", retryAfter)
	}
}

func TestDifferentKeysIndependent(t *testing.T) {
	l := New(1, time.Minute)

	if ok, _ := l.Allow("ip-a"); !ok {
		t.Fatal("ip-a 第一次应放行")
	}
	if ok, _ := l.Allow("ip-a"); ok {
		t.Error("ip-a 第二次应拒绝")
	}
	// 另一个 IP 不应受影响
	if ok, _ := l.Allow("ip-b"); !ok {
		t.Error("ip-b 不应被 ip-a 影响")
	}
}

func TestWindowExpires(t *testing.T) {
	l := New(1, 60*time.Millisecond)

	if ok, _ := l.Allow("k"); !ok {
		t.Fatal("第一次应放行")
	}
	if ok, _ := l.Allow("k"); ok {
		t.Fatal("窗口内第二次应拒绝")
	}

	// 等窗口过去
	time.Sleep(80 * time.Millisecond)

	if ok, _ := l.Allow("k"); !ok {
		t.Error("窗口过期后应重新放行")
	}
}

func TestReset(t *testing.T) {
	l := New(1, time.Minute)
	l.Allow("k")
	if ok, _ := l.Allow("k"); ok {
		t.Fatal("应被限制")
	}

	l.Reset()

	if ok, _ := l.Allow("k"); !ok {
		t.Error("Reset 后应放行")
	}
}
