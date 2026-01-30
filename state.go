package main

import (
	"sync"
	"sync/atomic"
)

type counterStore struct {
	value int64
}

func (c *counterStore) inc() {
	atomic.AddInt64(&c.value, 1)
}

func (c *counterStore) load() int64 {
	return atomic.LoadInt64(&c.value)
}

type counterActionType string

const (
	actionIncrement counterActionType = "increment"
	actionDecrement counterActionType = "decrement"
	actionReset     counterActionType = "reset"
)

type counterAction struct {
	ClientID string
	Action   counterActionType
}

type counterActionStore struct {
	mu      sync.Mutex
	actions []counterAction
}

const maxActionHistory = 100

var (
	counter       counterStore
	counterEvents counterActionStore
)

func (s *counterActionStore) push(action counterAction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actions = append(s.actions, action)
	if len(s.actions) > maxActionHistory {
		s.actions = s.actions[len(s.actions)-maxActionHistory:]
	}
}

func (s *counterActionStore) load() []counterAction {
	s.mu.Lock()
	defer s.mu.Unlock()
	actions := make([]counterAction, len(s.actions))
	copy(actions, s.actions)
	return actions
}

func applyCounterAction(action counterActionType) {
	switch action {
	case actionIncrement:
		counter.inc()
	case actionDecrement:
		atomic.AddInt64(&counter.value, -1)
	case actionReset:
		atomic.StoreInt64(&counter.value, 0)
	}
}
