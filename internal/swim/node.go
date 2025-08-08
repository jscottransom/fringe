package swim

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	serial "github.com/jscottransom/fringe/internal/proto"
	"github.com/prometheus/client_golang/prometheus"
	quic "github.com/quic-go/quic-go"
	"google.golang.org/protobuf/proto"
)

const PeerTTL = 60 * time.Second

// Prometheus metrics for SWIM protocol monitoring
var (
	pingLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "fringe_ping_latency_seconds",
		Help:    "Ping latency in seconds",
		Buckets: prometheus.DefBuckets,
	})

	messageCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "fringe_messages_total",
		Help: "Total number of messages by type",
	}, []string{"type"})
)

// NodeState represents the state of a node in the cluster
type NodeState int

const (
	Alive NodeState = iota
	Suspected
	Dead
	Left
)

// Peer represents membership level data exchanged across the network
type Peer struct {
	PeerID           string
	Address          string
	State            NodeState
	Incarnation      uint64
	SinceStateUpdate time.Time
}

// Node represents a Fringe node in the SWIM cluster with member table and gossip protocol
type Node struct {
	NodeId      string
	MemberTable *NodeTable
	Queue       *PiggyBackQueue
	Addr        string
	Bootstrap   bool
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NodeTable maps node IDs to Peer objects with thread-safe operations
type NodeTable struct {
	Members map[string]*Peer
	mu      sync.RWMutex
}

// Safely adds a peer to the member table with thread-safe access
func (n *NodeTable) AddPeer(nodeID string, peer *Peer) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Members[nodeID] = peer
	fmt.Printf("Successfully added %s to Node Table", nodeID)
}

// Safely updates a peer's state based on membership updates with incarnation handling
func (n *NodeTable) UpdatePeer(update *serial.MembershipUpdate, suspect bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	peer := n.Members[update.NodeId]
	if peer == nil {
		return
	}

	switch {
	case peer.Incarnation > update.Incarnation:
		return
	case peer.Incarnation < update.Incarnation:
		peer.Incarnation = update.Incarnation
		peer.State = NodeState(*update.State.Enum())
	case peer.Incarnation == update.Incarnation:
		peer.State = max(peer.State, NodeState(*update.State.Enum()))
	}

	for _, state := range []NodeState{Suspected, Dead, Left} {
		if (peer.State == state) && (time.Since(peer.SinceStateUpdate) > PeerTTL) {
			delete(n.Members, update.NodeId)
		}
	}
}

// Returns a list of alive peers for ping selection with read-safe access
func (n *NodeTable) GetAlivePeers() []*Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var alivePeers []*Peer
	for _, peer := range n.Members {
		if peer.State == Alive {
			alivePeers = append(alivePeers, peer)
		}
	}
	return alivePeers
}

// Returns the number of alive nodes in the cluster with read-safe access
func (n *NodeTable) GetClusterSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	count := 0
	for _, peer := range n.Members {
		if peer.State == Alive {
			count++
		}
	}
	return count
}

// Initializes and starts the SWIM gossip protocol with periodic ping and cleanup
func (n *Node) StartGossip(ctx context.Context) {
	n.ctx = ctx

	go n.periodicPing()
	go n.periodicCleanup()

	log.Printf("Started gossip protocol for node %s", n.NodeId)
}

// Attempts to join an existing cluster via a known node with ping-based discovery
func (n *Node) JoinCluster(knownNodeAddr string) error {
	log.Printf("Joining cluster via node: %s", knownNodeAddr)

	ping := &serial.Ping{
		SenderId:      n.NodeId,
		SenderAddress: n.Addr,
		TargetId:      knownNodeAddr,
		Updates:       []*serial.MembershipUpdate{},
	}

	return n.sendPing(ping)
}

// Sends periodic pings to random peers every 5 seconds for failure detection
func (n *Node) periodicPing() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.sendRandomPing()
		}
	}
}

// Selects a random alive peer and sends a ping with piggybacked updates
func (n *Node) sendRandomPing() {
	alivePeers := n.MemberTable.GetAlivePeers()
	if len(alivePeers) == 0 {
		return
	}

	// Find first non-self peer
	var targetPeer *Peer
	for _, peer := range alivePeers {
		if peer.PeerID != n.NodeId {
			targetPeer = peer
			break
		}
	}
	if targetPeer == nil {
		return
	}

	// Build updates
	updates := n.Queue.GetEntries(n.NodeId, 5)
	serialUpdates := make([]*serial.MembershipUpdate, len(updates))
	for i, entry := range updates {
		serialUpdates[i] = entry.Update
	}

	ping := &serial.Ping{
		SenderId:      n.NodeId,
		SenderAddress: n.Addr,
		TargetId:      targetPeer.Address,
		Updates:       serialUpdates,
	}

	start := time.Now()
	if err := n.sendPing(ping); err != nil {
		log.Printf("Failed to ping %s: %v", targetPeer.Address, err)
	} else {
		pingLatency.Observe(time.Since(start).Seconds())
		messageCounter.WithLabelValues("ping").Inc()
	}
}

// Removes expired entries and updates metrics every 10 seconds
func (n *Node) periodicCleanup() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.Queue.EvictEntry()
		}
	}
}

// Establishes a QUIC connection and stream to the target address with TLS configuration
func (n *Node) initUDPStream(addr string) (*quic.Stream, error) {
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to listen UDP: %w", err)
	}
	defer udpConn.Close()

	tr := &quic.Transport{Conn: udpConn}
	defer tr.Close()

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic"},
	}

	quicConf := &quic.Config{
		HandshakeIdleTimeout: 30 * time.Second,
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := tr.Dial(context.Background(), udpAddr, tlsConf, quicConf)
	if err != nil {
		return nil, fmt.Errorf("failed to dial QUIC connection: %w", err)
	}
	defer conn.CloseWithError(quic.ApplicationErrorCode(0), "")

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return stream, nil
}

// Sends a ping message to a target node and handles the response with timeout
func (n *Node) sendPing(ping *serial.Ping) error {
	stream, err := n.initUDPStream(ping.TargetId)
	if err != nil {
		return fmt.Errorf("failed to open stream to %s: %w", ping.TargetId, err)
	}
	defer stream.Close()

	stream.SetDeadline(time.Now().Add(3 * time.Second))

	data, err := proto.Marshal(ping)
	if err != nil {
		return fmt.Errorf("failed to marshal ping: %w", err)
	}

	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write ping to %s: %w", ping.TargetId, err)
	}

	resp, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("failed to read response from %s: %w", ping.TargetId, err)
	}

	var ack serial.Ack
	if err := proto.Unmarshal(resp, &ack); err != nil {
		return n.handleAck(&ack)
	} else {
		return n.handleNack(ping.TargetId)
	}
}

// Processes incoming ping messages and sends acknowledgments with piggybacked updates
func (n *Node) handlePing(sess *quic.Conn) {
	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		log.Printf("failed to accept stream: %v", err)
		return
	}

	go func() {
		defer stream.Close()

		data, err := io.ReadAll(stream)
		if err != nil {
			log.Printf("failed to read stream: %v", err)
			return
		}

		var ping serial.Ping
		if err := proto.Unmarshal(data, &ping); err != nil {
			log.Printf("failed to Unmarshal Ping: %v", err)
			return
		}

		if ping.SenderId == n.NodeId {
			log.Printf("ignoring ping from self (%s)", ping.SenderId)
			return
		}

		// Build updates
		entries := n.Queue.GetEntries(n.NodeId, 5)
		updates := make([]*serial.MembershipUpdate, len(entries))
		for i, entry := range entries {
			updates[i] = entry.Update
		}

		ack := &serial.Ack{
			Response:      "Ack",
			SenderId:      n.NodeId,
			SenderAddress: n.MemberTable.Members[n.NodeId].Address,
			Incarnation:   n.MemberTable.Members[n.NodeId].Incarnation,
			TargetId:      ping.SenderId,
			Updates:       updates,
		}

		ackData, err := proto.Marshal(ack)
		if err != nil {
			log.Printf("failed to marshal ack: %v", err)
			return
		}
		if _, err := stream.Write(ackData); err != nil {
			stream.Write([]byte("Nack"))
			log.Printf("failed to write ack to stream: %v", err)
			return
		}
	}()
}

// Sends a ping request to probe a target node via an intermediate node with timeout
func (n *Node) sendPingReq(pingReq *serial.PingReq) error {
	stream, err := n.initUDPStream(pingReq.RequestAddress)
	if err != nil {
		return fmt.Errorf("failed to open stream to %s: %w", pingReq.RequestAddress, err)
	}
	defer stream.Close()

	stream.SetDeadline(time.Now().Add(3 * time.Second))

	message, err := proto.Marshal(pingReq)
	if err != nil {
		return fmt.Errorf("failed to marshal Ping Req to %s: %w", pingReq.RequestAddress, err)
	}
	_, err = stream.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to stream: %v", err)
	}

	resp, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %v", err)
	}

	var ack serial.Ack
	if err := proto.Unmarshal(resp, &ack); err != nil {
		return n.handleAck(&ack)
	} else {
		return n.handleNack(pingReq.RequestId)
	}
}

// Processes incoming ping requests and forwards them to target nodes with piggybacked updates
func (n *Node) handlePingReq(sess *quic.Conn) {
	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		log.Printf("failed to accept stream: %v", err)
		return
	}

	go func() {
		defer stream.Close()

		data, err := io.ReadAll(stream)
		if err != nil {
			log.Printf("failed to read stream: %v", err)
			return
		}

		var pingReq serial.PingReq
		if err := proto.Unmarshal(data, &pingReq); err != nil {
			log.Printf("failed to Unmarshal PingReq: %v", err)
			return
		}

		if pingReq.SenderId == n.NodeId {
			log.Printf("ignoring ping from self (%s)", pingReq.SenderId)
			return
		}

		// Build updates
		entries := n.Queue.GetEntries(n.NodeId, 5)
		updates := make([]*serial.MembershipUpdate, len(entries))
		for i, entry := range entries {
			updates[i] = entry.Update
		}

		ping := &serial.Ping{
			SenderId:      n.NodeId,
			SenderAddress: pingReq.RequestAddress,
			TargetId:      pingReq.TargetId,
			Updates:       updates,
		}
		err = n.sendPing(ping)
		if err != nil {
			stream.Write([]byte("Nack"))
		} else {
			ack := &serial.Ack{
				Response:      "Ack",
				SenderId:      n.NodeId,
				SenderAddress: pingReq.RequestAddress,
				Incarnation:   n.MemberTable.Members[n.NodeId].Incarnation,
				TargetId:      ping.SenderId,
				Updates:       updates,
			}

			ackData, _ := proto.Marshal(ack)
			stream.Write(ackData)
		}
	}()
}

// Processes acknowledgment messages and updates member table with received updates
func (n *Node) handleAck(ack *serial.Ack) error {
	for _, update := range ack.Updates {
		if update.NodeId == ack.SenderId {
			n.MemberTable.UpdatePeer(update, true)
		}
	}
	return nil
}

// Processes negative acknowledgments and updates node states for failure detection
func (n *Node) handleNack(id string) error {
	n.MemberTable.mu.Lock()
	defer n.MemberTable.mu.Unlock()

	peer := n.MemberTable.Members[id]
	if peer == nil {
		return fmt.Errorf("peer %s not found", id)
	}

	switch peer.State {
	case Alive:
		peer.State = Suspected
	case Suspected:
		if time.Since(peer.SinceStateUpdate) > PeerTTL {
			peer.State = Dead
		}
	}

	return nil
}
