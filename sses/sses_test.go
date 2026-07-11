package sses

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Broadcast(t *testing.T) {
	h := NewHub(
		WithBufferSize(1000),
		WithHeartbeat(time.Second*30),
		WithRetryLimit(3),
		WithRetryTimeout(time.Second*3),
		WithStore(newTestMemoryStore()),
	)
	defer h.Close() // nolint: errcheck
	stats := h.Stats()

	ch1 := "default1"
	ch2 := "default2"
	ch3 := "default3"
	u1 := "u1"
	u2 := "U2"
	s1 := &Session{
		Channel:   ch1,
		UserId:    u1,
		SessionId: NewSessionId(),
		Message:   make(chan *Event, 1),
	}
	s2 := &Session{
		Channel:   ch1,
		UserId:    u2,
		SessionId: NewSessionId(),
		Message:   make(chan *Event, 1),
	}
	s3 := &Session{
		Channel:   ch2,
		UserId:    u1,
		SessionId: NewSessionId(),
		Message:   make(chan *Event, 1),
	}
	h.sessions.Add(s1)
	h.sessions.Add(s2)
	h.sessions.Add(s3)
	require.Equal(t, 3, h.SessionTotal())

	// test push event
	event := &Event{
		Id:    "e1",
		Event: "test",
		Data:  "data",
	}
	err := h.Broadcast(context.Background(), ch1, event)
	require.NoError(t, err)

	wantReceivedMsg := 2
	gotReceivedMsg := 0
	for {
		select {
		case e := <-s1.Message: // s1 订阅了ch1, 所以会收到e1
			gotReceivedMsg++
			require.Equal(t, "e1", e.Id)
			require.Equal(t, "test", e.Event)
			require.Equal(t, "data", e.Data)
		case e := <-s2.Message: // s2 订阅了ch1, 所以会收到e1
			gotReceivedMsg++
			require.Equal(t, "e1", e.Id)
			require.Equal(t, "test", e.Event)
			require.Equal(t, "data", e.Data)
		case e := <-s3.Message: // s3 订阅了ch2, 所以不会收到e1
			t.Errorf("unexpected event publish to session, ch: %s, uid: %s, e: %v", ch2, u1, e)
		case <-time.After(time.Millisecond * 100):
			t.Error("expected event publish to session but got timeout")
		}
		if gotReceivedMsg == wantReceivedMsg {
			break
		}
	}
	require.Equal(t, int64(2), stats.ReqSuccess.Load())
	require.Equal(t, int64(0), stats.ReqFailure.Load())
	require.Equal(t, int64(0), stats.ReqTimeout.Load())

	// channel ch3 no session, ignore
	err = h.Broadcast(context.Background(), ch3, event)
	require.NoError(t, err)
	require.Equal(t, int64(2), stats.ReqSuccess.Load())
	require.Equal(t, int64(0), stats.ReqFailure.Load())
	require.Equal(t, int64(0), stats.ReqTimeout.Load())

	h.sessions.Delete(s1)
	h.sessions.Delete(s2)
	require.Equal(t, 1, h.SessionTotal())
	// channel ch1 session cleanup, no session, ignore
	err = h.Broadcast(context.Background(), ch1, event)
	require.NoError(t, err)

	require.Equal(t, int64(2), stats.ReqSuccess.Load())
	require.Equal(t, int64(0), stats.ReqFailure.Load())
	require.Equal(t, int64(0), stats.ReqTimeout.Load())
	require.Equal(t, int64(0), stats.SendSuccess.Load())
	require.Equal(t, int64(0), stats.SendFailure.Load())
}

func Test_Publish(t *testing.T) {
	h := NewHub()
	defer h.Close() // nolint: errcheck
	stats := h.Stats()

	ch1 := "default1"
	ch2 := "default2"
	u1 := "u1"
	u2 := "u2"
	s1 := &Session{
		Channel:   ch1,
		UserId:    u1,
		SessionId: NewSessionId(),
		Message:   make(chan *Event, 1),
	}
	s2 := &Session{
		Channel:   ch1,
		UserId:    u2,
		SessionId: NewSessionId(),
		Message:   make(chan *Event, 1),
	}
	s3 := &Session{
		Channel:   ch2,
		UserId:    u1,
		SessionId: NewSessionId(),
		Message:   make(chan *Event, 1),
	}
	h.sessions.Add(s1)
	h.sessions.Add(s2)
	h.sessions.Add(s3)
	require.Equal(t, 3, h.SessionTotal())

	// test push event
	event := &Event{
		Id:    "e1",
		Event: "test",
		Data:  "data",
	}
	err := h.PublishSync(context.Background(), ch1, u1, event)
	require.NoError(t, err)
	wantReceivedMsg := 1
	gotReceivedMsg := 0
	for {
		select {
		case e := <-s1.Message: // s1 u1订阅了ch1, 所以会收到e1
			gotReceivedMsg++
			require.Equal(t, "e1", e.Id)
			require.Equal(t, "test", e.Event)
			require.Equal(t, "data", e.Data)
		case e := <-s2.Message: // s2 u2订阅了ch1, 所以不会收到e1
			t.Errorf("unexpected event publish to session, ch: %s, uid: %s, e: %v", ch1, u2, e)
		case e := <-s3.Message: // s3 u1订阅了ch2, 所以不会收到e1
			t.Errorf("unexpected event publish to session, ch: %s, uid: %s, e: %v", ch2, u1, e)
		case <-time.After(time.Millisecond * 100):
			t.Error("expected event publish to session but got timeout")
		}
		if gotReceivedMsg == wantReceivedMsg {
			break
		}
	}
	require.Equal(t, int64(1), stats.ReqSuccess.Load())
	require.Equal(t, int64(0), stats.ReqFailure.Load())
	require.Equal(t, int64(0), stats.ReqTimeout.Load())

	// not found any session, ignore
	err = h.Publish(context.Background(), ch2, u2, event)
	require.NoError(t, err)

	require.Equal(t, int64(1), stats.ReqSuccess.Load())
	require.Equal(t, int64(0), stats.ReqFailure.Load())
	require.Equal(t, int64(0), stats.ReqTimeout.Load())
}

func Test_PublishSession(t *testing.T) {
	h := NewHub()
	defer h.Close() // nolint: errcheck
	stats := h.Stats()

	ch1 := "default"
	u1 := "u1"
	u2 := "u2"
	sid1 := NewSessionId()
	sid2 := NewSessionId()
	s1 := &Session{
		Channel:   ch1,
		UserId:    u1,
		SessionId: sid1,
		Message:   make(chan *Event, 1),
	}
	s2 := &Session{
		Channel:   ch1,
		UserId:    u1,
		SessionId: sid2,
		Message:   make(chan *Event, 1),
	}
	h.sessions.Add(s1)
	h.sessions.Add(s2)
	require.Equal(t, 2, h.SessionTotal())

	// test push event
	event := &Event{
		Id:    "e1",
		Event: "test",
		Data:  "data",
	}
	err := h.PublishSession(context.Background(), ch1, u1, sid1, event)
	require.NoError(t, err)
	wantReceivedMsg := 1
	gotReceivedMsg := 0
	for {
		select {
		case e := <-s1.Message: // s1 u1-sid1订阅了ch1, 所以会收到e1
			gotReceivedMsg++
			require.Equal(t, "e1", e.Id)
			require.Equal(t, "test", e.Event)
			require.Equal(t, "data", e.Data)
		case e := <-s2.Message: // s2 u1-sid2订阅了ch1, 所以不会收到e1
			t.Errorf("unexpected event publish to session, ch: %s, uid: %s, e: %v", ch1, u1, e)
		case <-time.After(time.Millisecond * 100):
			t.Error("expected event broadcast to session but got timeout")
		}
		if gotReceivedMsg == wantReceivedMsg {
			break
		}
	}
	require.Equal(t, int64(1), stats.ReqSuccess.Load())
	require.Equal(t, int64(0), stats.ReqFailure.Load())
	require.Equal(t, int64(0), stats.ReqTimeout.Load())

	// not found any session, ignore
	err = h.PublishSessionSync(context.Background(), ch1, u2, sid2, event)
	require.NoError(t, err)

	require.Equal(t, int64(1), stats.ReqSuccess.Load())
	require.Equal(t, int64(0), stats.ReqFailure.Load())
	require.Equal(t, int64(0), stats.ReqTimeout.Load())
}

func Test_Publish_Timeout(t *testing.T) {
	h := NewHub(WithRetryTimeout(time.Millisecond * 100))
	defer h.Close() // nolint: errcheck
	stats := h.Stats()

	ch := "default"
	u1 := "u1"
	session := &Session{
		Channel:   ch,
		UserId:    u1,
		SessionId: NewSessionId(),
		Message:   make(chan *Event, 1),
	}
	h.sessions.Add(session)
	require.Equal(t, 1, h.SessionTotal())

	// test push event
	event := &Event{
		Id:    "e1",
		Event: "test",
		Data:  "data",
	}
	err := h.PublishSync(context.Background(), ch, u1, event)
	require.NoError(t, err)
	err = h.Publish(context.Background(), ch, u1, event)
	require.NoError(t, err)
	err = h.PublishSync(context.Background(), ch, u1, event)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	select {
	case e := <-session.Message:
		require.Equal(t, "e1", e.Id)
		require.Equal(t, "test", e.Event)
		require.Equal(t, "data", e.Data)
	case <-time.After(time.Millisecond * 100):
		t.Error("expected event broadcast to session but got timeout")
	}

	require.Equal(t, int64(1), stats.ReqSuccess.Load())
	require.Equal(t, int64(2), stats.ReqFailure.Load())
	require.Equal(t, int64(2), stats.ReqTimeout.Load())
}

type testMemoryStore struct {
	mu     sync.RWMutex
	events map[string]map[string][]*Event
}

func newTestMemoryStore() Store {
	return &testMemoryStore{
		events: make(map[string]map[string][]*Event),
	}
}

func (m *testMemoryStore) Save(_ context.Context, channel string, e *Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.events[channel]; !ok {
		m.events[channel] = make(map[string][]*Event)
	}
	m.events[channel][e.Event] = append(m.events[channel][e.Event], e)
	return nil
}

func (m *testMemoryStore) ListByLastId(_ context.Context, channel, eventType, lastEventId string, pageSize int) ([]*Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	byChannel, ok := m.events[channel]
	if !ok {
		return nil, nil
	}
	events, ok := byChannel[eventType]
	if !ok {
		return nil, nil
	}
	// find the starting position
	start := 0
	for i, e := range events {
		if e.Id > lastEventId {
			start = i
			break
		}
	}
	return events[start:min(start+pageSize, len(events))], nil
}
