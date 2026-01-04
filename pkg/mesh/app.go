package mesh

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MeshApp represents the main mesh application instance for mobile devices
type MeshApp struct {
	Node                   *Node
	Manager                *Manager
	Router                 *Router
	ProxyManager           *ProxyManager
	PersonalNetworkMgr     *PersonalNetworkManager
	Discovery              *Discovery
	Transport              *Transport
	InternetProxy          *InternetProxy
	InternetClient         *InternetClient
	IsConnected            bool
	IsInternetSharing      bool
	DiscoveredPeers        map[string]*Peer
	ctx                    context.Context
	cancel                 context.CancelFunc
	mu                     sync.RWMutex
	connectionListeners    []ConnectionListener
	peerDiscoveryListeners []PeerDiscoveryListener
}

// ConnectionListener is called when connection state changes
type ConnectionListener interface {
	OnConnectionStateChanged(connected bool)
	OnConnectionError(err error)
}

// PeerDiscoveryListener is called when peers are discovered
type PeerDiscoveryListener interface {
	OnPeerDiscovered(peer *Peer)
	OnPeerLost(peerID string)
}

// NetworkStats holds current network statistics
type NetworkStats struct {
	NodeID                 string
	PeerCount              int
	AvailableProxies       int
	InternetStatus         bool
	InternetSharingEnabled bool
	ConnectedNetworks      int
	DataTransferred        int64
	LastUpdate             time.Time
}

// NewMeshApp creates a new mesh application instance
func NewMeshApp(nodeID, nodeName, ip, mac string) *MeshApp {
	ctx, cancel := context.WithCancel(context.Background())

	node := NewNode(nodeID, nodeName, ip, mac)

	// Create networking components
	discovery := NewDiscovery(nodeID, nodeName, DefaultPort, false)
	transport := NewTransport(nodeID, DefaultPort)
	internetProxy := NewInternetProxy(nodeID, transport)
	internetClient := NewInternetClient(nodeID)

	return &MeshApp{
		Node:                   node,
		Manager:                NewManager(node),
		Router:                 NewRouter(nodeID),
		ProxyManager:           NewProxyManager(node),
		PersonalNetworkMgr:     NewPersonalNetworkManager(),
		Discovery:              discovery,
		Transport:              transport,
		InternetProxy:          internetProxy,
		InternetClient:         internetClient,
		IsConnected:            false,
		IsInternetSharing:      false,
		DiscoveredPeers:        make(map[string]*Peer),
		ctx:                    ctx,
		cancel:                 cancel,
		connectionListeners:    make([]ConnectionListener, 0),
		peerDiscoveryListeners: make([]PeerDiscoveryListener, 0),
	}
}

// Start initializes and starts the mesh application
func (ma *MeshApp) Start() error {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	// Check for internet connectivity
	hasInternet := CheckInternetConnectivity()
	ma.Node.HasInternet = hasInternet

	// Setup discovery callbacks
	ma.Discovery.SetCallbacks(
		func(peer *DiscoveredPeer) {
			ma.handlePeerDiscovered(peer)
		},
		func(peerID string) {
			ma.handlePeerLost(peerID)
		},
	)

	// Setup transport message handler
	ma.Transport.SetMessageHandler(func(peerID string, msg *Message) {
		ma.handleMessage(peerID, msg)
	})

	// Start transport layer
	if err := ma.Transport.Start(); err != nil {
		ma.notifyConnectionError(err)
		return fmt.Errorf("failed to start transport: %w", err)
	}

	// Start discovery
	if err := ma.Discovery.Start(); err != nil {
		ma.Transport.Stop()
		ma.notifyConnectionError(err)
		return fmt.Errorf("failed to start discovery: %w", err)
	}

	// Start manager
	if err := ma.Manager.Start(ma.ctx); err != nil {
		ma.Discovery.Stop()
		ma.Transport.Stop()
		ma.notifyConnectionError(err)
		return err
	}

	ma.IsConnected = true
	ma.notifyConnectionChanged(true)

	// Start background tasks
	go ma.internetCheckLoop()
	go ma.routingUpdateLoop()

	return nil
}

// Stop gracefully stops the mesh application
func (ma *MeshApp) Stop() {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	// Stop all networking components
	ma.Discovery.Stop()
	ma.Transport.Stop()
	ma.InternetProxy.Disable()
	ma.InternetClient.Disconnect()
	ma.Manager.Stop()

	ma.IsConnected = false
	ma.IsInternetSharing = false
	ma.cancel()
	// Reset context for restart
	ma.ctx, ma.cancel = context.WithCancel(context.Background())

	ma.notifyConnectionChanged(false)
}

// ConnectToNetwork attempts to connect to the mesh network
func (ma *MeshApp) ConnectToNetwork() error {
	return ma.Start()
}

// DisconnectFromNetwork disconnects from the mesh network
func (ma *MeshApp) DisconnectFromNetwork() {
	ma.Stop()
}

// EnableInternetSharing enables sharing of internet connection with mesh
func (ma *MeshApp) EnableInternetSharing() bool {
	ma.mu.Lock()
	hasInternet := ma.Node.HasInternet
	ma.mu.Unlock()

	if !hasInternet {
		return false
	}

	// Enable proxy server
	if err := ma.InternetProxy.Enable(); err != nil {
		return false
	}

	// Register as proxy in proxy manager
	proxyPeer := &Peer{
		NodeID:      ma.Node.ID,
		IP:          ma.Node.IP,
		MAC:         ma.Node.MAC,
		HasInternet: true,
		LastSeen:    time.Now().Unix(),
	}
	ma.ProxyManager.RegisterProxy(proxyPeer)

	// Update discovery to announce internet availability
	ma.Discovery.UpdateInternetStatus(true)

	ma.mu.Lock()
	ma.IsInternetSharing = true
	ma.mu.Unlock()

	return true
}

// DisableInternetSharing stops sharing internet connection
func (ma *MeshApp) DisableInternetSharing() {
	ma.InternetProxy.Disable()
	ma.ProxyManager.UnregisterProxy(ma.Node.ID)
	ma.Discovery.UpdateInternetStatus(false)

	ma.mu.Lock()
	ma.IsInternetSharing = false
	ma.mu.Unlock()
}

// RequestInternetAccess requests internet access from the mesh network
func (ma *MeshApp) RequestInternetAccess() bool {
	ma.mu.RLock()
	hasInternet := ma.Node.HasInternet
	ma.mu.RUnlock()

	if hasInternet {
		return true // Already have internet
	}

	// Find available proxy from discovered peers
	peers := ma.Discovery.GetPeers()
	var proxyPeer *DiscoveredPeer
	for _, peer := range peers {
		if peer.HasInternet {
			proxyPeer = peer
			break
		}
	}

	if proxyPeer == nil {
		return false // No proxy available
	}

	// Connect to peer first
	if err := ma.Transport.ConnectToPeer(proxyPeer.ID, proxyPeer.IP, proxyPeer.Port); err != nil {
		return false
	}

	// Connect to proxy
	if err := ma.InternetClient.ConnectToProxy(proxyPeer.ID, proxyPeer.IP, ProxyPort); err != nil {
		return false
	}

	// Send proxy request message
	msg := &Message{
		Type:      "proxy_request",
		Source:    ma.Node.ID,
		Dest:      proxyPeer.ID,
		Timestamp: time.Now(),
	}
	ma.Transport.SendMessage(proxyPeer.ID, msg)

	return true
}

// GetNetworkStats returns current network statistics
func (ma *MeshApp) GetNetworkStats() *NetworkStats {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	connectedPeers := ma.Transport.GetConnectedPeers()
	availableProxies := len(ma.Discovery.GetPeers())
	for _, peer := range ma.Discovery.GetPeers() {
		if !peer.HasInternet {
			availableProxies--
		}
	}

	return &NetworkStats{
		NodeID:                 ma.Node.ID,
		PeerCount:              len(connectedPeers),
		AvailableProxies:       availableProxies,
		InternetStatus:         ma.Node.HasInternet,
		InternetSharingEnabled: ma.IsInternetSharing,
		ConnectedNetworks:      ma.PersonalNetworkMgr.GetNetworkCount(),
		LastUpdate:             time.Now(),
	}
}

// GetConnectedPeers returns list of currently connected peers
func (ma *MeshApp) GetConnectedPeers() []string {
	return ma.Transport.GetConnectedPeers()
}

// GetAvailableProxies returns list of nodes offering internet access
func (ma *MeshApp) GetAvailableProxies() []string {
	peers := ma.Discovery.GetPeers()
	proxies := make([]string, 0)

	for _, peer := range peers {
		if peer.HasInternet {
			proxies = append(proxies, peer.ID)
		}
	}

	return proxies
}

// GetConnectionStatus returns whether the app is connected to the mesh
func (ma *MeshApp) GetConnectionStatus() bool {
	ma.mu.RLock()
	defer ma.mu.RUnlock()
	return ma.IsConnected
}

// GetInternetSharingStatus returns whether internet sharing is enabled
func (ma *MeshApp) GetInternetSharingStatus() bool {
	ma.mu.RLock()
	defer ma.mu.RUnlock()
	return ma.IsInternetSharing
}

// SetInternetStatus manually sets internet status (for testing)
func (ma *MeshApp) SetInternetStatus(hasInternet bool) {
	ma.mu.Lock()
	ma.Node.HasInternet = hasInternet
	ma.mu.Unlock()
	ma.Discovery.UpdateInternetStatus(hasInternet)
}

// ReleaseInternetAccess disconnects from the proxy
func (ma *MeshApp) ReleaseInternetAccess() error {
	ma.InternetClient.Disconnect()
	return nil
}

// AddConnectionListener adds a connection state listener (deprecated, use RegisterConnectionListener)
func (ma *MeshApp) AddConnectionListener(listener ConnectionListener) {
	ma.RegisterConnectionListener(listener)
}

// RegisterConnectionListener registers a connection state listener
func (ma *MeshApp) RegisterConnectionListener(listener ConnectionListener) {
	ma.mu.Lock()
	defer ma.mu.Unlock()
	ma.connectionListeners = append(ma.connectionListeners, listener)
}

// RegisterPeerDiscoveryListener registers a peer discovery listener
func (ma *MeshApp) RegisterPeerDiscoveryListener(listener PeerDiscoveryListener) {
	ma.mu.Lock()
	defer ma.mu.Unlock()
	ma.peerDiscoveryListeners = append(ma.peerDiscoveryListeners, listener)
}

// Internal methods

func (ma *MeshApp) handlePeerDiscovered(peer *DiscoveredPeer) {
	// Create Peer object
	meshPeer := &Peer{
		NodeID:      peer.ID,
		IP:          peer.IP,
		MAC:         peer.MAC,
		HasInternet: peer.HasInternet,
		LastSeen:    time.Now().Unix(),
	}

	ma.mu.Lock()
	ma.DiscoveredPeers[peer.ID] = meshPeer
	ma.mu.Unlock()

	// Try to connect to the peer
	if err := ma.Transport.ConnectToPeer(peer.ID, peer.IP, peer.Port); err == nil {
		// Add to router
		ma.Router.UpdateRoute(peer.ID, peer.ID, 1, 10*time.Millisecond)

		// Notify listeners
		ma.notifyPeerDiscovered(meshPeer)
	}

	// If peer has internet, register as proxy
	if peer.HasInternet {
		proxyPeer := &Peer{
			NodeID:      peer.ID,
			IP:          peer.IP,
			MAC:         peer.MAC,
			HasInternet: true,
			LastSeen:    time.Now().Unix(),
		}
		ma.ProxyManager.RegisterProxy(proxyPeer)
	}
}

func (ma *MeshApp) handlePeerLost(peerID string) {
	ma.mu.Lock()
	delete(ma.DiscoveredPeers, peerID)
	ma.mu.Unlock()

	// Disconnect from peer
	ma.Transport.DisconnectPeer(peerID)

	// Remove from router
	ma.Router.RemoveRoute(peerID)

	// Unregister proxy if applicable
	ma.ProxyManager.UnregisterProxy(peerID)

	// Notify listeners
	ma.notifyPeerLost(peerID)
}

func (ma *MeshApp) handleMessage(peerID string, msg *Message) {
	switch msg.Type {
	case "proxy_request":
		ma.handleProxyRequest(peerID, msg)
	case "proxy_response":
		ma.handleProxyResponse(peerID, msg)
	case "data":
		ma.handleDataMessage(peerID, msg)
	case "route_update":
		ma.handleRouteUpdate(peerID, msg)
	}
}

func (ma *MeshApp) handleProxyRequest(peerID string, msg *Message) {
	if !ma.InternetProxy.IsEnabled() {
		return
	}

	// Authorize the client
	ma.InternetProxy.AuthorizeClient(peerID)

	// Send response
	response := &Message{
		Type:      "proxy_response",
		Source:    ma.Node.ID,
		Dest:      peerID,
		Timestamp: time.Now(),
		Metadata:  map[string]string{"status": "authorized"},
	}
	ma.Transport.SendMessage(peerID, response)
}

func (ma *MeshApp) handleProxyResponse(peerID string, msg *Message) {
	if status, ok := msg.Metadata["status"]; ok && status == "authorized" {
		// Successfully authorized to use proxy
	}
}

func (ma *MeshApp) handleDataMessage(peerID string, msg *Message) {
	// Route the message if not for us
	if msg.Dest != ma.Node.ID {
		route := ma.Router.GetRoute(msg.Dest)
		if route != nil && route.NextHop != ma.Node.ID {
			ma.Transport.SendMessage(route.NextHop, msg)
		}
	}
}

func (ma *MeshApp) handleRouteUpdate(peerID string, msg *Message) {
	// Update routing table based on received information
}

func (ma *MeshApp) internetCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ma.ctx.Done():
			return
		case <-ticker.C:
			hasInternet := CheckInternetConnectivity()

			ma.mu.Lock()
			changed := ma.Node.HasInternet != hasInternet
			ma.Node.HasInternet = hasInternet
			ma.mu.Unlock()

			if changed {
				ma.Discovery.UpdateInternetStatus(hasInternet)

				if hasInternet && ma.IsInternetSharing {
					// Re-enable sharing if it was enabled
					ma.InternetProxy.Enable()
				} else if !hasInternet {
					// Disable sharing if we lost internet
					ma.InternetProxy.Disable()
				}
			}
		}
	}
}

func (ma *MeshApp) routingUpdateLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ma.ctx.Done():
			return
		case <-ticker.C:
			// Broadcast routing information to peers
			peers := ma.Transport.GetConnectedPeers()
			for _, peerID := range peers {
				msg := &Message{
					Type:      "route_update",
					Source:    ma.Node.ID,
					Dest:      peerID,
					Timestamp: time.Now(),
				}
				ma.Transport.SendMessage(peerID, msg)
			}
		}
	}
}

func (ma *MeshApp) notifyConnectionChanged(connected bool) {
	for _, listener := range ma.connectionListeners {
		listener.OnConnectionStateChanged(connected)
	}
}

func (ma *MeshApp) notifyConnectionError(err error) {
	for _, listener := range ma.connectionListeners {
		listener.OnConnectionError(err)
	}
}

func (ma *MeshApp) notifyPeerDiscovered(peer *Peer) {
	for _, listener := range ma.peerDiscoveryListeners {
		listener.OnPeerDiscovered(peer)
	}
}

func (ma *MeshApp) notifyPeerLost(peerID string) {
	for _, listener := range ma.peerDiscoveryListeners {
		listener.OnPeerLost(peerID)
	}
}
