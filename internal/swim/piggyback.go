package swim

import (
	"time"

	serial "github.com/jscottransom/fringe/internal/proto"
)

const maxDelivery = 3

// Entry represents a membership update with delivery tracking
type Entry struct {
	Update        *serial.MembershipUpdate
	Expiry        time.Time
	DeliveryCount int
	SeenPeers     map[string]bool
}

// PiggyBackQueue manages membership updates for efficient network propagation
type PiggyBackQueue struct {
	Entries  []*Entry
	Capacity int
}

// Adds a new entry to the front of the queue with duplicate detection and capacity management
func (p *PiggyBackQueue) AddEntry(entry *Entry) {
	// Check for duplicates
	for _, dupe := range p.Entries {
		if entry == dupe {
			return
		}
	}

	// Add to front, remove oldest if at capacity
	if len(p.Entries) >= p.Capacity {
		p.Entries = p.Entries[:len(p.Entries)-1]
	}
	p.Entries = append([]*Entry{entry}, p.Entries...)
}

// Removes expired entries from the queue based on TTL expiration
func (p *PiggyBackQueue) EvictEntry() {
	now := time.Now()
	for i := len(p.Entries) - 1; i >= 0; i-- {
		if now.After(p.Entries[i].Expiry) {
			p.Entries = append(p.Entries[:i], p.Entries[i+1:]...)
		}
	}
}

// Returns entries that haven't been seen by the specified node and haven't exceeded delivery limit
func (p *PiggyBackQueue) GetEntries(nodeID string, max int) []*Entry {
	var entries []*Entry
	for i := 0; i < len(p.Entries) && i < max; i++ {
		entry := p.Entries[i]
		if !entry.SeenPeers[nodeID] && entry.DeliveryCount < maxDelivery {
			entries = append(entries, entry)
			entry.DeliveryCount++
			entry.SeenPeers[nodeID] = true
		}
	}
	return entries
}
