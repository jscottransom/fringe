package swim

import (
	"fmt"
	"net"
	"sync"
	"time"
	"github.com/google/uuid"
	serial "github.com/jscottransom/fringe/internal/proto"
)


const PeerTTL = 60 * time.Second
// Construct for a custom enum representing the Node Status
type NodeState int

const (
	Alive NodeState = iota
	Suspected
	Dead
	Left
)

// Membership level data that we exchange across rounds and throughout the network
type Peer struct {
	PeerID           string
	Address          string
	State            NodeState
	Incarnation      uint64
	SinceStateUpdate time.Time
}

// Setup the basic structure of a Node
// Nodes contain a UUID represent the process that should survive restarts and failures (stable)
// Peer Table contains a mapping of these identifiers to Node Objects
type Node struct {
	NodeId      string
	MemberTable *NodeTable
    Queue		*PiggyBackQueue
}

// Table mapping UUID to Node objects
type NodeTable struct {
	Members map[string]*Peer
	mu      sync.Mutex
}

// Add a node
func (n *NodeTable) addPeer(nodeID string, peer *Peer) {
	// Safe access, prevent multiple current updates
	n.mu.Lock()
	defer n.mu.Unlock()

	// Add an entry to the NodeTable map
	n.Members[nodeID] = peer

	// Write success
	fmt.Println("Successfully added %s to Node Table", nodeID)
}

// Update a Peer in the NodeTable
func (n *NodeTable) updatePeer(update serial.MembershipUpdate) {
	// Safe access, prevent multiple current updates
	n.mu.Lock()
	defer n.mu.Unlock()

	// For a given ID, check if the incarnation is less than the update.
	// If it is, then the peer table incarnation syncs with the update
	peer := n.Members[update.NodeId]

	switch {
	case peer.Incarnation > update.Incarnation:
		return

	case peer.Incarnation < update.Incarnation:
		// Set the incarnation
		peer.Incarnation = update.Incarnation

		// Take the state from the update
		// Construct a NodeState Enum
		peer.State = NodeState(*update.State.Enum())

	case peer.Incarnation == update.Incarnation:
		// Take the state with the higher State Enum value
		// This corresponds to the "Least Alive" principle of Node status
		peer.State = max(peer.State, NodeState(*update.State.Enum()))

	}

	// Remove Suspect / Dead Nodes by checking the time since 
	for _, state := range []NodeState{Suspected, Dead, Left} {
		if (peer.State == state) && (time.Since(peer.SinceStateUpdate) > PeerTTL) {
			delete(n.Members, update.NodeId)
		}
	}

	
}

