package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// MerkleNode represents a node in the Merkle tree with hash and data
type MerkleNode struct {
	Hash     string
	Left     *MerkleNode
	Right    *MerkleNode
	Data     []byte
	IsLeaf   bool
	Key      string
	Modified time.Time
}

// MerkleTree provides efficient data synchronization with hash-based diff detection
type MerkleTree struct {
	Root     *MerkleNode
	Leaves   map[string]*MerkleNode
	mu       sync.RWMutex
	MaxDepth int
}

// DataItem represents a piece of data in the system with versioning
type DataItem struct {
	Key      string
	Value    []byte
	Modified time.Time
	Version  uint64
}

// SyncRequest represents a synchronization request between nodes
type SyncRequest struct {
	RequestorID string
	TreeHash    string
	Timestamp   time.Time
	Depth       int
}

// SyncResponse represents a synchronization response with diff data
type SyncResponse struct {
	ResponderID string
	TreeHash    string
	Diff        []DataItem
	Timestamp   time.Time
}

// Creates a new Merkle tree with specified maximum depth for efficient synchronization
func NewMerkleTree(maxDepth int) *MerkleTree {
	return &MerkleTree{
		Leaves:   make(map[string]*MerkleNode),
		MaxDepth: maxDepth,
	}
}

// Adds data to the Merkle tree and rebuilds the tree structure with hash computation
func (mt *MerkleTree) AddData(key string, value []byte, version uint64) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	leaf := &MerkleNode{
		Hash:     hashData(value),
		Data:     value,
		IsLeaf:   true,
		Key:      key,
		Modified: time.Now(),
	}

	mt.Leaves[key] = leaf
	mt.rebuildTree()
	return nil
}

// Updates existing data in the Merkle tree and rebuilds the tree with new hash
func (mt *MerkleTree) UpdateData(key string, value []byte, version uint64) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if leaf, exists := mt.Leaves[key]; exists {
		leaf.Hash = hashData(value)
		leaf.Data = value
		leaf.Modified = time.Now()
		mt.rebuildTree()
		return nil
	}

	return fmt.Errorf("key %s not found", key)
}

// Removes data from the Merkle tree and rebuilds the tree structure
func (mt *MerkleTree) DeleteData(key string) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if _, exists := mt.Leaves[key]; exists {
		delete(mt.Leaves, key)
		mt.rebuildTree()
		return nil
	}

	return fmt.Errorf("key %s not found", key)
}

// Retrieves data from the Merkle tree by key with read-safe access
func (mt *MerkleTree) GetData(key string) ([]byte, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if leaf, exists := mt.Leaves[key]; exists {
		return leaf.Data, nil
	}

	return nil, fmt.Errorf("key %s not found", key)
}

// Returns the root hash of the Merkle tree for quick comparison and synchronization
func (mt *MerkleTree) GetTreeHash() string {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if mt.Root == nil {
		return ""
	}
	return mt.Root.Hash
}

// Rebuilds the Merkle tree from leaves with consistent ordering for deterministic hashing
func (mt *MerkleTree) rebuildTree() {
	if len(mt.Leaves) == 0 {
		mt.Root = nil
		return
	}

	leaves := make([]*MerkleNode, 0, len(mt.Leaves))
	for _, leaf := range mt.Leaves {
		leaves = append(leaves, leaf)
	}
	sort.Slice(leaves, func(i, j int) bool {
		return leaves[i].Key < leaves[j].Key
	})

	mt.Root = mt.buildTreeLevel(leaves, 0)
}

// Recursively builds tree levels with configurable depth limit for efficient storage
func (mt *MerkleTree) buildTreeLevel(nodes []*MerkleNode, depth int) *MerkleNode {
	if len(nodes) == 1 {
		return nodes[0]
	}

	if depth >= mt.MaxDepth {
		combinedHash := ""
		for _, node := range nodes {
			combinedHash += node.Hash
		}
		return &MerkleNode{
			Hash:   hashData([]byte(combinedHash)),
			IsLeaf: false,
		}
	}

	var pairs []*MerkleNode
	for i := 0; i < len(nodes); i += 2 {
		if i+1 < len(nodes) {
			parentHash := hashData([]byte(nodes[i].Hash + nodes[i+1].Hash))
			pairs = append(pairs, &MerkleNode{
				Hash:   parentHash,
				Left:   nodes[i],
				Right:  nodes[i+1],
				IsLeaf: false,
			})
		} else {
			pairs = append(pairs, nodes[i])
		}
	}

	return mt.buildTreeLevel(pairs, depth+1)
}

// Returns the difference between two Merkle trees for efficient synchronization with minimal data transfer
func (mt *MerkleTree) GetDiff(otherHash string, otherLeaves map[string]*MerkleNode) []DataItem {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if mt.GetTreeHash() == otherHash {
		return nil
	}

	var diff []DataItem

	// Check for changes and additions
	for key, leaf := range mt.Leaves {
		if otherLeaf, exists := otherLeaves[key]; !exists || leaf.Hash != otherLeaf.Hash {
			diff = append(diff, DataItem{
				Key:      key,
				Value:    leaf.Data,
				Modified: leaf.Modified,
				Version:  0,
			})
		}
	}

	// Check for deletions
	for key, otherLeaf := range otherLeaves {
		if _, exists := mt.Leaves[key]; !exists {
			diff = append(diff, DataItem{
				Key:      key,
				Value:    otherLeaf.Data,
				Modified: otherLeaf.Modified,
				Version:  0,
			})
		}
	}

	return diff
}

// Applies a diff to the Merkle tree and rebuilds the structure for consistency
func (mt *MerkleTree) ApplyDiff(diff []DataItem) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	for _, item := range diff {
		if item.Value == nil {
			delete(mt.Leaves, item.Key)
		} else {
			mt.Leaves[item.Key] = &MerkleNode{
				Hash:     hashData(item.Value),
				Data:     item.Value,
				IsLeaf:   true,
				Key:      item.Key,
				Modified: item.Modified,
			}
		}
	}

	mt.rebuildTree()
	return nil
}

// Returns a copy of the leaves map for external access with read-safe operations
func (mt *MerkleTree) GetLeaves() map[string]*MerkleNode {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	leaves := make(map[string]*MerkleNode, len(mt.Leaves))
	for key, leaf := range mt.Leaves {
		leaves[key] = leaf
	}
	return leaves
}

// Creates a SHA256 hash of the data for tree construction and integrity verification
func hashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Returns statistics about the Merkle tree for monitoring and debugging
func (mt *MerkleTree) GetStats() map[string]interface{} {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	return map[string]interface{}{
		"total_leaves": len(mt.Leaves),
		"tree_hash":    mt.GetTreeHash(),
		"max_depth":    mt.MaxDepth,
	}
}
