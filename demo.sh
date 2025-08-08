#!/bin/bash

# Fringe Demo Script
# This script demonstrates the SWIM gossip membership and Merkle tree sync

set -e

echo "🚀 Starting Fringe Demo"
echo "========================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.24+"
        exit 1
    fi
    
    if ! command -v cargo &> /dev/null; then
        print_error "Rust is not installed. Please install Rust 1.70+"
        exit 1
    fi
    
    print_success "All dependencies are installed"
}

# Build the project
build_project() {
    print_status "Building Fringe components..."
    
    # Build Go components
    go mod tidy
    go build -o bin/fringe-edge cmd/edge/main.go
    go build -o bin/fringe-dashboard cmd/dashboard/main.go
    
    # Build Rust CLI
    cd cli && cargo build --release && cd ..
    cp cli/target/release/fringe-cli bin/
    
    print_success "Build completed"
}

# Start the demo cluster
start_demo() {
    print_status "Starting Fringe demo cluster..."
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Start bootstrap node in background
    print_status "Starting bootstrap node..."
    ./bin/fringe-edge --bootstrap --port 8080 --metrics-port 9090 > logs/node1.log 2>&1 &
    NODE1_PID=$!
    sleep 3
    
    # Start second node in background
    print_status "Starting second node..."
    ./bin/fringe-edge --node 127.0.0.1:8080 --port 8081 --metrics-port 9091 > logs/node2.log 2>&1 &
    NODE2_PID=$!
    sleep 3
    
    # Start third node in background
    print_status "Starting third node..."
    ./bin/fringe-edge --node 127.0.0.1:8080 --port 8082 --metrics-port 9092 > logs/node3.log 2>&1 &
    NODE3_PID=$!
    sleep 3
    
    # Start dashboard in background
    print_status "Starting dashboard..."
    ./bin/fringe-dashboard 8080 > logs/dashboard.log 2>&1 &
    DASHBOARD_PID=$!
    sleep 2
    
    print_success "Demo cluster started"
    echo "Node PIDs: $NODE1_PID, $NODE2_PID, $NODE3_PID, $DASHBOARD_PID"
    
    # Store PIDs for cleanup
    echo "$NODE1_PID $NODE2_PID $NODE3_PID $DASHBOARD_PID" > .demo_pids
}

# Demonstrate CLI operations
demo_cli() {
    print_status "Demonstrating CLI operations..."
    
    # Wait for nodes to be ready
    sleep 5
    
    # List nodes
    print_status "Listing cluster nodes..."
    ./bin/fringe-cli list --node 127.0.0.1:8080
    
    # Add some test data
    print_status "Adding test data to node 1..."
    ./bin/fringe-cli add --node 127.0.0.1:8080 --key sensor-data --value '{"temperature": 25.5, "humidity": 60}'
    
    print_status "Adding test data to node 2..."
    ./bin/fringe-cli add --node 127.0.0.1:8081 --key config --value '{"version": 1, "features": ["sync", "gossip"]}'
    
    # Get data
    print_status "Retrieving data from nodes..."
    ./bin/fringe-cli get --node 127.0.0.1:8080 --key sensor-data
    ./bin/fringe-cli get --node 127.0.0.1:8081 --key config
    
    # Sync data
    print_status "Syncing data between nodes..."
    ./bin/fringe-cli sync --from 127.0.0.1:8080 --to 127.0.0.1:8081
    
    print_success "CLI demo completed"
}

# Show dashboard info
show_dashboard() {
    print_status "Dashboard is running at: http://localhost:8080"
    print_status "Metrics endpoints:"
    echo "  - Node 1: http://localhost:9090/metrics"
    echo "  - Node 2: http://localhost:9091/metrics"
    echo "  - Node 3: http://localhost:9092/metrics"
    echo "  - Dashboard: http://localhost:8080/metrics"
}

# Cleanup function
cleanup() {
    print_status "Cleaning up demo processes..."
    
    if [ -f .demo_pids ]; then
        for pid in $(cat .demo_pids); do
            if kill -0 $pid 2>/dev/null; then
                kill $pid
                print_status "Killed process $pid"
            fi
        done
        rm -f .demo_pids
    fi
    
    print_success "Cleanup completed"
}

# Main demo function
main() {
    # Create logs directory
    mkdir -p logs
    
    # Set up cleanup on exit
    trap cleanup EXIT
    
    check_dependencies
    build_project
    start_demo
    demo_cli
    show_dashboard
    
    print_success "Demo is running! Press Ctrl+C to stop."
    
    # Keep the script running
    while true; do
        sleep 10
        print_status "Demo is still running... (Press Ctrl+C to stop)"
    done
}

# Handle command line arguments
case "${1:-}" in
    "cleanup")
        cleanup
        ;;
    "build")
        check_dependencies
        build_project
        ;;
    "start")
        start_demo
        show_dashboard
        print_success "Demo started! Press Ctrl+C to stop."
        while true; do sleep 10; done
        ;;
    *)
        main
        ;;
esac
