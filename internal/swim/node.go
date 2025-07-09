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

	stream, err := n.initUDPStream(ping.TargetAddress)
	if err != nil {
		return err
	}

	// Write data to the stream
	// Serialize (marshal) the ping message
	message, err := proto.Marshal(ping)
	_, err = stream.Write([]byte(message))
	if err != nil {
		log.Fatalf("failed to write to stream: %v", err)
		return err
	}
	fmt.Println("Sent Ping")

	// Read response from the stream
	resp, err := io.ReadAll(stream)
	if err != nil {
		log.Fatalf("failed to read from stream: %v", err)
	}
	var ack serial.Ack
	err = proto.Unmarshal(resp, &ack)
	if err != nil {
		log.Fatalf("failed to unmarshal response: %v", err)
	}

	fmt.Println(ack.Response)
	return nil
}

func (n *Node) handlePing(sess *quic.Conn) {

	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		return
	}

	go func() {
		defer stream.Close()

		data, err := io.ReadAll(stream)
		if err != nil {
			return
		}

		var ping serial.Ping
		if err := proto.Unmarshal(data, &ping); err != nil {
			return
		}

		fmt.Printf("Received Ping from %s\n", ping.SenderId)

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
			SenderId:      ping.SenderId,
			SenderAddress: ping.SenderAddress,
			Incarnation:   n.MemberTable.Members[n.NodeId].Incarnation,
			TargetId:      n.NodeId,
			Updates:       updates,
		}

		ackData, _ := proto.Marshal(ack)
		stream.Write(ackData)
	}()
}

func (n *Node) sendPingReq(pingReq *serial.PingReq) error {

	stream, err := n.initUDPStream(pingReq.TargetAddress)
	if err != nil {
		return err
	}

	// Write data to the stream
	// Serialize (marshal) the ping message
	message, err := proto.Marshal(ping)
	_, err = stream.Write([]byte(message))
	if err != nil {
		log.Fatalf("failed to write to stream: %v", err)
		return err
	}
	fmt.Println("Sent Ping")

	// Read response from the stream
	resp, err := io.ReadAll(stream)
	if err != nil {
		log.Fatalf("failed to read from stream: %v", err)
	}
	ack := &serial.Ack{}
	err = proto.Unmarshal(resp, ack)
	if err != nil {
		log.Fatalf("failed to unmarshal response: %v", err)
	}

	fmt.Println(ack.Response)
	return nil
}
