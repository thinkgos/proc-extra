package sses

import (
	crand "crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var defaultEntropy = &ulid.LockedMonotonicReader{MonotonicReader: ulid.Monotonic(crand.Reader, 0)}

func NextId() string {
	return ulid.MustNew(uint64(time.Now().UTC().UnixMilli()), defaultEntropy).String()
}

func NewEventId() string { return NextId() }

func NewSessionId() string { return "sse-session-" + NextId() }
