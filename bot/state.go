package main

import "sync"

type pendingAction string

const (
	actionNone   pendingAction = ""
	actionAdd    pendingAction = "add"
	actionSearch pendingAction = "search"
	actionDelete pendingAction = "delete"
)

type userState struct {
	mu      sync.Mutex
	pending map[int64]pendingAction
}

func newUserState() *userState {
	return &userState{pending: map[int64]pendingAction{}}
}

func (s *userState) set(userID int64, action pendingAction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if action == actionNone {
		delete(s.pending, userID)
		return
	}
	s.pending[userID] = action
}

func (s *userState) get(userID int64) pendingAction {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pending[userID]
}
