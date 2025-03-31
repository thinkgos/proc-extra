package observer

import (
	"context"
	"fmt"
	"sync"

	"github.com/things-go/proc/topic"
	"github.com/thinkgos/proc-extra/gpool"
)

type Observer interface {
	Name() string
	Dispose(context.Context, any) error
}

type AllowOptions struct {
	spawn             bool // 使用协程并发, 默认不使用
	ignoreError       bool // 允许忽略错误, 仅调用链有效, 默认不忽略
	mustExistObserver bool // 必须存在观察者, 默认可以没有观察者
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

// AllowSpawn 允许使用协程并发, 默认不使用
func AllowSpawn() AllowOption {
	return func(a *AllowOptions) {
		a.spawn = true
	}
}

// AllowIgnoreError 允许忽略错误, 仅调用链有效, 协程并发不生效, 默认不忽略
func AllowIgnoreError() AllowOption {
	return func(ao *AllowOptions) {
		ao.ignoreError = true
	}
}

// AllowMustExistObserver 必须存在观察者, 默认可以没有观察者
func AllowMustExistObserver() AllowOption {
	return func(ao *AllowOptions) {
		ao.mustExistObserver = true
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

func (cc *ConcreteObserver) Notify(ctx context.Context, topic string, m any, opts ...AllowOption) error {
	ao := newAllowOption(opts...)
	values := cc.subs.Match(topic)
	if ao.mustExistObserver && len(values) == 0 {
		return fmt.Errorf("observer: must exist observer for topic '%s'", topic)
	}
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
			if err != nil && !ao.ignoreError {
				return fmt.Errorf("observer: '%v' dispose failure, %w", ob.Name(), err)
			}
		}
	}
	return nil
}

func (cc *ConcreteObserver) AddObserver(topic string, ob Observer) {
	cc.subs.Add(topic, ob)
}

func (cc *ConcreteObserver) DeleteObserver(topic string, ob Observer) {
	cc.subs.Remove(topic, ob)
}

func (cc *ConcreteObserver) AddObserverFunc(topic string, dispose func(context.Context, any) error) *ConcreteObserver {
	cc.subs.Add(topic, &ObserverFunc{dispose: dispose})
	return cc
}

func (cc *ConcreteObserver) Close() error {
	cc.subs.Reset()
	cc.wg.Wait()
	return nil
}
