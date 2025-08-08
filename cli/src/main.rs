use clap::{Parser, Subcommand};
use serde::{Deserialize, Serialize};
use std::process;
use tracing::{error, info, Level};
use tracing_subscriber;

#[derive(Parser)]
#[command(name = "fringe-cli")]
#[command(about = "Fringe cluster management CLI")]
#[command(version = "0.1.0")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Start a new node
    Start {
        /// Node address (e.g., 127.0.0.1:8080)
        #[arg(long)]
        addr: Option<String>,
        
        /// Bootstrap node (first node in cluster)
        #[arg(long)]
        bootstrap: bool,
        
        /// Join existing cluster via known node
        #[arg(long)]
        join: Option<String>,
        
        /// Metrics port
        #[arg(long, default_value = "9090")]
        metrics_port: u16,
    },
    
    /// List all nodes in the cluster
    List {
        /// Node to query
        #[arg(long)]
        node: String,
    },
    
    /// Get node status
    Status {
        /// Node to query
        #[arg(long)]
        node: String,
    },
    
    /// Add data to a node
    Add {
        /// Node to add data to
        #[arg(long)]
        node: String,
        
        /// Data key
        #[arg(long)]
        key: String,
        
        /// Data value
        #[arg(long)]
        value: String,
    },
    
    /// Get data from a node
    Get {
        /// Node to query
        #[arg(long)]
        node: String,
        
        /// Data key
        #[arg(long)]
        key: String,
    },
    
    /// Delete data from a node
    Delete {
        /// Node to delete data from
        #[arg(long)]
        node: String,
        
        /// Data key
        #[arg(long)]
        key: String,
    },
    
    /// Sync data between nodes
    Sync {
        /// Source node
        #[arg(long)]
        from: String,
        
        /// Target node
        #[arg(long)]
        to: String,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct NodeInfo {
    id: String,
    address: String,
    state: String,
    incarnation: u64,
    since_state_update: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct ClusterStatus {
    nodes: Vec<NodeInfo>,
    total_nodes: usize,
    alive_nodes: usize,
}

#[derive(Debug, Serialize, Deserialize)]
struct DataItem {
    key: String,
    value: String,
    modified: String,
    version: u64,
}

#[derive(Debug, Serialize, Deserialize)]
struct SyncStats {
    tree_hash: String,
    total_leaves: usize,
    max_depth: usize,
}

// Starts a new Fringe node with configurable bootstrap and join behavior
async fn start_node(addr: Option<String>, bootstrap: bool, join: Option<String>, metrics_port: u16) -> anyhow::Result<()> {
    info!("Starting Fringe node...");
    
    let node_addr = addr.unwrap_or_else(|| {
        let port = 8080 + rand::random::<u16>() % 1000;
        format!("127.0.0.1:{}", port)
    });
    
    let mut cmd = std::process::Command::new("go");
    cmd.arg("run")
        .arg("cmd/edge/main.go")
        .arg("--port")
        .arg(node_addr.split(':').nth(1).unwrap_or("8080"))
        .arg("--metrics-port")
        .arg(metrics_port.to_string());
    
    if bootstrap {
        cmd.arg("--bootstrap");
    }
    
    if let Some(join_node) = join {
        cmd.arg("--node").arg(join_node);
    }
    
    info!("Executing: {:?}", cmd);
    let status = cmd.status()?;
    
    if status.success() {
        info!("Node started successfully");
    } else {
        error!("Failed to start node");
        process::exit(1);
    }
    
    Ok(())
}

// Retrieves and displays all nodes in the cluster with status information
async fn list_nodes(node: String) -> anyhow::Result<()> {
    let url = format!("http://{}:9090/cluster", node);
    let client = reqwest::Client::new();
    
    match client.get(&url).send().await {
        Ok(response) => {
            if response.status().is_success() {
                let cluster_status: ClusterStatus = response.json().await?;
                println!("Cluster Status:");
                println!("Total nodes: {}", cluster_status.total_nodes);
                println!("Alive nodes: {}", cluster_status.alive_nodes);
                println!("\nNodes:");
                
                for node_info in cluster_status.nodes {
                    println!("  ID: {}", node_info.id);
                    println!("  Address: {}", node_info.address);
                    println!("  State: {}", node_info.state);
                    println!("  Incarnation: {}", node_info.incarnation);
                    println!("  Since: {}", node_info.since_state_update);
                    println!();
                }
            } else {
                error!("Failed to get cluster status: {}", response.status());
                process::exit(1);
            }
        }
        Err(e) => {
            error!("Failed to connect to node {}: {}", node, e);
            process::exit(1);
        }
    }
    
    Ok(())
}

// Checks and reports the health status of a specific node
async fn get_node_status(node: String) -> anyhow::Result<()> {
    let url = format!("http://{}:9090/health", node);
    let client = reqwest::Client::new();
    
    match client.get(&url).send().await {
        Ok(response) => {
            if response.status().is_success() {
                println!("Node {} is healthy", node);
            } else {
                println!("Node {} is unhealthy", node);
                process::exit(1);
            }
        }
        Err(e) => {
            error!("Failed to connect to node {}: {}", node, e);
            process::exit(1);
        }
    }
    
    Ok(())
}

// Adds data to a specified node with proper JSON serialization
async fn add_data(node: String, key: String, value: String) -> anyhow::Result<()> {
    let url = format!("http://{}:9090/data", node);
    let client = reqwest::Client::new();
    
    let data = serde_json::json!({
        "key": key,
        "value": value,
        "action": "add"
    });
    
    match client.post(&url)
        .header("Content-Type", "application/json")
        .json(&data)
        .send()
        .await
    {
        Ok(response) => {
            if response.status().is_success() {
                println!("Data added successfully to node {}", node);
            } else {
                error!("Failed to add data: {}", response.status());
                process::exit(1);
            }
        }
        Err(e) => {
            error!("Failed to connect to node {}: {}", node, e);
            process::exit(1);
        }
    }
    
    Ok(())
}

// Retrieves data from a specified node with proper error handling
async fn get_data(node: String, key: String) -> anyhow::Result<()> {
    let url = format!("http://{}:9090/data/{}", node, key);
    let client = reqwest::Client::new();
    
    match client.get(&url).send().await {
        Ok(response) => {
            if response.status().is_success() {
                let data: DataItem = response.json().await?;
                println!("Key: {}", data.key);
                println!("Value: {}", data.value);
                println!("Modified: {}", data.modified);
                println!("Version: {}", data.version);
            } else {
                error!("Failed to get data: {}", response.status());
                process::exit(1);
            }
        }
        Err(e) => {
            error!("Failed to connect to node {}: {}", node, e);
            process::exit(1);
        }
    }
    
    Ok(())
}

// Removes data from a specified node with proper validation
async fn delete_data(node: String, key: String) -> anyhow::Result<()> {
    let url = format!("http://{}:9090/data", node);
    let client = reqwest::Client::new();
    
    let data = serde_json::json!({
        "key": key,
        "action": "delete"
    });
    
    match client.post(&url)
        .header("Content-Type", "application/json")
        .json(&data)
        .send()
        .await
    {
        Ok(response) => {
            if response.status().is_success() {
                println!("Data deleted successfully from node {}", node);
            } else {
                error!("Failed to delete data: {}", response.status());
                process::exit(1);
            }
        }
        Err(e) => {
            error!("Failed to connect to node {}: {}", node, e);
            process::exit(1);
        }
    }
    
    Ok(())
}

// Synchronizes data between two nodes with progress reporting
async fn sync_data(from: String, to: String) -> anyhow::Result<()> {
    let url = format!("http://{}:9090/sync", to);
    let client = reqwest::Client::new();
    
    let sync_request = serde_json::json!({
        "from": from,
        "to": to
    });
    
    match client.post(&url)
        .header("Content-Type", "application/json")
        .json(&sync_request)
        .send()
        .await
    {
        Ok(response) => {
            if response.status().is_success() {
                let stats: SyncStats = response.json().await?;
                println!("Sync completed successfully");
                println!("Tree hash: {}", stats.tree_hash);
                println!("Total leaves: {}", stats.total_leaves);
                println!("Max depth: {}", stats.max_depth);
            } else {
                error!("Failed to sync data: {}", response.status());
                process::exit(1);
            }
        }
        Err(e) => {
            error!("Failed to sync data: {}", e);
            process::exit(1);
        }
    }
    
    Ok(())
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize logging with configurable level
    tracing_subscriber::fmt()
        .with_max_level(Level::INFO)
        .init();
    
    let cli = Cli::parse();
    
    match cli.command {
        Commands::Start { addr, bootstrap, join, metrics_port } => {
            start_node(addr, bootstrap, join, metrics_port).await?;
        }
        Commands::List { node } => {
            list_nodes(node).await?;
        }
        Commands::Status { node } => {
            get_node_status(node).await?;
        }
        Commands::Add { node, key, value } => {
            add_data(node, key, value).await?;
        }
        Commands::Get { node, key } => {
            get_data(node, key).await?;
        }
        Commands::Delete { node, key } => {
            delete_data(node, key).await?;
        }
        Commands::Sync { from, to } => {
            sync_data(from, to).await?;
        }
    }
    
    Ok(())
}
