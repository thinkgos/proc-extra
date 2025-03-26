package observer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ Observer = (*c1)(nil)
var _ Observer = (*c2)(nil)
var _ Observer = (*c3)(nil)

type c1 struct{}

func (c *c1) Name() string { return "c1" }
func (c *c1) Dispose(ctx context.Context, m Message) error {
	log.Printf("dispose c1! notifier: %v\n", m.(*CallMessage).Name)
	return nil
}

type c2 struct{}

func (c *c2) Name() string { return "c2" }
func (c *c2) Dispose(ctx context.Context, m Message) error {
	cm := m.(*CallMessage)
	if cm.Name == "c2nodo" {
		return errors.New("no do for this notifier!")
	}
	log.Printf("dispose c2! notifier: %v\n", cm.Name)
	return nil
}

type c3 struct{}

func (c *c3) Name() string { return "c3" }
func (c *c3) Dispose(ctx context.Context, m Message) error {
	log.Printf("dispose c3! notifier: %v\n", m.(*CallMessage).Name)
	return nil
}

type CallMessage struct {
	Name string `json:"name"`
}

func (c *CallMessage) IntoPayload() ([]byte, error) {
	return json.Marshal(c)
}

func Test_ObserverCallChain(t *testing.T) {
	var err error

	topic := "/observer/call-chain"
	cc := NewConcreteObserver()
	defer cc.Close()// nolint: errcheck
	cc1 := &c1{}
	cc2 := &c2{}
	cc3 := &c3{}
	err = cc.AddObserver(topic, cc1)
	require.NoError(t, err)
	err = cc.AddObserver(topic, cc2)
	require.NoError(t, err)
	err = cc.AddObserver(topic, cc3)
	require.NoError(t, err)

	t.Run("match names", func(t *testing.T) {
		names := cc.GetObserverNames(topic)
		t.Logf("Observers for topic '%s': %v", topic, names)
	})

	t.Run("normal", func(t *testing.T) {
		err := cc.Notify(context.Background(), topic, &CallMessage{
			Name: "aaa",
		})
		require.NoError(t, err)
	})

	t.Run("error stop", func(t *testing.T) {
		err := cc.Notify(context.Background(), topic, &CallMessage{
			Name: "c2nodo",
		})
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("drop c2(which can't do) then notify", func(t *testing.T) {
		err := cc.DeleteObserver(topic, cc2)
		require.NoError(t, err)
		err = cc.Notify(context.Background(), topic, &CallMessage{
			Name: "c2nodo",
		})
		require.NoError(t, err)
	})
}

func Test_ObserverSpawn(t *testing.T) {
	var err error

	topic := "/observer/spawn"
	cc := NewConcreteObserver()
	defer cc.Close()// nolint: errcheck
	cc1 := &c1{}
	cc2 := &c2{}
	cc3 := &c3{}
	err = cc.AddObserver(topic, cc1)
	require.NoError(t, err)
	err = cc.AddObserver(topic, cc2)
	require.NoError(t, err)
	err = cc.AddObserver(topic, cc3)
	require.NoError(t, err)

	t.Run("normal", func(t *testing.T) {
		err := cc.Notify(context.Background(), topic, &CallMessage{
			Name: "aaa",
		}, AllowSpawn())
		require.NoError(t, err)
	})

	t.Run("error stop", func(t *testing.T) {
		err := cc.Notify(context.Background(), topic, &CallMessage{
			Name: "c2nodo",
		}, AllowSpawn())
		require.NoError(t, err)
	})
}
