package token_limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/thinkgos/proc-extra/limiter/token_limiter"
	v9 "github.com/thinkgos/proc-extra/limiter/token_limiter/redis/v9"
)

func Test_TokenLimiter_Take(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	const (
		total = 100
		rate  = 5
		burst = 10
	)

	l := token_limiter.NewTokenLimiter(
		v9.NewTokenLimiterStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		"tokenlimit:",
		rate,
		burst,
	)
	var allowed int
	for range total {
		time.Sleep(time.Second / time.Duration(total))
		if l.Allow(context.Background(), "1") {
			allowed++
		}
	}

	assert.True(t, allowed >= burst+rate)
}

func Test_TokenLimiter_TakeBurst(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := token_limiter.NewTokenLimiter(
		v9.NewTokenLimiterStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
		"tokenlimit:",
		rate,
		burst,
	)
	var allowed int
	for range total {
		if l.Allow(context.Background(), "1") {
			allowed++
		}
	}
	assert.True(t, allowed >= burst)
}
