// Package intermesh provides mobile bindings for the InterMesh mesh networking library
package intermesh

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kiyotaka-koji-0/intermesh/pkg/mesh"
)

// MobileApp is the main mobile application interface
type MobileApp struct {
	app             *mesh.MeshApp
	bleProxyHandler *BLEProxyHandler
	httpProxy       *HTTPProxyServer
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
	mobileApp := &MobileApp{
		app: mesh.NewMeshApp(nodeID, nodeName, ip, mac),
	}
	mobileApp.bleProxyHandler = NewBLEProxyHandler(nodeID, mobileApp)
	mobileApp.httpProxy = NewHTTPProxyServer(mobileApp)
	return mobileApp
}

// Start initializes and starts the mesh application
func (ma *MobileApp) Start() error {
	return ma.app.Start()
}

// Stop gracefully stops the mesh application
func (ma *MobileApp) Stop() {
	if ma.httpProxy != nil {
		ma.httpProxy.Stop()
	}
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

// HasAnyInternet returns whether the device has any internet (direct or mesh)
func (ma *MobileApp) HasAnyInternet() bool {
	return ma.app.Node.HasInternet || ma.app.InternetClient.IsConnected()
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

// RegisterBLEProxy registers a BLE peer as an available proxy
func (ma *MobileApp) RegisterBLEProxy(peerID, peerIP, peerMAC string, hasInternet bool) {
	peer := &mesh.Peer{
		NodeID:      peerID,
		IP:          peerIP,
		MAC:         peerMAC,
		HasInternet: hasInternet,
		LastSeen:    time.Now().Unix(),
		RSSI:        -50, // Reasonable BLE RSSI value
	}

	// Add to node's peers so it's visible to discovery
	ma.app.Node.AddPeer(peer)

	if hasInternet {
		ma.app.ProxyManager.RegisterProxy(peer)
	}
}

// UnregisterBLEProxy unregisters a BLE peer as proxy
func (ma *MobileApp) UnregisterBLEProxy(peerID string) {
	ma.app.ProxyManager.UnregisterProxy(peerID)
}

// HandleBLEProxyMessage handles incoming BLE proxy messages
func (ma *MobileApp) HandleBLEProxyMessage(senderID string, data []byte) error {
	return ma.bleProxyHandler.HandleBLEProxyMessage(senderID, data)
}

// SetBLEMessageSender sets the callback for sending BLE messages
func (ma *MobileApp) SetBLEMessageSender(sender func(peerID string, messageType string, data []byte) error) {
	ma.bleProxyHandler.SetBLEMessageSender(sender)
}

// BLEMessageCallback is a simple callback interface for mobile platforms
type BLEMessageCallback interface {
	SendMessage(message string) bool
}

// SetSimpleBLEMessageSender sets a simple callback for sending BLE messages
// This is easier to use from iOS/Android than the full SetBLEMessageSender
func (ma *MobileApp) SetSimpleBLEMessageSender(callback BLEMessageCallback) {
	if callback == nil {
		ma.bleProxyHandler.SetBLEMessageSender(nil)
		return
	}

	// Wrap the simple callback in the full interface
	ma.bleProxyHandler.SetBLEMessageSender(func(peerID string, messageType string, data []byte) error {
		// For HTTP tunnel, the data is already the JSON message
		message := string(data)
		if callback.SendMessage(message) {
			return nil
		}
		return fmt.Errorf("failed to send BLE message")
	})
}

// RequestInternetThroughBLE requests internet access through a BLE proxy
func (ma *MobileApp) RequestInternetThroughBLE(proxyPeerID, url, method string, headers map[string]string, body string) (string, error) {
	return ma.bleProxyHandler.SendProxyRequest(proxyPeerID, url, method, headers, []byte(body))
}

// SimpleHTTPRequestThroughBLE performs a simple GET request through BLE proxy
// This is a simplified version for mobile that doesn't require map parameters
func (ma *MobileApp) SimpleHTTPRequestThroughBLE(url string) (string, error) {
	// Find any available BLE proxy from ProxyManager (includes BLE-registered proxies)
	proxies := ma.app.ProxyManager.GetAvailableProxies()
	if len(proxies) == 0 {
		return "", fmt.Errorf("no BLE proxies available")
	}

	proxyPeer := proxies[0]
	return ma.bleProxyHandler.SendProxyRequest(proxyPeer.NodeID, url, "GET", nil, nil)
}

// CreateProxyRequest creates a JSON proxy request for sending via BLE
// Returns a JSON string that should be sent to the proxy device via BLE
func (ma *MobileApp) CreateProxyRequest(url, method string) (string, error) {
	return ma.bleProxyHandler.CreateProxyRequest(url, method)
}

// ExecuteProxyRequest executes an HTTP proxy request (for devices with internet)
// Takes a JSON request string and returns a JSON response string
func (ma *MobileApp) ExecuteProxyRequest(requestJSON string) (string, error) {
	if !ma.HasInternet() {
		return "", fmt.Errorf("this device does not have internet access")
	}
	return ma.bleProxyHandler.ExecuteProxyRequestSync(requestJSON)
}

// ParseProxyResponseBody extracts just the body from a proxy response JSON
func (ma *MobileApp) ParseProxyResponseBody(responseJSON string) string {
	return ma.bleProxyHandler.ParseProxyResponseBody(responseJSON)
}

// StartHTTPProxy starts the local HTTP proxy server on the specified port
// Configure your browser/apps to use 127.0.0.1:<port> as HTTP proxy
func (ma *MobileApp) StartHTTPProxy(port int64) error {
	return ma.httpProxy.Start(int(port))
}

// StopHTTPProxy stops the local HTTP proxy server
func (ma *MobileApp) StopHTTPProxy() {
	ma.httpProxy.Stop()
}

// IsHTTPProxyRunning returns whether the HTTP proxy is running
func (ma *MobileApp) IsHTTPProxyRunning() bool {
	return ma.httpProxy.IsRunning()
}

// GetHTTPProxyPort returns the HTTP proxy port
func (ma *MobileApp) GetHTTPProxyPort() int64 {
	return int64(ma.httpProxy.GetPort())
}

// HandleTunnelResponse handles incoming tunnel response from BLE (for proxy client)
func (ma *MobileApp) HandleTunnelResponse(responseJSON string) error {
	return ma.httpProxy.HandleTunnelResponse(responseJSON)
}

// ExecuteTunnelRequest executes a tunnel request (for device with internet)
func (ma *MobileApp) ExecuteTunnelRequest(requestJSON string) (string, error) {
	return ma.executeTunnelRequestInternal(requestJSON)
}

func (ma *MobileApp) executeTunnelRequestInternal(reqJSON string) (string, error) {
	var req TunnelRequest
	err := json.Unmarshal([]byte(reqJSON), &req)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// 1. Try local internet first
	if ma.HasInternet() {
		// Handle TUNNEL method for HTTPS
		if req.Method == "TUNNEL" {
			return executeTunnelData(&req)
		}
		// Execute regular HTTP request
		return executeHTTPTunnel(&req)
	}

	// 2. If no local internet, try to relay through the mesh internet client
	if ma.app.InternetClient.IsConnected() {
		return ma.relayTunnelToMesh(&req)
	}

	return createErrorResponse(req.ID, "No internet access or mesh proxy available")
}

func (ma *MobileApp) relayTunnelToMesh(req *TunnelRequest) (string, error) {
	// Decode body
	var body io.Reader
	if req.Body != "" {
		bodyData, err := base64.StdEncoding.DecodeString(req.Body)
		if err == nil && len(bodyData) > 0 {
			body = bytes.NewReader(bodyData)
		}
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Invalid bridged request: %v", err))
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute through Mesh InternetClient
	resp, err := ma.app.InternetClient.DoRequest(httpReq)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Mesh tunnel relay failed: %v", err))
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Failed to read mesh tunnel response: %v", err))
	}

	// Create response
	tunnelResp := &TunnelResponse{
		ID:         req.ID,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    make(map[string]string),
		Body:       base64.StdEncoding.EncodeToString(respBody),
	}

	// Copy headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			tunnelResp.Headers[key] = values[0]
		}
	}

	respJSON, err := json.Marshal(tunnelResp)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Failed to marshal tunnel response: %v", err))
	}

	return string(respJSON), nil
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
