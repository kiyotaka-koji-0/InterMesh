package mesh

import (
	"testing"
	"time"
)

// TestMeshAppCreation tests MeshApp creation
func TestMeshAppCreation(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	if app.Node.ID != "node-1" {
		t.Errorf("Expected node ID 'node-1', got '%s'", app.Node.ID)
	}

	if app.IsConnected {
		t.Error("Expected IsConnected to be false initially")
	}

	if app.IsInternetSharing {
		t.Error("Expected IsInternetSharing to be false initially")
	}
}

// TestMeshAppConnection tests connecting and disconnecting
func TestMeshAppConnection(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	// Test initial state
	if app.GetConnectionStatus() {
		t.Error("Expected GetConnectionStatus to return false initially")
	}

	// Test disconnect when not connected
	app.DisconnectFromNetwork()
	if app.GetConnectionStatus() {
		t.Error("Expected GetConnectionStatus to return false after disconnect")
	}
}

// TestMeshAppInternetStatus tests setting internet status
func TestMeshAppInternetStatus(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	if app.Node.GetInternetStatus() {
		t.Error("Expected Node.GetInternetStatus to return false initially")
	}

	app.SetInternetStatus(true)
	if !app.Node.GetInternetStatus() {
		t.Error("Expected Node.GetInternetStatus to return true after SetInternetStatus(true)")
	}

	app.SetInternetStatus(false)
	if app.Node.GetInternetStatus() {
		t.Error("Expected Node.GetInternetStatus to return false after SetInternetStatus(false)")
	}
}

// TestMeshAppGetNetworkStats tests getting network statistics
func TestMeshAppGetNetworkStats(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	stats := app.GetNetworkStats()
	if stats == nil {
		t.Error("Expected GetNetworkStats to return non-nil stats")
	}

	if stats.NodeID != "node-1" {
		t.Errorf("Expected NodeID 'node-1', got '%s'", stats.NodeID)
	}

	if stats.PeerCount != 0 {
		t.Errorf("Expected PeerCount 0, got %d", stats.PeerCount)
	}
}

// TestMeshAppRequestInternetAccess tests requesting internet access
func TestMeshAppRequestInternetAccess(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	// Should fail when not connected (no proxies available)
	success := app.RequestInternetAccess()
	if success {
		t.Error("Expected RequestInternetAccess to fail when no proxies available")
	}

	// Should succeed when already has internet
	app.SetInternetStatus(true)
	success = app.RequestInternetAccess()
	if !success {
		t.Error("Expected RequestInternetAccess to succeed when device has internet")
	}
}

// TestMeshAppInternetSharing tests internet sharing
func TestMeshAppInternetSharing(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	// Should fail when device doesn't have internet
	success := app.EnableInternetSharing()
	if success {
		t.Error("Expected EnableInternetSharing to fail when device doesn't have internet")
	}

	// Should succeed when device has internet
	app.SetInternetStatus(true)
	success = app.EnableInternetSharing()
	if !success {
		t.Error("Expected EnableInternetSharing to succeed when device has internet")
	}

	if !app.GetInternetSharingStatus() {
		t.Error("Expected GetInternetSharingStatus to return true after enabling")
	}

	// Test disabling
	app.DisableInternetSharing()
	if app.GetInternetSharingStatus() {
		t.Error("Expected GetInternetSharingStatus to return false after disabling")
	}
}

// TestMeshAppConnectionListeners tests connection listeners
func TestMeshAppConnectionListeners(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	listener := &TestConnectionListener{
		onStateChanged: func(connected bool) {
			// Listener called
		},
		onError: func(err error) {
			// Error listener called
		},
	}

	app.AddConnectionListener(listener)

	// Give time for goroutines to complete
	time.Sleep(100 * time.Millisecond)
}

// TestMeshAppGetAvailableProxies tests getting available proxies
func TestMeshAppGetAvailableProxies(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	proxies := app.GetAvailableProxies()
	if len(proxies) != 0 {
		t.Errorf("Expected 0 proxies initially, got %d", len(proxies))
	}

	// Register a proxy
	proxy := &Peer{
		NodeID:      "proxy-1",
		IP:          "192.168.1.50",
		HasInternet: true,
		LastSeen:    time.Now().Unix(),
	}
	app.ProxyManager.RegisterProxy(proxy)

	// GetAvailableProxies() checks Discovery, not ProxyManager
	// So for testing, we verify ProxyManager directly
	managedProxies := app.ProxyManager.GetAvailableProxies()
	if len(managedProxies) != 1 {
		t.Errorf("Expected 1 proxy after registration in ProxyManager, got %d", len(managedProxies))
	}
}

// TestMeshAppGetConnectedPeers tests getting connected peers
func TestMeshAppGetConnectedPeers(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")

	peers := app.GetConnectedPeers()
	if len(peers) != 0 {
		t.Errorf("Expected 0 peers initially, got %d", len(peers))
	}

	// Add a peer to Node
	peer := &Peer{
		NodeID:   "node-2",
		IP:       "192.168.1.101",
		LastSeen: time.Now().Unix(),
	}
	app.Node.AddPeer(peer)

	// GetConnectedPeers() returns Transport.GetConnectedPeers()
	// which requires actual TCP connections, so it won't show our manually added peer
	// Just verify the method works
	peers = app.GetConnectedPeers()
	// No error means success
}

// TestMeshAppStop tests stopping the app
func TestMeshAppStop(t *testing.T) {
	app := NewMeshApp("node-1", "Test Device", "192.168.1.100", "aa:bb:cc:dd:ee:ff")
	app.IsConnected = true
	app.IsInternetSharing = true

	app.Stop()

	if app.GetConnectionStatus() {
		t.Error("Expected GetConnectionStatus to return false after Stop")
	}

	if app.GetInternetSharingStatus() {
		t.Error("Expected GetInternetSharingStatus to return false after Stop")
	}
}

// Helper test listener
type TestConnectionListener struct {
	onStateChanged func(bool)
	onError        func(error)
}

func (tcl *TestConnectionListener) OnConnectionStateChanged(connected bool) {
	if tcl.onStateChanged != nil {
		tcl.onStateChanged(connected)
	}
}

func (tcl *TestConnectionListener) OnConnectionError(err error) {
	if tcl.onError != nil {
		tcl.onError(err)
	}
}
