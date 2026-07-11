package sses

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SessionManager(t *testing.T) {
	u1 := "u1"
	u2 := "u2"
	u3 := "u3"
	ch1 := "ch1"
	ch2 := "ch2"
	ch3 := "ch3"
	sm := NewSessionManager()
	require.Equal(t, 0, sm.SessionTotal())

	u1s1 := &Session{
		UserId:    u1,
		SessionId: NewSessionId(),
		Channel:   ch1,
		Message:   make(chan *Event, 1),
	}
	u1s2 := &Session{
		UserId:    u1,
		SessionId: NewSessionId(),
		Channel:   ch1,
		Message:   make(chan *Event, 1),
	}
	u1s3 := &Session{
		UserId:    u1,
		SessionId: NewSessionId(),
		Channel:   ch2,
		Message:   make(chan *Event, 1),
	}
	u1s4 := &Session{
		UserId:    u1,
		SessionId: NewSessionId(),
		Channel:   ch2,
		Message:   make(chan *Event, 1),
	}
	u2s1 := &Session{
		UserId:    u2,
		SessionId: NewSessionId(),
		Channel:   ch1,
		Message:   make(chan *Event, 1),
	}
	u2s2 := &Session{
		UserId:    u2,
		SessionId: NewSessionId(),
		Channel:   ch1,
		Message:   make(chan *Event, 1),
	}
	u2s3 := &Session{
		UserId:    u2,
		SessionId: NewSessionId(),
		Channel:   ch2,
		Message:   make(chan *Event, 1),
	}
	u2s4 := &Session{
		UserId:    u2,
		SessionId: NewSessionId(),
		Channel:   ch2,
		Message:   make(chan *Event, 1),
	}

	sm.Add(u1s1)
	sm.Add(u1s2)
	sm.Add(u1s3)
	sm.Add(u1s4)
	sm.Add(u2s1)
	sm.Add(u2s2)
	sm.Add(u2s3)
	sm.Add(u2s4)
	require.Equal(t, 8, sm.SessionTotal())

	//* collect
	require.Len(t, sm.Collect(ch1), 4)
	require.Len(t, sm.Collect(ch2), 4)
	require.Len(t, sm.Collect(ch3), 0)

	//* collect for user
	require.Len(t, sm.CollectForUser(ch1, u1), 2)
	require.Len(t, sm.CollectForUser(ch1, u2), 2)
	require.Len(t, sm.CollectForUser(ch1, u3), 0)

	//* collect for user session
	require.Len(t, sm.CollectForUserSession(ch1, u1, u1s1.SessionId), 1)
	require.Len(t, sm.CollectForUserSession(ch1, u3, u1s2.SessionId), 0)

	// delete u1
	sm.Delete(u1s1)
	require.Equal(t, 7, sm.SessionTotal())
	sm.DeleteByUserId(u1)
	require.Equal(t, 4, sm.SessionTotal())
	// dup delete
	sm.DeleteByUserId(u1)
	require.Equal(t, 4, sm.SessionTotal())

	// delete u2 one by one
	sm.Delete(u2s1)
	sm.Delete(u2s2)
	sm.Delete(u2s3)
	sm.Delete(u2s4)
	require.Equal(t, 0, sm.SessionTotal())
	require.Equal(t, 0, len(sm.byUserId))
	require.Equal(t, 0, len(sm.byChannel))

	u1s5 := &Session{
		UserId:    u1,
		SessionId: NewSessionId(),
		Channel:   ch1,
		Message:   make(chan *Event, 1),
	}
	sm.Add(u1s5)
	require.Equal(t, 1, sm.SessionTotal())
	sm.DeleteByUserId(u1)
	require.Equal(t, 0, sm.SessionTotal())
}

func Test_SessionManager_Add_DuplicateSession(t *testing.T) {
	u1 := "u1"
	ch1 := "ch1"
	t.Run("重复增加 - 删除旧的项", func(t *testing.T) {
		sm := NewSessionManager()
		require.Equal(t, 0, sm.SessionTotal())

		s1 := &Session{
			UserId:    u1,
			SessionId: NewSessionId(),
			Channel:   ch1,
			Message:   make(chan *Event, 1),
		}
		s2 := &Session{
			UserId:    u1,
			SessionId: NewSessionId(),
			Channel:   ch1,
			Message:   make(chan *Event, 1),
		}
		s3 := &Session{
			UserId:    u1,
			SessionId: s1.SessionId,
			Channel:   ch1,
			Message:   make(chan *Event, 1),
		}
		sm.Add(s1)
		sm.Add(s2)
		require.Equal(t, 2, sm.SessionTotal())
		sm.Add(s3)
		require.Equal(t, 2, sm.SessionTotal())
	})
	t.Run("重复增加 - 如果通道无会话, 删除通道", func(t *testing.T) {
		sm := NewSessionManager()
		require.Equal(t, 0, sm.SessionTotal())

		s1 := &Session{
			UserId:    u1,
			SessionId: NewSessionId(),
			Channel:   ch1,
			Message:   make(chan *Event, 1),
		}
		s2 := &Session{
			UserId:    u1,
			SessionId: s1.SessionId,
			Channel:   ch1,
			Message:   make(chan *Event, 1),
		}
		sm.Add(s1)
		require.Equal(t, 1, sm.SessionTotal())
		sm.Add(s2)
		require.Equal(t, 1, sm.SessionTotal())
	})
}

func Test_SessionManager_Delete_SessionNotFound(t *testing.T) {
	u1 := "u1"
	ch1 := "ch1"
	sm := NewSessionManager()
	require.Equal(t, 0, sm.SessionTotal())

	s1 := &Session{
		UserId:    u1,
		SessionId: NewSessionId(),
		Channel:   ch1,
		Message:   make(chan *Event, 1),
	}
	sm.Delete(s1)
	require.Equal(t, 0, sm.SessionTotal())
}
