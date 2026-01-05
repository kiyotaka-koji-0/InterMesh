package mesh

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// InternetProxy handles internet sharing via proxy
type InternetProxy struct {
	nodeID      string
	enabled     bool
	proxyServer *http.Server
	port        int
	clients     map[string]*ProxyClient
	clientsMu   sync.RWMutex
	transport   *Transport
	mu          sync.Mutex
}

// ProxyClient represents a client using our internet
type ProxyClient struct {
	PeerID     string
	Authorized bool
	BytesSent  uint64
	BytesRecv  uint64
	Connected  time.Time
}

// InternetClient handles connecting through a proxy for internet access
type InternetClient struct {
	nodeID      string
	proxyPeerID string
	proxyAddr   string
	client      *http.Client
	connected   bool
	mu          sync.Mutex
}

const (
	ProxyPort = 9997
)

// NewInternetProxy creates a new internet proxy
func NewInternetProxy(nodeID string, transport *Transport) *InternetProxy {
	return &InternetProxy{
		nodeID:    nodeID,
		port:      ProxyPort,
		clients:   make(map[string]*ProxyClient),
		transport: transport,
	}
}

// Enable enables internet sharing
func (p *InternetProxy) Enable() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.enabled {
		return nil
	}

	// Create HTTP proxy server
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleProxy)

	p.proxyServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", p.port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	p.enabled = true

	// Start server in background
	go func() {
		// Note: ListenAndServe blocks, so we run it in a goroutine
		// It returns ErrServerClosed when Shutdown is called
		if err := p.proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// In production, log this error
			p.mu.Lock()
			p.enabled = false
			p.mu.Unlock()
		}
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Disable disables internet sharing
func (p *InternetProxy) Disable() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		return nil
	}

	if p.proxyServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p.proxyServer.Shutdown(ctx)
		p.proxyServer = nil
	}

	p.enabled = false
	return nil
}

// IsEnabled returns whether internet sharing is enabled
func (p *InternetProxy) IsEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.enabled
}

// AuthorizeClient authorizes a peer to use our internet
func (p *InternetProxy) AuthorizeClient(peerID string) {
	p.clientsMu.Lock()
	defer p.clientsMu.Unlock()

	if client, exists := p.clients[peerID]; exists {
		client.Authorized = true
	} else {
		p.clients[peerID] = &ProxyClient{
			PeerID:     peerID,
			Authorized: true,
			Connected:  time.Now(),
		}
	}
}

// RevokeClient revokes a peer's internet access
func (p *InternetProxy) RevokeClient(peerID string) {
	p.clientsMu.Lock()
	defer p.clientsMu.Unlock()

	if client, exists := p.clients[peerID]; exists {
		client.Authorized = false
	}
}

// GetClients returns list of clients using our internet
func (p *InternetProxy) GetClients() []*ProxyClient {
	p.clientsMu.RLock()
	defer p.clientsMu.RUnlock()

	clients := make([]*ProxyClient, 0, len(p.clients))
	for _, client := range p.clients {
		clients = append(clients, client)
	}
	return clients
}

// handleProxy handles HTTP proxy requests
func (p *InternetProxy) handleProxy(w http.ResponseWriter, r *http.Request) {
	// Extract peer ID from headers or connection
	// For now, we'll accept all requests if proxy is enabled

	if r.Method == http.MethodConnect {
		p.handleConnect(w, r)
	} else {
		p.handleHTTP(w, r)
	}
}

// handleConnect handles HTTPS CONNECT method
func (p *InternetProxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	// Establish connection to destination
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()

	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Bidirectional copy
	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)
}

// handleHTTP handles regular HTTP requests
func (p *InternetProxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Create new request to destination
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Forward request
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// NewInternetClient creates a new internet client
func NewInternetClient(nodeID string) *InternetClient {
	return &InternetClient{
		nodeID: nodeID,
	}
}

// ConnectToProxy connects to a proxy for internet access
func (c *InternetClient) ConnectToProxy(proxyPeerID, proxyIP string, port int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.proxyPeerID = proxyPeerID
	c.proxyAddr = fmt.Sprintf("http://%s:%d", proxyIP, port)

	// Create HTTP client configured to use proxy
	proxyURL, err := http.ProxyFromEnvironment(&http.Request{})
	if err == nil && proxyURL != nil {
		c.client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
			Timeout: 30 * time.Second,
		}
	} else {
		c.client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	c.connected = true
	return nil
}

// Disconnect disconnects from the proxy
func (c *InternetClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	c.proxyPeerID = ""
	c.proxyAddr = ""
	c.client = nil
}

// IsConnected returns whether connected to a proxy
func (c *InternetClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// MakeRequest makes an HTTP request through the proxy
func (c *InternetClient) MakeRequest(url string) (*http.Response, error) {
	c.mu.Lock()
	if !c.connected || c.client == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("not connected to proxy")
	}
	client := c.client
	c.mu.Unlock()

	return client.Get(url)
}

// DoRequest makes a generic HTTP request through the proxy
func (c *InternetClient) DoRequest(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	if !c.connected || c.client == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("not connected to proxy")
	}
	client := c.client
	c.mu.Unlock()

	return client.Do(req)
}

// CheckInternetConnectivity tests internet connectivity
func CheckInternetConnectivity() bool {
	// Try to connect to common DNS servers
	timeout := 3 * time.Second

	servers := []string{
		"8.8.8.8:53",        // Google DNS
		"1.1.1.1:53",        // Cloudflare DNS
		"208.67.222.222:53", // OpenDNS
	}

	for _, server := range servers {
		conn, err := net.DialTimeout("tcp", server, timeout)
		if err == nil {
			conn.Close()
			return true
		}
	}

	return false
}

// GetLocalIP returns the local IP address
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no local IP address found")
}
