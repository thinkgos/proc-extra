package sses

import (
	"maps"
	"slices"
	"sync"
)

// Session
type Session struct {
	UserId    string // user id, 所属用户id
	SessionId string // session id, 会话id
	Channel   string // channel, 会话订阅的频道
	//* NOTE: 这个chan会被关闭, 在写时要注意处理panic
	Message chan *Event // message chan, 消息通道.
}

// SessionManager session manager
type SessionManager struct {
	locker    sync.RWMutex                     // 会话锁
	total     int                              // session total, 会话总数
	byUserId  map[string][]*Session            // userId -> []*Session, 用户的会话列表
	byChannel map[string]map[*Session]struct{} // channel -> map[*Session]struct{}// 订阅了channel的会话集合
}

// NewSessionManager 创建会话管理
func NewSessionManager() *SessionManager {
	return &SessionManager{
		locker:    sync.RWMutex{},
		total:     0,
		byUserId:  make(map[string][]*Session),
		byChannel: make(map[string]map[*Session]struct{}),
	}
}

// Total 获取会话总数
func (sm *SessionManager) SessionTotal() int {
	sm.locker.RLock()
	defer sm.locker.RUnlock()
	return sm.total
}

// Total 获取会话总数
func (sm *SessionManager) SessionTotalByChannel(channel string) int {
	sm.locker.RLock()
	defer sm.locker.RUnlock()
	return len(sm.byChannel[channel])
}

// Add 添加会话
func (sm *SessionManager) Add(ses *Session) {
	sm.locker.Lock()
	defer sm.locker.Unlock()
	sessions := sm.byUserId[ses.UserId]
	idx := slices.IndexFunc(sessions, func(v *Session) bool {
		return v.SessionId == ses.SessionId
	})
	//* 如果找到旧会话, 则关闭旧会话
	if idx >= 0 {
		sm.total--
		found := sessions[idx]
		// userId -> 删除对应sessionId的session
		sessions = slices.Delete(sessions, idx, idx+1)
		// channel -> 删除对应session
		cs, ok := sm.byChannel[found.Channel]
		if ok {
			if len(cs) == 1 {
				delete(sm.byChannel, found.Channel)
			} else {
				delete(cs, found)
			}
		}
		close(found.Message) // 关闭会话
	}
	//* 添加新的会话
	sm.total++
	// userId -> 增加新的session
	sessions = append(sessions, ses)
	sm.byUserId[ses.UserId] = sessions
	// channel -> 增加新的session
	cs, ok := sm.byChannel[ses.Channel]
	if !ok {
		cs = make(map[*Session]struct{})
		sm.byChannel[ses.Channel] = cs
	}
	cs[ses] = struct{}{}
}

// Delete 删除会话
func (sm *SessionManager) Delete(ses *Session) {
	sm.locker.Lock()
	defer sm.locker.Unlock()
	sessions := sm.byUserId[ses.UserId]
	idx := slices.IndexFunc(sessions, func(v *Session) bool {
		return v.SessionId == ses.SessionId
	})
	if idx >= 0 {
		//* 找到对应会话
		sm.total--
		found := sessions[idx]
		// userId ->  删除对应sessionId的session
		sessions = slices.Delete(sessions, idx, idx+1)
		if len(sessions) == 0 {
			delete(sm.byUserId, ses.UserId)
		} else {
			sm.byUserId[ses.UserId] = sessions
		}
		// channel -> 删除对应的session
		sp, ok := sm.byChannel[found.Channel]
		if ok {
			if len(sp) == 1 {
				delete(sm.byChannel, found.Channel)
			} else {
				delete(sp, found)
			}
		}
		close(found.Message) // 关闭会话
	} else {
		close(ses.Message) // 关闭会话
	}
}

// DeleteByUserId 删除用户的所有会话
func (sm *SessionManager) DeleteByUserId(userId string) {
	sm.locker.Lock()
	defer sm.locker.Unlock()
	sessions, ok := sm.byUserId[userId]
	if !ok {
		return
	}
	sm.total -= len(sessions)
	// userId -> 删除所有session
	delete(sm.byUserId, userId)
	// channel -> 删除对应的session
	for _, v := range sessions {
		sp, ok := sm.byChannel[v.Channel]
		if ok {
			if len(sp) == 1 {
				delete(sm.byChannel, v.Channel)
			} else {
				delete(sp, v)
			}
		}
		close(v.Message) // 关闭会话
	}
}

// Collect 获取channel的所有会话
func (sm *SessionManager) Collect(channel string) []*Session {
	sm.locker.RLock()
	defer sm.locker.RUnlock()
	sessions, ok := sm.byChannel[channel]
	if !ok {
		return []*Session{}
	}
	return slices.Collect(maps.Keys(sessions))
}

// CollectForUser 获取channel的指定用户的会话
func (sm *SessionManager) CollectForUser(channel, userId string) []*Session {
	sm.locker.RLock()
	defer sm.locker.RUnlock()
	sessions, ok := sm.byUserId[userId]
	if !ok {
		return []*Session{}
	}
	sess := make([]*Session, 0, len(sessions))
	for _, v := range sessions {
		if v.Channel == channel {
			sess = append(sess, v)
		}
	}
	return sess
}

// CollectForUserSession 获取指定channel的指定用户的指定会话
func (sm *SessionManager) CollectForUserSession(channel, userId, sessionId string) []*Session {
	sm.locker.RLock()
	defer sm.locker.RUnlock()
	sessions, ok := sm.byUserId[userId]
	if !ok {
		return []*Session{}
	}
	sess := make([]*Session, 0, len(sessions))
	for _, v := range sessions {
		if v.Channel == channel && v.SessionId == sessionId {
			sess = append(sess, v)
		}
	}
	return sess
}
