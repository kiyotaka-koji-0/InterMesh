package intermesh

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPProxyServer runs a local HTTP proxy that tunnels through BLE
type HTTPProxyServer struct {
	listener       net.Listener
	port           int
	isRunning      bool
	mu             sync.RWMutex
	mobileApp      *MobileApp
	activeConns    map[string]net.Conn
	connMu         sync.RWMutex
	pendingReqs    map[string]chan *TunnelResponse
	pendingMu      sync.RWMutex
	onStatusChange func(running bool, port int)
}

// TunnelRequest represents a request to tunnel through BLE
type TunnelRequest struct {
	ID      string            `json:"id"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"` // Base64 encoded
}

// TunnelResponse represents a response from the tunnel
type TunnelResponse struct {
	ID         string            `json:"id"`
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"` // Base64 encoded
	Error      string            `json:"error,omitempty"`
}

// NewHTTPProxyServer creates a new HTTP proxy server
func NewHTTPProxyServer(mobileApp *MobileApp) *HTTPProxyServer {
	return &HTTPProxyServer{
		mobileApp:   mobileApp,
		activeConns: make(map[string]net.Conn),
		pendingReqs: make(map[string]chan *TunnelResponse),
	}
}

// Start starts the HTTP proxy server on the given port
func (p *HTTPProxyServer) Start(port int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return fmt.Errorf("proxy server already running")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	p.listener = listener
	p.port = port
	p.isRunning = true

	go p.acceptConnections()

	if p.onStatusChange != nil {
		p.onStatusChange(true, port)
	}

	return nil
}

// Stop stops the HTTP proxy server
func (p *HTTPProxyServer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isRunning {
		return
	}

	p.isRunning = false
	if p.listener != nil {
		p.listener.Close()
	}

	// Close all active connections
	p.connMu.Lock()
	for _, conn := range p.activeConns {
		conn.Close()
	}
	p.activeConns = make(map[string]net.Conn)
	p.connMu.Unlock()

	if p.onStatusChange != nil {
		p.onStatusChange(false, 0)
	}
}

// IsRunning returns whether the proxy is running
func (p *HTTPProxyServer) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isRunning
}

// GetPort returns the proxy port
func (p *HTTPProxyServer) GetPort() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.port
}

// SetStatusCallback sets the callback for status changes
func (p *HTTPProxyServer) SetStatusCallback(callback func(running bool, port int)) {
	p.onStatusChange = callback
}

func (p *HTTPProxyServer) acceptConnections() {
	for {
		p.mu.RLock()
		running := p.isRunning
		p.mu.RUnlock()

		if !running {
			return
		}

		conn, err := p.listener.Accept()
		if err != nil {
			if !p.isRunning {
				return
			}
			continue
		}

		go p.handleConnection(conn)
	}
}

func (p *HTTPProxyServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	connID := fmt.Sprintf("%d", time.Now().UnixNano())
	p.connMu.Lock()
	p.activeConns[connID] = conn
	p.connMu.Unlock()

	defer func() {
		p.connMu.Lock()
		delete(p.activeConns, connID)
		p.connMu.Unlock()
	}()

	reader := bufio.NewReader(conn)

	// Read the HTTP request
	req, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	// Handle CONNECT method for HTTPS
	if req.Method == "CONNECT" {
		p.handleConnect(conn, req, connID)
		return
	}

	// Handle regular HTTP request
	p.handleHTTPRequest(conn, req, connID)
}

func (p *HTTPProxyServer) handleHTTPRequest(conn net.Conn, req *http.Request, connID string) {
	// Read body
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

	// Build full URL
	url := req.URL.String()
	if !strings.HasPrefix(url, "http") {
		url = "http://" + req.Host + url
	}

	// Create tunnel request
	tunnelReq := &TunnelRequest{
		ID:      fmt.Sprintf("%s-%d", connID, time.Now().UnixNano()),
		Method:  req.Method,
		URL:     url,
		Headers: make(map[string]string),
		Body:    base64.StdEncoding.EncodeToString(body),
	}

	// Copy headers
	for key, values := range req.Header {
		if len(values) > 0 {
			tunnelReq.Headers[key] = values[0]
		}
	}

	// Send through BLE and wait for response
	resp, err := p.sendThroughBLE(tunnelReq)
	if err != nil {
		// Send error response
		errorResp := fmt.Sprintf("HTTP/1.1 502 Bad Gateway\r\nContent-Type: text/plain\r\n\r\nProxy Error: %s", err.Error())
		conn.Write([]byte(errorResp))
		return
	}

	// Write response
	p.writeHTTPResponse(conn, resp)
}

func (p *HTTPProxyServer) handleConnect(conn net.Conn, req *http.Request, connID string) {
	// For HTTPS CONNECT, we need to establish a tunnel
	// Send 200 OK to client
	conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Now we need to tunnel all data through BLE
	// This is more complex - we'll use a simple chunked approach
	p.tunnelHTTPS(conn, req.Host, connID)
}

func (p *HTTPProxyServer) tunnelHTTPS(clientConn net.Conn, host string, connID string) {
	// For HTTPS tunneling, we create a persistent connection through BLE
	// This is challenging due to BLE's packet-based nature

	// Read data from client in chunks and forward through BLE
	buffer := make([]byte, 32*1024) // 32KB chunks

	for {
		clientConn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := clientConn.Read(buffer)
		if err != nil {
			return
		}

		if n > 0 {
			// Create tunnel request for raw data
			tunnelReq := &TunnelRequest{
				ID:     fmt.Sprintf("%s-%d", connID, time.Now().UnixNano()),
				Method: "TUNNEL",
				URL:    host,
				Body:   base64.StdEncoding.EncodeToString(buffer[:n]),
			}

			resp, err := p.sendThroughBLE(tunnelReq)
			if err != nil {
				return
			}

			// Write response data back to client
			if resp.Body != "" {
				data, err := base64.StdEncoding.DecodeString(resp.Body)
				if err == nil && len(data) > 0 {
					clientConn.Write(data)
				}
			}
		}
	}
}

func (p *HTTPProxyServer) sendThroughBLE(req *TunnelRequest) (*TunnelResponse, error) {
	// Create response channel
	respChan := make(chan *TunnelResponse, 1)

	p.pendingMu.Lock()
	p.pendingReqs[req.ID] = respChan
	p.pendingMu.Unlock()

	defer func() {
		p.pendingMu.Lock()
		delete(p.pendingReqs, req.ID)
		p.pendingMu.Unlock()
	}()

	// Serialize request
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send through BLE
	if p.mobileApp.bleProxyHandler.onBLEMessage == nil {
		return nil, fmt.Errorf("BLE not connected")
	}

	// Find a proxy peer
	proxies := p.mobileApp.app.ProxyManager.GetAvailableProxies()
	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxy available")
	}

	proxyPeer := proxies[0]
	err = p.mobileApp.bleProxyHandler.onBLEMessage(proxyPeer.NodeID, "http_tunnel", reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to send BLE message: %w", err)
	}

	// Wait for response with timeout
	select {
	case resp := <-respChan:
		if resp.Error != "" {
			return nil, fmt.Errorf("%s", resp.Error)
		}
		return resp, nil
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("request timeout")
	}
}

// HandleTunnelResponse handles a tunnel response from BLE
func (p *HTTPProxyServer) HandleTunnelResponse(respJSON string) error {
	var resp TunnelResponse
	err := json.Unmarshal([]byte(respJSON), &resp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	p.pendingMu.RLock()
	respChan, ok := p.pendingReqs[resp.ID]
	p.pendingMu.RUnlock()

	if ok {
		select {
		case respChan <- &resp:
		default:
		}
	}

	return nil
}

func (p *HTTPProxyServer) writeHTTPResponse(conn net.Conn, resp *TunnelResponse) {
	// Build HTTP response
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, resp.Status)
	conn.Write([]byte(statusLine))

	// Write headers
	for key, value := range resp.Headers {
		header := fmt.Sprintf("%s: %s\r\n", key, value)
		conn.Write([]byte(header))
	}
	conn.Write([]byte("\r\n"))

	// Write body
	if resp.Body != "" {
		body, err := base64.StdEncoding.DecodeString(resp.Body)
		if err == nil {
			conn.Write(body)
		}
	}
}

// ExecuteTunnelRequest executes a tunnel request (for the device with internet)
func ExecuteTunnelRequest(reqJSON string) (string, error) {
	var req TunnelRequest
	err := json.Unmarshal([]byte(reqJSON), &req)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Handle TUNNEL method for HTTPS
	if req.Method == "TUNNEL" {
		return executeTunnelData(&req)
	}

	// Execute regular HTTP request
	return executeHTTPTunnel(&req)
}

func executeHTTPTunnel(req *TunnelRequest) (string, error) {
	// Decode body
	var body io.Reader
	if req.Body != "" {
		bodyData, err := base64.StdEncoding.DecodeString(req.Body)
		if err == nil && len(bodyData) > 0 {
			body = strings.NewReader(string(bodyData))
		}
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Invalid request: %v", err))
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Request failed: %v", err))
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Failed to read response: %v", err))
	}

	// Create response
	resp := &TunnelResponse{
		ID:         req.ID,
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    make(map[string]string),
		Body:       base64.StdEncoding.EncodeToString(respBody),
	}

	// Copy headers
	for key, values := range httpResp.Header {
		if len(values) > 0 {
			resp.Headers[key] = values[0]
		}
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Failed to marshal response: %v", err))
	}

	return string(respJSON), nil
}

// Persistent HTTPS tunnel connections
var httpsConnections = make(map[string]net.Conn)
var httpsConnMu sync.RWMutex

func executeTunnelData(req *TunnelRequest) (string, error) {
	// Get or create connection to target host
	httpsConnMu.Lock()
	conn, exists := httpsConnections[req.URL]
	if !exists || conn == nil {
		var err error
		conn, err = net.DialTimeout("tcp", req.URL, 30*time.Second)
		if err != nil {
			httpsConnMu.Unlock()
			return createErrorResponse(req.ID, fmt.Sprintf("Failed to connect: %v", err))
		}
		httpsConnections[req.URL] = conn
	}
	httpsConnMu.Unlock()

	// Decode and send data
	data, err := base64.StdEncoding.DecodeString(req.Body)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Failed to decode data: %v", err))
	}

	_, err = conn.Write(data)
	if err != nil {
		// Connection broken, remove it
		httpsConnMu.Lock()
		delete(httpsConnections, req.URL)
		httpsConnMu.Unlock()
		return createErrorResponse(req.ID, fmt.Sprintf("Failed to send data: %v", err))
	}

	// Read response with timeout
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	respBuffer := make([]byte, 64*1024)
	n, err := conn.Read(respBuffer)
	if err != nil && err != io.EOF {
		// May be timeout, which is ok for keep-alive connections
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// No data available yet, return empty
			resp := &TunnelResponse{
				ID:         req.ID,
				StatusCode: 200,
				Body:       "",
			}
			respJSON, _ := json.Marshal(resp)
			return string(respJSON), nil
		}
		return createErrorResponse(req.ID, fmt.Sprintf("Failed to read response: %v", err))
	}

	// Create response
	resp := &TunnelResponse{
		ID:         req.ID,
		StatusCode: 200,
		Body:       base64.StdEncoding.EncodeToString(respBuffer[:n]),
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		return createErrorResponse(req.ID, fmt.Sprintf("Failed to marshal response: %v", err))
	}

	return string(respJSON), nil
}

func createErrorResponse(id, errorMsg string) (string, error) {
	resp := &TunnelResponse{
		ID:         id,
		StatusCode: 502,
		Status:     "502 Bad Gateway",
		Error:      errorMsg,
	}
	respJSON, _ := json.Marshal(resp)
	return string(respJSON), nil
}
