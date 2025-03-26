package observer

import (
	"context"
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
func (c *c1) Dispose(ctx context.Context, m any) error {
	log.Printf("dispose c1! notifier: %v\n", m.(*CallMessage).Name)
	return nil
}

type c2 struct{}

func (c *c2) Name() string { return "c2" }
func (c *c2) Dispose(ctx context.Context, m any) error {
	cm := m.(*CallMessage)
	if cm.Name == "c2err" {
		return errors.New("cause error for this notifier!")
	}
	log.Printf("dispose c2! notifier: %v\n", cm.Name)
	return nil
}

type c3 struct{}

func (c *c3) Name() string { return "c3" }
func (c *c3) Dispose(ctx context.Context, m any) error {
	log.Printf("dispose c3! notifier: %v\n", m.(*CallMessage).Name)
	return nil
}

type CallMessage struct {
	Name string `json:"name"`
}

func Test_Observer_CallChain(t *testing.T) {
	var err error

	topic := "/observer/call-chain"
	cc := NewConcreteObserver()
	defer cc.Close() // nolint: errcheck

	cc1 := &c1{}
	cc2 := &c2{}
	cc3 := &c3{}
	cc.AddObserver(topic, cc1)
	cc.AddObserver(topic, cc2)
	cc.AddObserver(topic, cc3)

	t.Run("match names", func(t *testing.T) {
		names := cc.GetObserverNames(topic)
		t.Logf("Observers for topic '%s': %v", topic, names)
	})

	t.Run("normal", func(t *testing.T) {
		err := cc.Notify(context.Background(),
			topic,
			&CallMessage{
				Name: "aaa",
			},
		)
		require.NoError(t, err)
	})

	t.Run("error stop", func(t *testing.T) {
		err := cc.Notify(context.Background(),
			topic,
			&CallMessage{
				Name: "c2err",
			},
		)
		require.Error(t, err)
	})

	t.Run("ignore error then continue", func(t *testing.T) {
		err := cc.Notify(context.Background(),
			topic,
			&CallMessage{
				Name: "c2err",
			},
			AllowIgnoreError(),
		)
		require.NoError(t, err)
	})

	t.Run("drop c2(which cause err)", func(t *testing.T) {
		cc.DeleteObserver(topic, cc2)
		err = cc.Notify(context.Background(),
			topic,
			&CallMessage{
				Name: "c2err",
			},
		)
		require.NoError(t, err)
	})
}

func Test_Observer_Spawn(t *testing.T) {
	topic := "/observer/spawn"
	cc := NewConcreteObserver().
		SetErrHandler(func(ctx context.Context, name string, err error) {
			t.Logf("Name: %s, err: %v\r\n", name, err)
		})
	defer cc.Close() // nolint: errcheck

	cc1 := &c1{}
	cc2 := &c2{}
	cc3 := &c3{}
	cc.AddObserver(topic, cc1)
	cc.AddObserver(topic, cc2)
	cc.AddObserver(topic, cc3)

	t.Run("normal", func(t *testing.T) {
		err := cc.Notify(context.Background(),
			topic,
			&CallMessage{
				Name: "aaa",
			},
			AllowSpawn(),
		)
		require.NoError(t, err)
	})

	t.Run("error on spawn", func(t *testing.T) {
		err := cc.Notify(context.Background(),
			topic,
			&CallMessage{
				Name: "c2err",
			},
			AllowSpawn(),
		)
		require.NoError(t, err)
	})
}

func Test_Observer_MustExistObserver(t *testing.T) {
	var err error

	topic := "/observer/must-exist-observer"
	cc := NewConcreteObserver()
	defer cc.Close()

	err = cc.Notify(context.Background(),
		topic,
		&CallMessage{
			Name: "aaa",
		},
		AllowMustExistObserver(),
	)
	require.Error(t, err)

	cc1 := &c1{}
	cc.AddObserver(topic, cc1)

	err = cc.Notify(context.Background(),
		topic,
		&CallMessage{
			Name: "bbb",
		},
		AllowMustExistObserver(),
	)
	require.NoError(t, err)
}
