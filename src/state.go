package main

import (
	"sync"
	"time"
)

type counterActionType string

const (
	actionIncrement counterActionType = "increment"
	actionDecrement counterActionType = "decrement"
	actionReset     counterActionType = "reset"
)

type counterAction struct {
	ClientID string
	Action   counterActionType
	At       string
}

type counterActionStore struct {
	mu      sync.Mutex
	actions []counterAction
}

const maxActionHistory = 100

type appState struct {
	mu      sync.Mutex
	counter int64
	actions []counterAction
}

var state appState

func (s *appState) applyAction(action counterActionType, clientID string) int64 {
	// Apply a counter action and record it.
	s.mu.Lock()
	defer s.mu.Unlock()
	switch action {
	case actionIncrement:
		s.counter++
	case actionDecrement:
		s.counter--
	case actionReset:
		s.counter = 0
	}
	s.actions = append(s.actions, counterAction{
		ClientID: clientID,
		Action:   action,
		At:       time.Now().Format(time.RFC3339),
	})
	if len(s.actions) > maxActionHistory {
		s.actions = s.actions[len(s.actions)-maxActionHistory:]
	}
	return s.counter
}

func (s *appState) snapshot() (int64, []counterAction) {
	// Return a consistent snapshot of counter and actions.
	s.mu.Lock()
	defer s.mu.Unlock()
	actions := make([]counterAction, len(s.actions))
	copy(actions, s.actions)
	return s.counter, actions
}
