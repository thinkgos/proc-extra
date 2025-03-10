package observer

import (
	"context"
	"fmt"
	"sync"

	"github.com/things-go/proc-extra/gpool"
	"github.com/things-go/proc/topic"
)

type Message interface {
	IntoPayload() ([]byte, error)
}

type Observer interface {
	Name() string
	Dispose(context.Context, Message) error
}

type AllowOptions struct {
	spawn bool
}

type AllowOption func(*AllowOptions)

func newAllowOption(opts ...AllowOption) AllowOptions {
	ao := AllowOptions{
		spawn: false,
	}
	for _, f := range opts {
		f(&ao)
	}
	return ao
}

func AllowSpawn() AllowOption {
	return func(a *AllowOptions) {
		a.spawn = true
	}
}

type ConcreteObserver struct {
	subs       *topic.Tree
	wg         sync.WaitGroup
	errHandler func(ctx context.Context, name string, err error)
}

func NewConcreteObserver() *ConcreteObserver {
	return &ConcreteObserver{
		subs:       topic.NewStandardTree(),
		errHandler: func(ctx context.Context, name string, err error) {},
	}
}
func (cc *ConcreteObserver) SetErrHandler(f func(ctx context.Context, name string, err error)) *ConcreteObserver {
	if f != nil {
		cc.errHandler = f
	}
	return cc
}

func (cc *ConcreteObserver) GetObserverNames(topic string) []string {
	values := cc.subs.Match(topic)
	names := make([]string, 0, len(values))
	for _, v := range values {
		names = append(names, v.(Observer).Name())
	}
	return names
}

func (cc *ConcreteObserver) Notify(ctx context.Context, topic string, m Message, opts ...AllowOption) error {
	ao := newAllowOption(opts...)
	values := cc.subs.Match(topic)
	for _, v := range values {
		ob := v.(Observer)
		if ao.spawn {
			cc.wg.Add(1)
			gpool.Go(func() {
				defer cc.wg.Done()
				err := ob.Dispose(ctx, m)
				if err != nil {
					cc.errHandler(ctx, ob.Name(), err)
				}
			})
		} else {
			err := ob.Dispose(ctx, m)
			if err != nil {
				return fmt.Errorf("observer: '%v' dispose failure, %w", ob.Name(), err)
			}
		}
	}
	return nil
}

func (cc *ConcreteObserver) AddObserver(topic string, ob Observer) error {
	cc.subs.Add(topic, ob)
	return nil
}

func (cc *ConcreteObserver) DeleteObserver(topic string, ob Observer) error {
	cc.subs.Remove(topic, ob)
	return nil
}

func (cc *ConcreteObserver) Close() error {
	cc.subs.Reset()
	cc.wg.Wait()
	return nil
}
