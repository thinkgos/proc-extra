package window_limiter

import (
	"errors"
	"math/rand/v2"
	"strconv"
	"time"
)

var ErrParamKindNotFound = errors.New("sliding_window_limiter: the param's kind not found")

// UniqueId 生成一个唯一的id.
func UniqueId() string {
	var buf [20]byte

	b := strconv.AppendUint(buf[:0], uint64(time.Now().UnixNano()), 36)
	b = strconv.AppendUint(b, uint64(rand.Uint32()), 36)
	return string(b)
}
