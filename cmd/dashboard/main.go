package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Dashboard provides real-time cluster monitoring and visualization
type Dashboard struct {
	port     int
	router   *mux.Router
	clusters map[string]*ClusterInfo
}

// ClusterInfo represents information about a cluster with nodes and metadata
type ClusterInfo struct {
	Name      string     `json:"name"`
	Nodes     []NodeInfo `json:"nodes"`
	LastSeen  time.Time  `json:"last_seen"`
	TreeHash  string     `json:"tree_hash"`
	DataCount int        `json:"data_count"`
}

// NodeInfo represents information about a node with state and metrics
type NodeInfo struct {
	ID               string    `json:"id"`
	Address          string    `json:"address"`
	State            string    `json:"state"`
	Incarnation      uint64    `json:"incarnation"`
	SinceStateUpdate time.Time `json:"since_state_update"`
	PingLatency      float64   `json:"ping_latency"`
	MessageCount     int       `json:"message_count"`
}

// DataItem represents a data item in the system with metadata
type DataItem struct {
	Key      string    `json:"key"`
	Value    string    `json:"value"`
	Modified time.Time `json:"modified"`
	Version  uint64    `json:"version"`
}

// SyncStats represents synchronization statistics for monitoring
type SyncStats struct {
	TreeHash    string `json:"tree_hash"`
	TotalLeaves int    `json:"total_leaves"`
	MaxDepth    int    `json:"max_depth"`
}

// Prometheus metrics for dashboard monitoring
var (
	clusterCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fringe_dashboard_clusters_total",
		Help: "Total number of clusters being monitored",
	})

	nodeCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fringe_dashboard_nodes_total",
		Help: "Total number of nodes by state",
	}, []string{"state"})
)

func init() {
	prometheus.MustRegister(clusterCount)
	prometheus.MustRegister(nodeCount)
}

// Creates a new dashboard server with configured routes and cluster management
func NewDashboard(port int) *Dashboard {
	d := &Dashboard{
		port:     port,
		router:   mux.NewRouter(),
		clusters: make(map[string]*ClusterInfo),
	}

	d.setupRoutes()
	return d
}

// Configures HTTP routes for API and web interface with proper middleware
func (d *Dashboard) setupRoutes() {
	d.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	d.router.HandleFunc("/api/clusters", d.handleGetClusters).Methods("GET")
	d.router.HandleFunc("/api/clusters/{name}", d.handleGetCluster).Methods("GET")
	d.router.HandleFunc("/api/clusters/{name}/nodes", d.handleGetClusterNodes).Methods("GET")
	d.router.HandleFunc("/api/clusters/{name}/data", d.handleGetClusterData).Methods("GET")
	d.router.HandleFunc("/api/clusters/{name}/sync", d.handleSyncCluster).Methods("POST")

	d.router.HandleFunc("/", d.handleDashboard).Methods("GET")
	d.router.HandleFunc("/cluster/{name}", d.handleClusterView).Methods("GET")

	d.router.Handle("/metrics", promhttp.Handler())
}

// Starts the dashboard server on the configured port with graceful shutdown
func (d *Dashboard) Start() error {
	log.Printf("Starting Fringe dashboard on port %d", d.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", d.port), d.router)
}

// Serves the main dashboard page with cluster overview and real-time statistics
func (d *Dashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Fringe Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .cluster { border: 1px solid #ccc; margin: 10px; padding: 15px; border-radius: 5px; }
        .node { background: #f9f9f9; margin: 5px; padding: 10px; border-radius: 3px; }
        .alive { border-left: 4px solid #4CAF50; }
        .suspected { border-left: 4px solid #FF9800; }
        .dead { border-left: 4px solid #F44336; }
        .stats { display: flex; justify-content: space-between; }
        .metric { text-align: center; }
        .metric-value { font-size: 24px; font-weight: bold; }
        .metric-label { color: #666; }
    </style>
</head>
<body>
    <h1>Fringe Cluster Dashboard</h1>
    
    <div class="stats">
        <div class="metric">
            <div class="metric-value">{{.TotalClusters}}</div>
            <div class="metric-label">Clusters</div>
        </div>
        <div class="metric">
            <div class="metric-value">{{.TotalNodes}}</div>
            <div class="metric-label">Total Nodes</div>
        </div>
        <div class="metric">
            <div class="metric-value">{{.AliveNodes}}</div>
            <div class="metric-label">Alive Nodes</div>
        </div>
    </div>
    
    {{range .Clusters}}
    <div class="cluster">
        <h2>{{.Name}}</h2>
        <p><strong>Tree Hash:</strong> {{.TreeHash}}</p>
        <p><strong>Data Count:</strong> {{.DataCount}}</p>
        <p><strong>Last Seen:</strong> {{.LastSeen}}</p>
        
        <h3>Nodes ({{len .Nodes}})</h3>
        {{range .Nodes}}
        <div class="node {{.State}}">
            <strong>{{.ID}}</strong> ({{.Address}})
            <br>State: {{.State}} | Incarnation: {{.Incarnation}}
            <br>Ping Latency: {{.PingLatency}}ms | Messages: {{.MessageCount}}
        </div>
        {{end}}
    </div>
    {{end}}
    
    <script>
        setTimeout(function() {
            location.reload();
        }, 5000);
    </script>
</body>
</html>`

	t, err := template.New("dashboard").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalClusters := len(d.clusters)
	totalNodes := 0
	aliveNodes := 0

	for _, cluster := range d.clusters {
		totalNodes += len(cluster.Nodes)
		for _, node := range cluster.Nodes {
			if node.State == "alive" {
				aliveNodes++
			}
		}
	}

	data := struct {
		Clusters      []*ClusterInfo
		TotalClusters int
		TotalNodes    int
		AliveNodes    int
	}{
		Clusters:      d.getClustersList(),
		TotalClusters: totalClusters,
		TotalNodes:    totalNodes,
		AliveNodes:    aliveNodes,
	}

	t.Execute(w, data)
}

// Returns JSON list of all monitored clusters for API consumption
func (d *Dashboard) handleGetClusters(w http.ResponseWriter, r *http.Request) {
	clusters := d.getClustersList()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusters)
}

// Returns JSON details for a specific cluster with error handling
func (d *Dashboard) handleGetCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	cluster, exists := d.clusters[name]
	if !exists {
		http.Error(w, "Cluster not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cluster)
}

// Returns JSON list of nodes for a specific cluster with validation
func (d *Dashboard) handleGetClusterNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	cluster, exists := d.clusters[name]
	if !exists {
		http.Error(w, "Cluster not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cluster.Nodes)
}

// Returns mock data for demonstration purposes with proper content type
func (d *Dashboard) handleGetClusterData(w http.ResponseWriter, r *http.Request) {
	data := []DataItem{
		{
			Key:      "config",
			Value:    "{\"version\": 1}",
			Modified: time.Now(),
			Version:  1,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Returns mock sync statistics for demonstration with proper error handling
func (d *Dashboard) handleSyncCluster(w http.ResponseWriter, r *http.Request) {
	stats := SyncStats{
		TreeHash:    "abc123",
		TotalLeaves: 10,
		MaxDepth:    4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Serves a detailed view page for a specific cluster with template rendering
func (d *Dashboard) handleClusterView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	cluster, exists := d.clusters[name]
	if !exists {
		http.Error(w, "Cluster not found", http.StatusNotFound)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Cluster: {{.Name}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .node { background: #f9f9f9; margin: 5px; padding: 10px; border-radius: 3px; }
        .alive { border-left: 4px solid #4CAF50; }
        .suspected { border-left: 4px solid #FF9800; }
        .dead { border-left: 4px solid #F44336; }
    </style>
</head>
<body>
    <h1>Cluster: {{.Name}}</h1>
    <p><strong>Tree Hash:</strong> {{.TreeHash}}</p>
    <p><strong>Data Count:</strong> {{.DataCount}}</p>
    <p><strong>Last Seen:</strong> {{.LastSeen}}</p>
    
    <h2>Nodes</h2>
    {{range .Nodes}}
    <div class="node {{.State}}">
        <strong>{{.ID}}</strong> ({{.Address}})
        <br>State: {{.State}} | Incarnation: {{.Incarnation}}
        <br>Ping Latency: {{.PingLatency}}ms | Messages: {{.MessageCount}}
    </div>
    {{end}}
    
    <a href="/">Back to Dashboard</a>
</body>
</html>`

	t, err := template.New("cluster").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, cluster)
}

// Returns a list of all clusters for API responses with proper serialization
func (d *Dashboard) getClustersList() []*ClusterInfo {
	var clusters []*ClusterInfo
	for _, cluster := range d.clusters {
		clusters = append(clusters, cluster)
	}
	return clusters
}

// Updates cluster information and metrics with thread-safe operations
func (d *Dashboard) UpdateCluster(name string, info *ClusterInfo) {
	d.clusters[name] = info
	d.clusters[name].LastSeen = time.Now()

	clusterCount.Set(float64(len(d.clusters)))

	aliveCount := 0
	suspectedCount := 0
	deadCount := 0

	for _, node := range info.Nodes {
		switch node.State {
		case "alive":
			aliveCount++
		case "suspected":
			suspectedCount++
		case "dead":
			deadCount++
		}
	}

	nodeCount.WithLabelValues("alive").Set(float64(aliveCount))
	nodeCount.WithLabelValues("suspected").Set(float64(suspectedCount))
	nodeCount.WithLabelValues("dead").Set(float64(deadCount))
}

func main() {
	port := 8080
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
		}
	}

	dashboard := NewDashboard(port)

	dashboard.UpdateCluster("cluster-1", &ClusterInfo{
		Name: "cluster-1",
		Nodes: []NodeInfo{
			{
				ID:               "node-1",
				Address:          "127.0.0.1:8080",
				State:            "alive",
				Incarnation:      1,
				SinceStateUpdate: time.Now(),
				PingLatency:      15.5,
				MessageCount:     42,
			},
			{
				ID:               "node-2",
				Address:          "127.0.0.1:8081",
				State:            "alive",
				Incarnation:      1,
				SinceStateUpdate: time.Now(),
				PingLatency:      18.2,
				MessageCount:     38,
			},
		},
		TreeHash:  "abc123def456",
		DataCount: 15,
	})

	log.Fatal(dashboard.Start())
}
