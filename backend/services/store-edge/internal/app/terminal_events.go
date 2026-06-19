package app

import (
	"sync"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type TerminalEvent struct {
	Type            string                `json:"type"`
	TerminalID      string                `json:"terminalId"`
	StoreID         string                `json:"storeId"`
	Kind            domain.TerminalKind   `json:"kind"`
	Status          domain.TerminalStatus `json:"status"`
	SoftwareVersion string                `json:"softwareVersion,omitempty"`
	LastSeenAt      time.Time             `json:"lastSeenAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

type TerminalEventHub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan TerminalEvent]struct{}
}

func NewTerminalEventHub() *TerminalEventHub {
	return &TerminalEventHub{
		subscribers: map[string]map[chan TerminalEvent]struct{}{},
	}
}

func (h *TerminalEventHub) Subscribe(storeID string) chan TerminalEvent {
	ch := make(chan TerminalEvent, 8)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subscribers[storeID] == nil {
		h.subscribers[storeID] = map[chan TerminalEvent]struct{}{}
	}
	h.subscribers[storeID][ch] = struct{}{}
	return ch
}

func (h *TerminalEventHub) Unsubscribe(storeID string, ch chan TerminalEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subscribers, ok := h.subscribers[storeID]; ok {
		delete(subscribers, ch)
		if len(subscribers) == 0 {
			delete(h.subscribers, storeID)
		}
	}
	close(ch)
}

func (h *TerminalEventHub) PublishTerminalHeartbeat(terminal domain.Terminal) {
	event := TerminalEvent{
		Type:            "terminal_heartbeat",
		TerminalID:      terminal.ID,
		StoreID:         terminal.StoreID,
		Kind:            terminal.Kind,
		Status:          terminal.Status,
		SoftwareVersion: terminal.SoftwareVersion,
		LastSeenAt:      terminal.LastSeenAt,
		UpdatedAt:       terminal.UpdatedAt,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subscribers[terminal.StoreID] {
		select {
		case ch <- event:
		default:
		}
	}
}
