package main

import (
	"sync"
	"sync/atomic"
)

type counterStore struct {
	value int64
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
	// Keep a bounded list of recent actions.
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actions = append(s.actions, action)
	if len(s.actions) > maxActionHistory {
		s.actions = s.actions[len(s.actions)-maxActionHistory:]
	}
}

func (s *counterActionStore) load() []counterAction {
	// Return a copy to avoid data races.
	s.mu.Lock()
	defer s.mu.Unlock()
	actions := make([]counterAction, len(s.actions))
	copy(actions, s.actions)
	return actions
}

func applyCounterAction(action counterActionType) {
	// Apply a counter action to in-memory state.
	switch action {
	case actionIncrement:
		atomic.AddInt64(&counter.value, 1)
	case actionDecrement:
		atomic.AddInt64(&counter.value, -1)
	case actionReset:
		atomic.StoreInt64(&counter.value, 0)
	}
}
