package window_limiter

import (
	"math/rand/v2"
	"strconv"
	"time"
)

type SceneValuer interface {
	comparable
	Value() string
}

// UniqueId 生成一个唯一的id.
func UniqueId() string {
	var buf [20]byte

	b := strconv.AppendUint(buf[:0], uint64(time.Now().UnixNano()), 36)
	b = strconv.AppendUint(b, uint64(rand.Uint32()), 36)
	return string(b)
}
