package mesh

import (
	"testing"
)

// TestNodeCreation tests node creation
func TestNodeCreation(t *testing.T) {
	node := NewNode("node-1", "Test Node", "192.168.1.1", "aa:bb:cc:dd:ee:ff")

	if node.ID != "node-1" {
		t.Errorf("Expected ID 'node-1', got '%s'", node.ID)
	}

	if node.Name != "Test Node" {
		t.Errorf("Expected name 'Test Node', got '%s'", node.Name)
	}

	if len(node.Peers) != 0 {
		t.Errorf("Expected 0 peers, got %d", len(node.Peers))
	}
}

// TestPeerManagement tests adding and removing peers
func TestPeerManagement(t *testing.T) {
	node := NewNode("node-1", "Test Node", "192.168.1.1", "aa:bb:cc:dd:ee:ff")

	peer := &Peer{
		NodeID:      "node-2",
		IP:          "192.168.1.2",
		MAC:         "11:22:33:44:55:66",
		RSSI:        -50,
		HasInternet: false,
	}

	node.AddPeer(peer)
	if len(node.Peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(node.Peers))
	}

	retrievedPeer, exists := node.GetPeer("node-2")
	if !exists {
		t.Error("Expected peer to exist")
	}

	if retrievedPeer.NodeID != "node-2" {
		t.Errorf("Expected peer ID 'node-2', got '%s'", retrievedPeer.NodeID)
	}

	node.RemovePeer("node-2")
	if len(node.Peers) != 0 {
		t.Errorf("Expected 0 peers after removal, got %d", len(node.Peers))
	}
}

// TestPersonalNetworkCreation tests personal network creation
func TestPersonalNetworkCreation(t *testing.T) {
	pn := NewPersonalNetwork("pnet-1", "My Network", "user-1")

	if pn.ID != "pnet-1" {
		t.Errorf("Expected ID 'pnet-1', got '%s'", pn.ID)
	}

	if pn.Owner != "user-1" {
		t.Errorf("Expected owner 'user-1', got '%s'", pn.Owner)
	}

	if !pn.Policies.AllowInternet {
		t.Error("Expected AllowInternet to be true")
	}
}

// TestPersonalNetworkMembership tests adding and removing members
func TestPersonalNetworkMembership(t *testing.T) {
	pn := NewPersonalNetwork("pnet-1", "My Network", "user-1")

	member := &NetworkMember{
		NodeID:      "node-1",
		HasInternet: true,
		IsProxy:     true,
	}

	pn.AddMember(member)
	if len(pn.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(pn.Members))
	}

	if !pn.IsMember("node-1") {
		t.Error("Expected node-1 to be a member")
	}

	pn.RemoveMember("node-1")
	if pn.IsMember("node-1") {
		t.Error("Expected node-1 to not be a member after removal")
	}
}

// TestRoutingTable tests routing table operations
func TestRoutingTable(t *testing.T) {
	rt := NewRoutingTable()

	rt.AddRoute("dest-1", "next-1", 1, 100)

	route, exists := rt.GetRoute("dest-1")
	if !exists {
		t.Error("Expected route to exist")
	}

	if route.NextHop != "next-1" {
		t.Errorf("Expected next hop 'next-1', got '%s'", route.NextHop)
	}

	if route.HopCount != 1 {
		t.Errorf("Expected hop count 1, got %d", route.HopCount)
	}
}

// TestRouter tests the router
func TestRouter(t *testing.T) {
	router := NewRouter("node-1")

	router.RoutingTable.AddRoute("dest-1", "next-1", 1, 100)

	nextHop, found := router.RoutePacket("dest-1")
	if !found {
		t.Error("Expected route to be found")
	}

	if nextHop != "next-1" {
		t.Errorf("Expected next hop 'next-1', got '%s'", nextHop)
	}
}

// TestProxyManager tests proxy manager
func TestProxyManager(t *testing.T) {
	node := NewNode("node-1", "Test Node", "192.168.1.1", "aa:bb:cc:dd:ee:ff")
	pm := NewProxyManager(node)

	proxy := &Peer{
		NodeID:      "node-2",
		IP:          "192.168.1.2",
		HasInternet: true,
		RSSI:        -50,
	}

	pm.RegisterProxy(proxy)
	availableProxies := pm.GetAvailableProxies()
	if len(availableProxies) != 1 {
		t.Errorf("Expected 1 available proxy, got %d", len(availableProxies))
	}

	pm.UnregisterProxy("node-2")
	availableProxies = pm.GetAvailableProxies()
	if len(availableProxies) != 0 {
		t.Errorf("Expected 0 available proxies after unregistration, got %d", len(availableProxies))
	}
}
