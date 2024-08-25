package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis/redistest"
	"testing"
	"time"
)

func init() {
	logx.Disable()
}

func TestTokenLimit_Take(t *testing.T) {
	store := redistest.CreateRedis(t)
	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := newTokenLimiter(rate, burst, store, "token limit")
	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if l.Allow() {
			allowed++
		}
	}
	assert.True(t, allowed >= burst+rate)
}
