package mesh

import (
	"sync"
	"time"
)

// ProxyConnection represents a connection through a proxy
type ProxyConnection struct {
	ClientID         string
	ProxyID          string
	EstablishedAt    time.Time
	BytesTransferred int64
	LastActivity     time.Time
}

// ProxyManager manages proxy connections and internet sharing
type ProxyManager struct {
	Node        *Node
	Proxies     map[string]*Peer            // Available proxy peers
	Connections map[string]*ProxyConnection // Active proxy connections
	mu          sync.RWMutex
}

// NewProxyManager creates a new proxy manager
func NewProxyManager(node *Node) *ProxyManager {
	return &ProxyManager{
		Node:        node,
		Proxies:     make(map[string]*Peer),
		Connections: make(map[string]*ProxyConnection),
	}
}

// RegisterProxy registers a peer as an available proxy
func (pm *ProxyManager) RegisterProxy(peer *Peer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.Proxies[peer.NodeID] = peer
}

// UnregisterProxy unregisters a proxy peer
func (pm *ProxyManager) UnregisterProxy(peerID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.Proxies, peerID)
}

// GetAvailableProxies returns all available proxy peers
func (pm *ProxyManager) GetAvailableProxies() []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	proxies := make([]*Peer, 0, len(pm.Proxies))
	for _, proxy := range pm.Proxies {
		if proxy.HasInternet {
			proxies = append(proxies, proxy)
		}
	}
	return proxies
}

// CreateProxyConnection creates a new proxy connection
func (pm *ProxyManager) CreateProxyConnection(clientID, proxyID string) (*ProxyConnection, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Verify proxy exists and has internet
	proxy, exists := pm.Proxies[proxyID]
	if !exists || !proxy.HasInternet {
		return nil, ErrProxyNotAvailable
	}

	connection := &ProxyConnection{
		ClientID:      clientID,
		ProxyID:       proxyID,
		EstablishedAt: time.Now(),
		LastActivity:  time.Now(),
	}

	connID := clientID + "-" + proxyID
	pm.Connections[connID] = connection
	return connection, nil
}

// CloseProxyConnection closes an existing proxy connection
func (pm *ProxyManager) CloseProxyConnection(clientID, proxyID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	connID := clientID + "-" + proxyID
	delete(pm.Connections, connID)
}

// GetProxyConnection retrieves a proxy connection
func (pm *ProxyManager) GetProxyConnection(clientID, proxyID string) (*ProxyConnection, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	connID := clientID + "-" + proxyID
	conn, exists := pm.Connections[connID]
	return conn, exists
}

// GetClientConnections returns all active connections for a client
func (pm *ProxyManager) GetClientConnections(clientID string) []*ProxyConnection {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	connections := make([]*ProxyConnection, 0)
	for _, conn := range pm.Connections {
		if conn.ClientID == clientID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// UpdateProxyActivity updates the last activity time for a connection
func (pm *ProxyManager) UpdateProxyActivity(clientID, proxyID string, bytesTransferred int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	connID := clientID + "-" + proxyID
	if conn, exists := pm.Connections[connID]; exists {
		conn.LastActivity = time.Now()
		conn.BytesTransferred += bytesTransferred
	}
}

// SelectBestProxy selects the best available proxy for a client
func (pm *ProxyManager) SelectBestProxy() (*Peer, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var bestProxy *Peer
	var bestSignal int = -150 // Worse than any real RSSI

	for _, proxy := range pm.Proxies {
		if proxy.HasInternet && proxy.RSSI > bestSignal {
			bestProxy = proxy
			bestSignal = proxy.RSSI
		}
	}

	if bestProxy == nil {
		return nil, ErrNoAvailableProxy
	}
	return bestProxy, nil
}

// ProxyStatistics tracks statistics for a proxy
type ProxyStatistics struct {
	ProxyID               string
	ActiveConnections     int
	TotalBytesTransferred int64
	UpTime                time.Duration
}

// GetProxyStatistics returns statistics for a specific proxy
func (pm *ProxyManager) GetProxyStatistics(proxyID string) *ProxyStatistics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	proxy, exists := pm.Proxies[proxyID]
	if !exists {
		return nil
	}

	activeConnections := 0
	totalBytes := int64(0)

	for _, conn := range pm.Connections {
		if conn.ProxyID == proxyID {
			activeConnections++
			totalBytes += conn.BytesTransferred
		}
	}

	upTime := time.Since(time.Unix(proxy.LastSeen/1000, 0))

	return &ProxyStatistics{
		ProxyID:               proxyID,
		ActiveConnections:     activeConnections,
		TotalBytesTransferred: totalBytes,
		UpTime:                upTime,
	}
}

// Custom errors
var (
	ErrProxyNotAvailable = NewMeshError("proxy not available")
	ErrNoAvailableProxy  = NewMeshError("no available proxy")
)

// MeshError represents a mesh network error
type MeshError struct {
	Message string
}

// NewMeshError creates a new mesh error
func NewMeshError(message string) *MeshError {
	return &MeshError{Message: message}
}

// Error implements the error interface
func (e *MeshError) Error() string {
	return e.Message
}
