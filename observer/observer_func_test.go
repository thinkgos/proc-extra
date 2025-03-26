package observer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ObserverFunc_CallChain(t *testing.T) {
	var err error

	topic := "/observer/call-chain"
	cc := NewConcreteObserver()
	defer cc.Close() // nolint: errcheck
	cc1 := &c1{}
	cc2 := &c2{}
	cc3 := &c3{}
	cc.AddObserverFunc(topic, cc1.Dispose).
		AddObserverFunc(topic, cc2.Dispose).
		AddObserverFunc(topic, cc3.Dispose)

	t.Run("match names", func(t *testing.T) {
		names := cc.GetObserverNames(topic)
		t.Logf("Observers for topic '%s': %v", topic, names)
	})

	t.Run("normal", func(t *testing.T) {
		err = cc.Notify(context.Background(), topic, &CallMessage{
			Name: "aaa",
		})
		require.NoError(t, err)
	})

	t.Run("error stop", func(t *testing.T) {
		err = cc.Notify(context.Background(), topic, &CallMessage{
			Name: "c2err",
		})
		require.Error(t, err)
	})
}
