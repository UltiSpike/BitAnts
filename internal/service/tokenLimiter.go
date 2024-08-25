package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	xrate "golang.org/x/time/rate"
)

const (
	tokenFormat     = "{%s}.tokens"
	timestampFormat = "{%s}.ts"
	pingInterval    = time.Millisecond * 100
)

var (
	tokenLuaScript string
	tokenScript    = redis.NewScript(tokenLuaScript)
)

type TokenLimiter struct {
	rate  int // 每秒产生速率
	burst int // 桶容量

	store          *redis.Redis
	tokenKey       string         // redis - key
	timestampKey   string         // 桶刷新时间 key
	rescueLock     sync.Mutex     // lock
	redisAlive     uint32         // redis 健康标识
	rescueLimiter  *xrate.Limiter // redis故障时采用进程内 令牌桶限流器
	monitorStarted bool           // redis检测探测任务标识
}

func newTokenLimiter(rate, burst int, store *redis.Redis, key string) *TokenLimiter {
	tokenKey := fmt.Sprintf(tokenFormat, key)
	timestampKey := fmt.Sprintf(timestampFormat, key)

	return &TokenLimiter{
		rate:          rate,
		burst:         burst,
		store:         store,
		tokenKey:      tokenKey,
		timestampKey:  timestampKey,
		redisAlive:    1,
		rescueLimiter: xrate.NewLimiter(xrate.Every((time.Second)/time.Duration(rate)), burst),
	}
}

func (lim *TokenLimiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

func (lim *TokenLimiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(context.Background(), now, n)
}

func (lim *TokenLimiter) AllowNCtx(ctx context.Context, now time.Time, n int) bool {
	return lim.reserveN(ctx, now, n)
}

func (lim *TokenLimiter) reserveN(ctx context.Context, now time.Time, n int) bool {
	if atomic.LoadUint32(&lim.redisAlive) == 0 {
		// 启用备用限流器
		return lim.rescueLimiter.AllowN(now, n)
	}
	resp, err := lim.store.ScriptRunCtx(ctx,
		tokenScript,
		[]string{
			lim.tokenKey,
			lim.timestampKey,
		},
		[]string{
			strconv.Itoa(lim.rate),
			strconv.Itoa(lim.burst),
			strconv.FormatInt(now.Unix(), 10),
			strconv.Itoa(n),
		},
	)

	if errors.Is(err, redis.Nil) {
		return false
	}
	if errorx.In(err, context.DeadlineExceeded, context.Canceled) {
		return false
	}

	if err != nil {
		logx.Errorf("fail to use rate limiter :%s", err)
		lim.startMonitor()
		return lim.rescueLimiter.AllowN(now, n)
	}

	code, ok := resp.(int64)
	if !ok {
		logx.Errorf("failed to eval redis script:%v", resp)
		lim.startMonitor()
		return lim.rescueLimiter.AllowN(now, n)
	}
	//	redis allowed == true
	//  Lua boolean true -> r integer reply with value of 1
	return code == 1

}

func (lim *TokenLimiter) startMonitor() {
	lim.rescueLock.Lock()
	defer lim.rescueLock.Unlock()
	if lim.monitorStarted {
		return
	}
	lim.monitorStarted = true
	atomic.StoreUint32(&lim.redisAlive, 0)
	go lim.waitForRedis()
}
func (lim *TokenLimiter) waitForRedis() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		lim.rescueLock.Lock()
		lim.monitorStarted = false
		lim.rescueLock.Unlock()
	}()

	for range ticker.C {
		if lim.store.Ping() {
			atomic.StoreUint32(&lim.redisAlive, 1)
			return
		}
	}
}
