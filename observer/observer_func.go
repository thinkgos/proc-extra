package observer

import (
	"context"
	"reflect"
	"runtime"
)

var _ Observer = (*ObserverFunc)(nil)

type ObserverFunc struct {
	dispose func(context.Context, any) error
}

func (ob *ObserverFunc) Name() string {
	fn := runtime.FuncForPC(reflect.ValueOf(ob.dispose).Pointer())
	if fn == nil {
		return "unknown function name"
	}
	return fn.Name()
}

func (ob *ObserverFunc) Dispose(ctx context.Context, m any) error {
	return ob.dispose(ctx, m)
}
