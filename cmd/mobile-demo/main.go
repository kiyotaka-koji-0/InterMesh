package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kiyotaka-koji-0/intermesh/pkg/mesh"
)

// This is an example CLI-based mobile app simulation
// It demonstrates how to use the mesh app interface

func main() {
	// Create a new mesh app instance
	app := mesh.NewMeshApp(
		"device-001",
		"My Mobile Device",
		"192.168.1.100",
		"aa:bb:cc:dd:ee:01",
	)

	fmt.Println("\n=== InterMesh Mobile App Demo ===\n")

	log.Println("Starting InterMesh Mobile App Demo...")

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start app: %v", err)
	}

	// Simulate user interactions
	simulateUserInteractions(app)

	// Cleanup
	app.Stop()
	fmt.Println("\nApp stopped.")
}

func simulateUserInteractions(app *mesh.MeshApp) {
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n[Demo] Simulating user actions...\n")

	// Action 1: Connect to network
	fmt.Println("→ Connecting to mesh network...")
	if err := app.ConnectToNetwork(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	displayAppStatus(app)
	time.Sleep(500 * time.Millisecond)

	// Action 2: Simulate having internet
	fmt.Println("\n→ Device detected internet connectivity...")
	app.SetInternetStatus(true)
	time.Sleep(500 * time.Millisecond)

	displayAppStatus(app)
	time.Sleep(500 * time.Millisecond)

	// Action 3: Enable internet sharing
	fmt.Println("\n→ Enabling internet sharing...")
	success := app.EnableInternetSharing()
	if !success {
		fmt.Println("Warning: Failed to enable internet sharing (no internet connection)")
	}
	time.Sleep(500 * time.Millisecond)

	displayAppStatus(app)
	time.Sleep(500 * time.Millisecond)

	// Action 4: Show network info
	fmt.Println("\n=== Network Information ===")
	stats := app.GetNetworkStats()
	fmt.Printf("Device ID: %s\n", stats.NodeID)
	fmt.Printf("Connected Peers: %d\n", stats.PeerCount)
	fmt.Printf("Available Proxies: %d\n", stats.AvailableProxies)
	fmt.Printf("Has Internet: %v\n", stats.InternetStatus)
	fmt.Printf("Sharing Enabled: %v\n", stats.InternetSharingEnabled)
	fmt.Printf("Connected Networks: %d\n", stats.ConnectedNetworks)
	time.Sleep(500 * time.Millisecond)

	// Action 5: Disable sharing
	fmt.Println("\n→ Disabling internet sharing...")
	app.DisableInternetSharing()
	time.Sleep(500 * time.Millisecond)

	displayAppStatus(app)
	time.Sleep(500 * time.Millisecond)

	// Action 6: Disconnect
	fmt.Println("\n→ Disconnecting from mesh network...")
	app.DisconnectFromNetwork()
	time.Sleep(500 * time.Millisecond)

	displayAppStatus(app)
}

func displayAppStatus(app *mesh.MeshApp) {
	fmt.Println("\n=== App Status ===")
	fmt.Printf("Connected: %v\n", app.GetConnectionStatus())
	fmt.Printf("Sharing Enabled: %v\n", app.GetInternetSharingStatus())
	fmt.Printf("Has Internet: %v\n", app.Node.GetInternetStatus())

	stats := app.GetNetworkStats()
	fmt.Printf("Connected Peers: %d\n", stats.PeerCount)
	fmt.Printf("Available Proxies: %d\n", stats.AvailableProxies)
}
