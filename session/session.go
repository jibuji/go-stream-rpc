package session

import (
	"context"
	"sync"
)

type Session interface {
	Get(key interface{}) interface{}
	Set(key, value interface{})
}

type MemSession struct {
	values sync.Map
}

func NewMemSession() *MemSession {
	return &MemSession{}
}

func (s *MemSession) Get(key interface{}) interface{} {
	value, _ := s.values.Load(key)
	return value
}

func (s *MemSession) Set(key, value interface{}) {
	s.values.Store(key, value)
}

type sessionKey int

const SessionContextKey sessionKey = 0

func From(ctx context.Context) Session {
	if session, ok := ctx.Value(SessionContextKey).(Session); ok {
		return session
	}
	return nil
}

func CreateDefaultSessionContext() context.Context {
	session := NewMemSession()
	return context.WithValue(context.Background(), SessionContextKey, session)
}
