package main

import (
	"fmt"
	"log"
	"net"
	"flag"
	serial "github.com/jscottransom/fringe/internal/proto"
	"github.com/jscottransom/fringe/internal/swim"
)


// Main entrypoint for a Node into the swim cluster
func main() {
	// Either a bootstrap node or connecting to a known node in the cluster 
	bootstrap := flag.Bool("start", true, "Set node as a bootstrap node")
	knownNode := flag.String("node", "127.0.0.1:1234", "Node ID of an existing Node")
	flag.Parse()

	// Generate a Node ID by obtaining a random, OS-assigned port
	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		log.Fatalf("failed to listen UDP: %v", err)
		return
	}

	port := udp.LocalAddr().(*net.UDPAddr).Port
	nodeID := udp.LocalAddr().String()

	// Construct the "Node" Object	
	peer, err := initPeer()
	if err != nil {
		log.Fatalf("failed to initialize Peer: %v", err)
		return
	}

	// Add Peer to MemberTable
	var memTable *swim.NodeTable
	memTable.Members[nodeID] = peer

	var updates []*serial.MembershipUpdate
	initUpdate := &serial.MembershipUpdate{
				NodeId: nodeID,
				Address: peer.Address,
				Incarnation: 1,
				State: serial.State_ALIVE,
	}
	updates = append(updates, initUpdate)
	self := &swim.Node{
		NodeId: nodeID,
		MemberTable: memTable,
		Queue: ,

	}






}
