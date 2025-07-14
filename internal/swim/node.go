package swim

import (
	"context"
	"crypto/tls"
	"fmt"
	serial "github.com/jscottransom/fringe/internal/proto"
	quic "github.com/quic-go/quic-go"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net"
	"sync"
	"time"
	"errors"
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

// Setup the basic structure of a Node.
// Nodes contain a string UUID represent the process that should survive restarts and failures (stable)
// Peer Table contains a mapping of these identifiers to Node Objects
type Node struct {
	NodeId      string
	MemberTable *NodeTable
	Queue       *PiggyBackQueue
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
func (n *NodeTable) updatePeer(update *serial.MembershipUpdate, suspect bool) {
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

// Initialize a connection and stream over UDP using QUIC
// Stream is closed once finished.
func (n *Node) initUDPStream(addr string) (*quic.Stream, error) {
	// 1. Initialize a quic.Transport
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0}) // Use port 0 for an ephemeral port
	if err != nil {
		log.Fatalf("failed to listen UDP: %v", err)
		return nil, nil
	}
	defer udpConn.Close()

	tr := &quic.Transport{
		Conn: udpConn,
	}
	defer tr.Close()

	tlsConf := &tls.Config{
		InsecureSkipVerify: true, // Will revisit
		NextProtos:         []string{"quic"},
	}

	quicConf := &quic.Config{
		HandshakeIdleTimeout: 30 * time.Second,
	}

	fmt.Printf("Dialing QUIC connection to %s...\n", addr)

	// Resolve the target address into a net addr object
	udpAddr, err := net.ResolveUDPAddr("udp", addr)

	conn, err := tr.Dial(context.Background(), udpAddr, tlsConf, quicConf)
	if err != nil {
		log.Fatalf("failed to dial QUIC connection: %v", err)
		return nil, nil
	}
	defer conn.CloseWithError(quic.ApplicationErrorCode(0), "")
	fmt.Println("QUIC connection established!")

	// Open a new stream for back and forth communication
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatalf("failed to open stream: %v", err)
		return nil, nil
	}


	defer stream.Close()

	return stream, nil

}

// Send a Ping to a Node
func (n *Node) sendPing(ping *serial.Ping) error {
	// Open stream to target
	stream, err := n.initUDPStream(ping.TargetAddress)
	if err != nil {
		return fmt.Errorf("failed to open stream to %s: %w", ping.TargetAddress, err)
	}
	defer stream.Close()

	
	_ = stream.SetDeadline(time.Now().Add(3 * time.Second))

	data, err := proto.Marshal(ping)
	if err != nil {
		return fmt.Errorf("failed to marshal ping: %w", err)
	}

	
	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write ping to %s: %w", ping.TargetAddress, err)
	}
	fmt.Printf("Ping sent to %s\n", ping.TargetAddress)

	
	resp, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("failed to read response from %s: %w", ping.TargetAddress, err)
	}

	var ack serial.Ack
	if err := proto.Unmarshal(resp, &ack); err != nil {
		return n.handleAck(&ack, ping.TargetID)
	} else {
		return n.handleNack()
	}
}

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
		log.Printf("Received Ping from %s\n", ping.SenderId)

		// 2. Build and send Ack response

		// Construct serial updates from the update queue
		var updates []*serial.MembershipUpdate
		entries := n.Queue.GetEntries(n.NodeId, 5)
		for _, entry := range entries {
			update := entry.update
			updates = append(updates, update)

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
			log.Printf("failed to write ack to stream: %v", err)
			return
		}

		log.Printf("Sent Ack to %s", ping.SenderId)

	}()
}

// Send a Ping to a Node
func (n *Node) sendPingReq(pingReq *serial.PingReq) error {

	// Send a PingReq for a given node, to a known, alive node
	stream, err := n.initUDPStream(pingReq.RequestAddress)
	if err != nil {
		return fmt.Errorf("failed to open stream to %s: %w", pingReq.RequestAddress, err)
	}

	defer stream.Close()

	// Set deadline for establishiing a stream connection
	_ = stream.SetDeadline(time.Now().Add(3 * time.Second))

	// Write data to the stream
	// Serialize (marshal) the pingReq message. 
	message, err := proto.Marshal(pingReq)
	if err != nil {
		return fmt.Errorf("failed to marshal Ping Req to %s: %w", pingReq.RequestAddress, err)
	}
	_, err = stream.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to stream: %v", err)
	}
	fmt.Printf("Ping sent to %s\n", pingReq.RequestAddress)

	// Read response from the stream
	resp, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %v", err)
	}
	
	// Handle Ack
	var ack serial.Ack
	if err := proto.Unmarshal(resp, &ack); err == nil {
		return n.handleAck(&ack)
	}

    // Handle Nack
	var nack serial.Nack
	if err := proto.Unmarshal(resp, &nack); err == nil {
		return n.handleNack(&nack)
	}

	return fmt.Errorf("received unknown or invalid response from %s", ping.TargetAddress)
	
}

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
		log.Printf("Received Ping from %s\n", pingReq.SenderId)

		// 2. Ping the Target node 

		// Serialize Ping Message to send across the network
		// Construct serial updates from the update queue
		var updates []*serial.MembershipUpdate
		entries := n.Queue.GetEntries(n.NodeId, 5)
		for _, entry := range entries {
			update := entry.update
			updates = append(updates, update)
		}
		
		ping := &serial.Ping{
			SenderId: n.NodeId,
			SenderAddress: pingReq.RequestAddress,
			TargetAddress: pingReq.TargetAddress,
			Updates: updates,
		}
		err = n.sendPing(ping)
		if err != nil {
			// Construct a Nack to send back to the original requestor
			nack := &serial.Ack{
			Response:      "Nack",
			SenderId:      ping.SenderId,
			SenderAddress: ping.SenderAddress,
			Incarnation:   n.MemberTable.Members[n.NodeId].Incarnation,
			TargetId:      pingReq.SenderId,
			Updates:       updates,
		}

		nackData, _ := proto.Marshal(nack)
		stream.Write(nackData)
		} 

		// Otherwise send the Ack to the original requestor

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
		
	}()
}

func (n *Node) handleAck(ack *serial.Ack) error {
	// Locate the Peer update
	for _, update := range ack.Updates {
		// Match the Node ID to the ack sender, which is the Peer we want to update
		if update.NodeId == ack.SenderId {
			n.MemberTable.updatePeer(update, true) 
		} 
	}
	return nil

}
// Handle a "non-response" from a Ping or a PingReq
func (n *Node) handleNack(id string) error {
	// Locate the Peer update
	n.MemberTable.mu.Lock()
	defer n.MemberTable.mu.Unlock()

	// Set the appropriate status 
	peer := n.MemberTable.Members[id]
	
	switch peer.State {
	case Alive:
		// If the state was previously Alive, we are now suspicious
		peer.State = Suspected
	case Suspected:
		// Move to "Dead" if the TTL for the uspected status is met
		if time.Since(peer.SinceStateUpdate) > PeerTTL {
			peer.State = Dead
		} else {
			peer.State = Suspected
		}
	}
	
	return nil

}


// Generic handler to forward response type
// switch m := env.Msg.(type) {
// case *serial.Envelope_Ping:
//     handlePing(m.Ping)
// case *serial.Envelope_Ack:
//     handleAck(m.Ack)
// case *serial.Envelope_PingReq:
//     handlePingReq(m.PingReq)
// }
