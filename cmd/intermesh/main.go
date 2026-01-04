package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kiyotaka-koji-0/intermesh/pkg/mesh"
)

func main() {
	// Command-line flags
	nodeID := flag.String("id", "node-1", "Unique identifier for this node")
	nodeName := flag.String("name", "InterMesh Node", "Human-readable name for this node")
	ip := flag.String("ip", "", "IP address of this node (auto-detected if empty)")
	mac := flag.String("mac", "", "MAC address of this node (auto-detected if empty)")
	hasInternet := flag.Bool("internet", false, "Force internet status (auto-detected if not set)")
	autoDetect := flag.Bool("auto", true, "Auto-detect network configuration")

	flag.Parse()

	// Auto-detect network info if enabled
	var nodeIP, nodeMAC string
	var internetStatus bool

	if *autoDetect && (*ip == "" || *mac == "") {
		log.Println("Auto-detecting network configuration...")
		netInfo := mesh.DetectNetworkInfo()
		
		nodeIP = netInfo.IP
		nodeMAC = netInfo.MAC
		internetStatus = netInfo.HasInternet
		
		log.Printf("Detected interface: %s", netInfo.Interface)
	}

	// Override with flags if provided
	if *ip != "" {
		nodeIP = *ip
	}
	if *mac != "" {
		nodeMAC = *mac
	}
	if *hasInternet {
		internetStatus = true
	}

	// Create the node
	node := mesh.NewNode(*nodeID, *nodeName, nodeIP, nodeMAC)
	node.SetInternetStatus(internetStatus)

	log.Printf("Starting InterMesh node: %s (%s)", node.Name, node.ID)
	log.Printf("IP: %s, MAC: %s", node.IP, node.MAC)
	log.Printf("Internet connectivity: %v", node.GetInternetStatus())

	// Create the mesh manager
	manager := mesh.NewManager(node)

	// Create the personal network manager
	pnManager := mesh.NewPersonalNetworkManager()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the mesh manager
	go func() {
		if err := manager.Start(context.Background()); err != nil {
			log.Printf("Error starting mesh manager: %v", err)
		}
	}()

	// Example: Create a personal network
	personalNet := pnManager.CreateNetwork("pnet-1", "My Home Network", *nodeID)
	log.Printf("Created personal network: %s (%s)", personalNet.Name, personalNet.ID)

	// Add the current node as a member
	member := &mesh.NetworkMember{
		NodeID:      *nodeID,
		HasInternet: *hasInternet,
		IsProxy:     *hasInternet,
	}
	personalNet.AddMember(member)

	log.Println("InterMesh node is running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down InterMesh node...")

	// Cleanup
	manager.Stop()
	log.Println("InterMesh node stopped.")
}
