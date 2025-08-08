package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	serial "github.com/jscottransom/fringe/internal/proto"
	"github.com/jscottransom/fringe/internal/swim"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics for cluster monitoring
var (
	clusterSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fringe_cluster_size",
		Help: "Current number of nodes in the cluster",
	})

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

func init() {
	prometheus.MustRegister(clusterSize)
	prometheus.MustRegister(pingLatency)
	prometheus.MustRegister(messageCounter)
}

// Starts a Fringe node in the SWIM cluster with configurable bootstrap and join behavior
func main() {
	bootstrap := flag.Bool("bootstrap", false, "Set node as a bootstrap node")
	knownNode := flag.String("node", "", "Address of an existing node in the cluster")
	port := flag.Int("port", 0, "Port to listen on (0 for random)")
	metricsPort := flag.Int("metrics-port", 9090, "Port for metrics endpoint")
	flag.Parse()

	udp, err := net.ListenUDP("udp", &net.UDPAddr{Port: *port})
	if err != nil {
		log.Fatalf("failed to listen UDP: %v", err)
	}
	defer udp.Close()

	nodeAddr := udp.LocalAddr().String()
	nodeID := fmt.Sprintf("node-%s", nodeAddr)

	log.Printf("Starting Fringe node: %s", nodeID)

	node, err := initNode(nodeID, nodeAddr, *bootstrap)
	if err != nil {
		log.Fatalf("failed to initialize node: %v", err)
	}

	go startMetricsServer(*metricsPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go node.StartGossip(ctx)

	if !*bootstrap && *knownNode != "" {
		if err := node.JoinCluster(*knownNode); err != nil {
			log.Printf("Failed to join cluster: %v", err)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down...")
	cancel()
}

// Creates and initializes a new Fringe node with member table and piggyback queue
func initNode(nodeID, nodeAddr string, bootstrap bool) (*swim.Node, error) {
	memberTable := &swim.NodeTable{
		Members: make(map[string]*swim.Peer),
	}

	queue := &swim.PiggyBackQueue{
		Entries:  make([]*swim.Entry, 0),
		Capacity: 100,
	}

	selfPeer := &swim.Peer{
		PeerID:           nodeID,
		Address:          nodeAddr,
		State:            swim.Alive,
		Incarnation:      1,
		SinceStateUpdate: time.Now(),
	}

	memberTable.AddPeer(nodeID, selfPeer)

	initUpdate := &serial.MembershipUpdate{
		NodeId:      nodeID,
		Address:     nodeAddr,
		Incarnation: 1,
		State:       serial.State_ALIVE,
	}

	queue.AddEntry(&swim.Entry{
		Update:        initUpdate,
		Expiry:        time.Now().Add(swim.PeerTTL),
		DeliveryCount: 0,
		SeenPeers:     make(map[string]bool),
	})

	node := &swim.Node{
		NodeId:      nodeID,
		MemberTable: memberTable,
		Queue:       queue,
		Addr:        nodeAddr,
		Bootstrap:   bootstrap,
	}

	return node, nil
}

// Starts the Prometheus metrics server on the specified port with health endpoint
func startMetricsServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting metrics server on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Printf("Metrics server error: %v", err)
	}
}
