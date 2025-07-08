package swim

import (
	"fmt"
	serial "github.com/jscottransom/fringe/internal/proto"
	"time"
)

const maxDelivery = 3 // Tunable, but will keep low for now
type Entry struct {
	update        serial.MembershipUpdate
	Expiry        time.Time
	DeliveryCount int
	SeenPeers     map[string]bool
}

type PiggyBackQueue struct {
	Entries  []*Entry
	capacity int
}

// Add an entry to the PiggyBackQueue
// Push this entry to the front ("the left")
func (p *PiggyBackQueue) addEntry(entry *Entry) {

	// Append entry if there is capacity
	if p.len() < int(p.capacity) {
		// Check if entry is a duplicate - if it is, ignore
		for _, dupe := range p.Entries {
			if entry == dupe {
				return
			} else {
				p.Entries = append([]*Entry{entry}, p.Entries...)
			}
		}
	} else {
		// Remove the "first"
		p.removeFirst()
		p.Entries = append(p.Entries, entry)
		return
	}

}

// Execute a TTL round
func (p *PiggyBackQueue) evictEntry() {
	for idx, entry := range p.Entries {
		if time.Now().After(entry.Expiry) {
			p.removeElement(idx)
		}
	}
}

// Remove a specific Element from the queue
func (p *PiggyBackQueue) removeElement(idx int) {
	p.Entries = append(p.Entries[:idx], p.Entries[idx+1:]...)
}

// Check the length of the queue
func (p *PiggyBackQueue) len() int {

	return len(p.Entries)
}

// Remove the first item
func (p *PiggyBackQueue) removeFirst() {

	// Remove from the end
	p.Entries = p.Entries[1:]
}

// Prepare the piggyback queue to send over the network
func (p *PiggyBackQueue) GetEntries(nodeID string, max int) []*Entry {

	var entries []*Entry
	// Collect the max entries from the "back"
	// Increment the delivery count
	for i := 0; i < max; i++ {
		entry := p.Entries[i]

		_, seen := entry.SeenPeers[nodeID]
		// If this Node has not seen this update and the delivery is less than the max
		if !seen && entry.DeliveryCount < maxDelivery {
			entries = append(entries, p.Entries[i])
			entry.DeliveryCount++
		}

	}

	return entries

}
