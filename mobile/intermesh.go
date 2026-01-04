// Package intermesh provides mobile bindings for the InterMesh mesh networking library
package intermesh

import (
	"context"
	"fmt"

	"github.com/kiyotaka-koji-0/intermesh/pkg/mesh"
)

// MobileApp is the main mobile application interface
type MobileApp struct {
	app *mesh.MeshApp
}

// MobileConnectionListener implements ConnectionListener for mobile callbacks
type MobileConnectionListener struct {
	onConnected func()
	onError     func(string)
}

// OnConnectionStateChanged is called when connection state changes
func (mcl *MobileConnectionListener) OnConnectionStateChanged(connected bool) {
	if connected && mcl.onConnected != nil {
		mcl.onConnected()
	}
}

// OnConnectionError is called on connection error
func (mcl *MobileConnectionListener) OnConnectionError(err error) {
	if mcl.onError != nil {
		mcl.onError(err.Error())
	}
}

// MobilePeerDiscoveryListener implements PeerDiscoveryListener for mobile callbacks
type MobilePeerDiscoveryListener struct {
	onPeerDiscovered func(string, string) // peerID, IP
	onPeerLost       func(string)         // peerID
}

// OnPeerDiscovered is called when a peer is discovered
func (mpdl *MobilePeerDiscoveryListener) OnPeerDiscovered(peer *mesh.Peer) {
	if mpdl.onPeerDiscovered != nil {
		mpdl.onPeerDiscovered(peer.NodeID, peer.IP)
	}
}

// OnPeerLost is called when a peer is lost
func (mpdl *MobilePeerDiscoveryListener) OnPeerLost(peerID string) {
	if mpdl.onPeerLost != nil {
		mpdl.onPeerLost(peerID)
	}
}

// NewMobileApp creates a new mobile application instance
func NewMobileApp(nodeID, nodeName, ip, mac string) *MobileApp {
	return &MobileApp{
		app: mesh.NewMeshApp(nodeID, nodeName, ip, mac),
	}
}

// Start initializes and starts the mesh application
func (ma *MobileApp) Start() error {
	return ma.app.Start()
}

// Stop gracefully stops the mesh application
func (ma *MobileApp) Stop() {
	ma.app.Stop()
}

// ConnectToNetwork attempts to connect to the mesh network
func (ma *MobileApp) ConnectToNetwork() error {
	return ma.app.ConnectToNetwork()
}

// DisconnectFromNetwork disconnects from the mesh network
func (ma *MobileApp) DisconnectFromNetwork() {
	ma.app.DisconnectFromNetwork()
}

// IsConnected returns whether the app is connected to the mesh
func (ma *MobileApp) IsConnected() bool {
	return ma.app.GetConnectionStatus()
}

// EnableInternetSharing enables the device to share its internet
func (ma *MobileApp) EnableInternetSharing() error {
	success := ma.app.EnableInternetSharing()
	if !success {
		return fmt.Errorf("failed to enable internet sharing")
	}
	return nil
}

// DisableInternetSharing disables internet sharing
func (ma *MobileApp) DisableInternetSharing() {
	ma.app.DisableInternetSharing()
}

// IsInternetSharingEnabled returns whether internet sharing is enabled
func (ma *MobileApp) IsInternetSharingEnabled() bool {
	return ma.app.GetInternetSharingStatus()
}

// SetInternetStatus updates the device's internet connectivity status
func (ma *MobileApp) SetInternetStatus(hasInternet bool) {
	ma.app.SetInternetStatus(hasInternet)
}

// HasInternet returns whether the device has internet connectivity
func (ma *MobileApp) HasInternet() bool {
	return ma.app.Node.HasInternet
}

// RequestInternetAccess requests internet access through the mesh network
func (ma *MobileApp) RequestInternetAccess() (string, error) {
	success := ma.app.RequestInternetAccess()
	if success {
		return "Connected to proxy", nil
	}
	return "", fmt.Errorf("no proxies available")
}

// ReleaseInternetAccess releases internet access from a proxy
func (ma *MobileApp) ReleaseInternetAccess(proxyID string) {
	ma.app.ReleaseInternetAccess()
}

// GetNetworkStats returns current network statistics
func (ma *MobileApp) GetNetworkStats() *MobileNetworkStats {
	stats := ma.app.GetNetworkStats()
	return &MobileNetworkStats{
		NodeID:                 stats.NodeID,
		PeerCount:              int64(stats.PeerCount),
		AvailableProxies:       int64(stats.AvailableProxies),
		InternetStatus:         stats.InternetStatus,
		InternetSharingEnabled: stats.InternetSharingEnabled,
		ConnectedNetworks:      int64(stats.ConnectedNetworks),
		DataTransferred:        stats.DataTransferred,
	}
}

// MobileNetworkStats holds network statistics for mobile
type MobileNetworkStats struct {
	NodeID                 string
	PeerCount              int64
	AvailableProxies       int64
	InternetStatus         bool
	InternetSharingEnabled bool
	ConnectedNetworks      int64
	DataTransferred        int64
}

// GetAvailableProxyCount returns the number of available proxies
func (ma *MobileApp) GetAvailableProxyCount() int64 {
	proxies := ma.app.GetAvailableProxies()
	return int64(len(proxies))
}

// GetConnectedPeerCount returns the number of connected peers
func (ma *MobileApp) GetConnectedPeerCount() int64 {
	peers := ma.app.GetConnectedPeers()
	return int64(len(peers))
}

// GetNodeID returns the node's ID
func (ma *MobileApp) GetNodeID() string {
	return ma.app.Node.ID
}

// GetNodeName returns the node's name
func (ma *MobileApp) GetNodeName() string {
	return ma.app.Node.Name
}

// SetNodeName updates the node's name
func (ma *MobileApp) SetNodeName(name string) {
	ma.app.Node.Name = name
}

// GetNodeIP returns the node's IP address
func (ma *MobileApp) GetNodeIP() string {
	return ma.app.Node.IP
}

// SetNodeIP updates the node's IP address
func (ma *MobileApp) SetNodeIP(ip string) {
	ma.app.Node.IP = ip
}

// MobileNode wraps mesh.Node for mobile platforms
type MobileNode struct {
	node *mesh.Node
}

// NewMobileNode creates a new mobile node wrapper
func NewMobileNode(id, name, ip, mac string) *MobileNode {
	return &MobileNode{
		node: mesh.NewNode(id, name, ip, mac),
	}
}

// GetID returns the node ID
func (mn *MobileNode) GetID() string {
	return mn.node.ID
}

// GetName returns the node name
func (mn *MobileNode) GetName() string {
	return mn.node.Name
}

// GetIP returns the node IP address
func (mn *MobileNode) GetIP() string {
	return mn.node.IP
}

// SetInternetStatus sets the internet connectivity status
func (mn *MobileNode) SetInternetStatus(hasInternet bool) {
	mn.node.SetInternetStatus(hasInternet)
}

// GetInternetStatus returns the internet connectivity status
func (mn *MobileNode) GetInternetStatus() bool {
	return mn.node.GetInternetStatus()
}

// GetPeerCount returns the number of connected peers
func (mn *MobileNode) GetPeerCount() int {
	return len(mn.node.Peers)
}

// MobileMeshManager wraps mesh.Manager for mobile platforms
type MobileMeshManager struct {
	manager *mesh.Manager
}

// NewMobileMeshManager creates a new mobile mesh manager
func NewMobileMeshManager(node *MobileNode) *MobileMeshManager {
	return &MobileMeshManager{
		manager: mesh.NewManager(node.node),
	}
}

// Start starts the mesh manager
func (mmm *MobileMeshManager) Start() error {
	return mmm.manager.Start(context.Background())
}

// Stop stops the mesh manager
func (mmm *MobileMeshManager) Stop() {
	mmm.manager.Stop()
}

// MobileProxyManager wraps mesh.ProxyManager for mobile platforms
type MobileProxyManager struct {
	proxyManager *mesh.ProxyManager
}

// NewMobileProxyManager creates a new mobile proxy manager
func NewMobileProxyManager(node *MobileNode) *MobileProxyManager {
	return &MobileProxyManager{
		proxyManager: mesh.NewProxyManager(node.node),
	}
}

// GetAvailableProxyCount returns the number of available proxies
func (mpm *MobileProxyManager) GetAvailableProxyCount() int {
	return len(mpm.proxyManager.GetAvailableProxies())
}

// SelectBestProxy selects the best available proxy
func (mpm *MobileProxyManager) SelectBestProxy() string {
	proxy, err := mpm.proxyManager.SelectBestProxy()
	if err != nil {
		return ""
	}
	return proxy.NodeID
}

// CreateProxyConnection creates a new proxy connection
func (mpm *MobileProxyManager) CreateProxyConnection(clientID, proxyID string) bool {
	_, err := mpm.proxyManager.CreateProxyConnection(clientID, proxyID)
	return err == nil
}

// CloseProxyConnection closes a proxy connection
func (mpm *MobileProxyManager) CloseProxyConnection(clientID, proxyID string) {
	mpm.proxyManager.CloseProxyConnection(clientID, proxyID)
}

// MobilePersonalNetworkManager wraps mesh.PersonalNetworkManager for mobile platforms
type MobilePersonalNetworkManager struct {
	pnManager *mesh.PersonalNetworkManager
}

// NewMobilePersonalNetworkManager creates a new mobile personal network manager
func NewMobilePersonalNetworkManager() *MobilePersonalNetworkManager {
	return &MobilePersonalNetworkManager{
		pnManager: mesh.NewPersonalNetworkManager(),
	}
}

// CreateNetwork creates a new personal network
func (mpnm *MobilePersonalNetworkManager) CreateNetwork(id, name, owner string) string {
	network := mpnm.pnManager.CreateNetwork(id, name, owner)
	return network.ID
}

// DeleteNetwork deletes a personal network
func (mpnm *MobilePersonalNetworkManager) DeleteNetwork(id string) {
	mpnm.pnManager.DeleteNetwork(id)
}

// GetNetworkCount returns the number of networks owned by a user
func (mpnm *MobilePersonalNetworkManager) GetNetworkCount(owner string) int {
	return len(mpnm.pnManager.GetNetworksByOwner(owner))
}
