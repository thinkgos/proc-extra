package tracing

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var _ sdktrace.IDGenerator = (*customIdGenerator)(nil)

type customIdGenerator struct {
	entropy io.Reader
	sync.Mutex
	randSource *rand.Rand
}

func NewIdGenerator() sdktrace.IDGenerator {
	gen := &customIdGenerator{
		entropy: &ulid.LockedMonotonicReader{
			MonotonicReader: ulid.Monotonic(crand.Reader, 0),
		},
	}
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	gen.randSource = rand.New(rand.NewSource(rngSeed))
	return gen
}

// NewIDs implements trace.IDGenerator.
func (gen *customIdGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	tid := trace.TraceID(ulid.MustNew(uint64(time.Now().UTC().UnixMilli()), gen.entropy))
	gen.Lock()
	defer gen.Unlock()
	sid := gen.newSpanID()
	return tid, sid
}

// NewSpanID implements trace.IDGenerator.
func (gen *customIdGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	gen.Lock()
	defer gen.Unlock()
	return gen.newSpanID()
}

// NewSpanID implements trace.IDGenerator.
func (gen *customIdGenerator) newSpanID() trace.SpanID {
	sid := trace.SpanID{}
	for {
		_, _ = gen.randSource.Read(sid[:])
		if sid.IsValid() {
			break
		}
	}
	return sid
}
