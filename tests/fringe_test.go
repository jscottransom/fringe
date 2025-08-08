package tests

import (
	"testing"
	"time"

	serial "github.com/jscottransom/fringe/internal/proto"
	"github.com/jscottransom/fringe/internal/swim"
	"github.com/jscottransom/fringe/internal/sync"
)

func TestMerkleTreeBasic(t *testing.T) {
	// Create a new Merkle tree
	tree := sync.NewMerkleTree(4)

	// Add some test data
	err := tree.AddData("key1", []byte("value1"), 1)
	if err != nil {
		t.Fatalf("Failed to add data: %v", err)
	}

	err = tree.AddData("key2", []byte("value2"), 1)
	if err != nil {
		t.Fatalf("Failed to add data: %v", err)
	}

	// Verify tree hash is generated
	hash := tree.GetTreeHash()
	if hash == "" {
		t.Fatal("Tree hash should not be empty")
	}

	// Verify data can be retrieved
	data, err := tree.GetData("key1")
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}
	if string(data) != "value1" {
		t.Fatalf("Expected 'value1', got '%s'", string(data))
	}

	// Test stats
	stats := tree.GetStats()
	if stats["total_leaves"] != 2 {
		t.Fatalf("Expected 2 leaves, got %v", stats["total_leaves"])
	}
}

func TestMerkleTreeDiff(t *testing.T) {
	// Create two Merkle trees
	tree1 := sync.NewMerkleTree(4)
	tree2 := sync.NewMerkleTree(4)

	// Add same data to both trees
	tree1.AddData("key1", []byte("value1"), 1)
	tree1.AddData("key2", []byte("value2"), 1)

	tree2.AddData("key1", []byte("value1"), 1)
	tree2.AddData("key2", []byte("value2"), 1)

	// Should be no diff
	diff := tree1.GetDiff(tree2.GetTreeHash(), tree2.GetLeaves())
	if len(diff) != 0 {
		t.Fatalf("Expected no diff, got %d items", len(diff))
	}

	// Add different data to tree2
	tree2.AddData("key3", []byte("value3"), 1)

	// Should have diff
	diff = tree1.GetDiff(tree2.GetTreeHash(), tree2.GetLeaves())
	if len(diff) != 1 {
		t.Fatalf("Expected 1 diff item, got %d", len(diff))
	}

	// Apply diff to tree1
	err := tree1.ApplyDiff(diff)
	if err != nil {
		t.Fatalf("Failed to apply diff: %v", err)
	}

	// Trees should now be identical
	diff = tree1.GetDiff(tree2.GetTreeHash(), tree2.GetLeaves())
	if len(diff) != 0 {
		t.Fatalf("Expected no diff after sync, got %d items", len(diff))
	}
}

func TestSWIMNodeBasic(t *testing.T) {
	// Create a node table
	memberTable := &swim.NodeTable{
		Members: make(map[string]*swim.Peer),
	}

	// Create a peer
	peer := &swim.Peer{
		PeerID:           "test-peer",
		Address:          "127.0.0.1:8080",
		State:            swim.Alive,
		Incarnation:      1,
		SinceStateUpdate: time.Now(),
	}

	// Add peer to table
	memberTable.AddPeer("test-peer", peer)

	// Verify peer was added
	if len(memberTable.Members) != 1 {
		t.Fatalf("Expected 1 member, got %d", len(memberTable.Members))
	}

	// Test cluster size
	size := memberTable.GetClusterSize()
	if size != 1 {
		t.Fatalf("Expected cluster size 1, got %d", size)
	}

	// Test alive peers
	alivePeers := memberTable.GetAlivePeers()
	if len(alivePeers) != 1 {
		t.Fatalf("Expected 1 alive peer, got %d", len(alivePeers))
	}
}

func TestSWIMUpdatePeer(t *testing.T) {
	// Create a node table
	memberTable := &swim.NodeTable{
		Members: make(map[string]*swim.Peer),
	}

	// Create initial peer
	peer := &swim.Peer{
		PeerID:           "test-peer",
		Address:          "127.0.0.1:8080",
		State:            swim.Alive,
		Incarnation:      1,
		SinceStateUpdate: time.Now(),
	}

	memberTable.AddPeer("test-peer", peer)

	// Create update with higher incarnation
	update := &serial.MembershipUpdate{
		NodeId:      "test-peer",
		Address:     "127.0.0.1:8080",
		Incarnation: 2,
		State:       serial.State_SUSPECT,
	}

	// Update peer
	memberTable.UpdatePeer(update, false)

	// Verify state changed
	updatedPeer := memberTable.Members["test-peer"]
	if updatedPeer.State != swim.Suspected {
		t.Fatalf("Expected Suspected state, got %v", updatedPeer.State)
	}

	if updatedPeer.Incarnation != 2 {
		t.Fatalf("Expected incarnation 2, got %d", updatedPeer.Incarnation)
	}
}

func TestPiggyBackQueue(t *testing.T) {
	// Create queue
	queue := &swim.PiggyBackQueue{
		Entries:  make([]*swim.Entry, 0),
		Capacity: 10,
	}

	// Create entry
	entry := &swim.Entry{
		Update: &serial.MembershipUpdate{
			NodeId:      "test-peer",
			Address:     "127.0.0.1:8080",
			Incarnation: 1,
			State:       serial.State_ALIVE,
		},
		Expiry:        time.Now().Add(swim.PeerTTL),
		DeliveryCount: 0,
		SeenPeers:     make(map[string]bool),
	}

	// Add entry
	queue.AddEntry(entry)

	// Verify entry was added
	if len(queue.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(queue.Entries))
	}

	// Get entries - should return 1 entry since it hasn't been seen by test-node
	entries := queue.GetEntries("test-node", 5)
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry from GetEntries, got %d", len(entries))
	}

	// Verify delivery count increased
	if entries[0].DeliveryCount != 1 {
		t.Fatalf("Expected delivery count 1, got %d", entries[0].DeliveryCount)
	}

	// Get entries again from same node - should return 0 since it was already seen
	entries = queue.GetEntries("test-node", 5)
	if len(entries) != 0 {
		t.Fatalf("Expected 0 entries from second GetEntries, got %d", len(entries))
	}
}
