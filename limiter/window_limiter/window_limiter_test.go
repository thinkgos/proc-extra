package window_limiter

import (
	"testing"
)

func BenchmarkUniqueId(b *testing.B) {
	for b.Loop() {
		UniqueId()
	}
}
