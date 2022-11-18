package server

import (
	"sync"
	"sync/atomic"
	"time"
)

type SessionId int
type SessionToken []byte
type FutureUnixTime int64

type Sessions struct {
	sessionTicker    *time.Ticker
	sessionTokens    map[SessionId]SessionToken // session id => secret
	sessionTokensMux sync.RWMutex
	expireSlots      map[FutureUnixTime][]SessionId // unix timestamp in future -> secrets
	expireSlotsMux   sync.RWMutex
	expiredSession   uint64
}

func (s *Sessions) ExpiredSession() uint64 {
	return atomic.LoadUint64(&s.expiredSession)
}

func (instance *Instance) getTokenFromSessionId(sessionId SessionId) SessionToken {
	instance.sessionTokensMux.RLock()
	token := instance.sessionTokens[sessionId]
	instance.sessionTokensMux.RUnlock()
	return token
}

func (instance *Instance) registerSessionToken(sessionId SessionId, token SessionToken) {
	instance.sessionTokensMux.RLock()
	_, existing := instance.sessionTokens[sessionId]
	instance.sessionTokensMux.RUnlock()
	if existing {
		return
	}

	instance.sessionTokensMux.Lock()
	instance.sessionTokens[sessionId] = token
	instance.sessionTokensMux.Unlock()

	// register for future time in expire via ticker
	nowUnix := time.Now().Unix()
	expireTs := FutureUnixTime(nowUnix + int64(ConnectionTimeout.Seconds()) + 1)
	instance.Sessions.expireSlotsMux.Lock()
	if instance.Sessions.expireSlots[expireTs] == nil {
		instance.Sessions.expireSlots[expireTs] = make([]SessionId, 0)
	}
	instance.Sessions.expireSlots[expireTs] = append(instance.Sessions.expireSlots[expireTs], sessionId)
	instance.Sessions.expireSlotsMux.Unlock()
}

func (instance *Instance) sessionExpire() int {
	nowUnix := time.Now().Unix()

	// scan for expired slots
	expiredSlots := make([]FutureUnixTime, 0)
	instance.Sessions.expireSlotsMux.RLock()
	for ts := range instance.Sessions.expireSlots {
		if ts >= FutureUnixTime(nowUnix) {
			// not yet expired
			continue
		}
		expiredSlots = append(expiredSlots, ts)
	}
	instance.Sessions.expireSlotsMux.RUnlock()

	if len(expiredSlots) < 1 {
		// nothing to expire
		return 0
	}

	// remove all expired tokens
	numDeleted := 0
	instance.Sessions.sessionTokensMux.Lock()
	for _, expiredSlot := range expiredSlots {
		sessionIds, found := instance.Sessions.expireSlots[expiredSlot]
		if !found {
			continue
		}
		for _, sessionId := range sessionIds {
			delete(instance.Sessions.sessionTokens, sessionId)
			numDeleted++
		}
	}
	instance.Sessions.sessionTokensMux.Unlock()

	instance.Sessions.expireSlotsMux.Lock()
	for _, expiredSlot := range expiredSlots {
		delete(instance.Sessions.expireSlots, expiredSlot)
	}
	instance.Sessions.expireSlotsMux.Unlock()

	atomic.AddUint64(&instance.Sessions.expiredSession, uint64(numDeleted))

	return numDeleted
}

func NewSessions() *Sessions {
	return &Sessions{
		sessionTicker: nil,
		sessionTokens: make(map[SessionId]SessionToken),
		expireSlots:   make(map[FutureUnixTime][]SessionId),
	}
}
