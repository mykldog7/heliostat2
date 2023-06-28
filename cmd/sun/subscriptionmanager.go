package main

import (
	"sync"
)

// SubManager manages 'subscripttions' to keep track of all the open connection's and pushes 'outbound' messages to them
type SubManager struct {
	mu      sync.Mutex
	subs    map[chan []byte]bool
	publish <-chan []byte
}

// NewSubManager returns an unlocked, empty list of channels to push messages to
func NewSubManger(publish <-chan []byte) *SubManager {
	sm := SubManager{}
	sm.subs = make(map[chan []byte]bool)
	sm.publish = publish
	return &sm
}
func (sm *SubManager) AddSub(c chan []byte) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.subs[c] = true
}
func (sm *SubManager) RemoveSub(c chan []byte) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.subs, c)
}

// Start waits for a msg on the publish channel and then writes the message to all registered 'out' channels, sending the same message to every ws connection
func (sm *SubManager) Start() error {
	for {
		msg := <-sm.publish
		sm.mu.Lock()
		for k := range sm.subs {
			k <- msg
		}
		sm.mu.Unlock()
	}
}
