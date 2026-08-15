package v9

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/thinkgos/proc-extra/limiter/token_rate"
)

func Test_TokenRate_Take(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	const (
		total = 100
		rate  = 5
		burst = 10
	)

	l := NewTokenRateStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	var allowed int
	for range total {
		time.Sleep(time.Second / time.Duration(total))
		b, err := l.AllowN(context.Background(), &token_rate.AllowNRequest{
			Key:   "tokenlimit",
			Rate:  rate,
			Burst: burst,
			Now:   time.Now(),
			N:     1,
		})
		if err == nil && b {
			allowed++
		}
	}

	assert.True(t, allowed >= burst+rate)
}

func Test_TokenRate_TakeBurst(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewTokenRateStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	var allowed int
	for range total {
		b, err := l.AllowN(context.Background(), &token_rate.AllowNRequest{
			Key:   "tokenlimit",
			Rate:  rate,
			Burst: burst,
			Now:   time.Now(),
			N:     1,
		})
		if err == nil && b {
			allowed++
		}
	}
	assert.True(t, allowed >= burst)
}
