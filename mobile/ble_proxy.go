package intermesh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// BLEProxyHandler handles internet proxy requests through BLE connections
type BLEProxyHandler struct {
	nodeID         string
	activeRequests map[string]*ProxyRequest
	requestsMu     sync.RWMutex
	onBLEMessage   func(peerID string, messageType string, data []byte) error
	mobileApp      *MobileApp
}

// ProxyRequest represents an ongoing internet request
type ProxyRequest struct {
	RequestID string
	ClientID  string
	URL       string
	Method    string
	Headers   map[string]string
	Body      []byte
	CreatedAt time.Time
}

// ProxyResponse represents a response to a proxy request
type ProxyResponse struct {
	RequestID  string
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// BLEProxyMessage represents messages sent over BLE for proxy functionality
type BLEProxyMessage struct {
	Type      string      `json:"type"` // "request", "response"
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// NewBLEProxyHandler creates a new BLE proxy handler
func NewBLEProxyHandler(nodeID string, mobileApp *MobileApp) *BLEProxyHandler {
	return &BLEProxyHandler{
		nodeID:         nodeID,
		activeRequests: make(map[string]*ProxyRequest),
		mobileApp:      mobileApp,
	}
}

// SetBLEMessageSender sets the callback for sending BLE messages
func (h *BLEProxyHandler) SetBLEMessageSender(sender func(peerID string, messageType string, data []byte) error) {
	h.onBLEMessage = sender
}

// HandleProxyRequest handles an incoming proxy request from a BLE peer
func (h *BLEProxyHandler) HandleProxyRequest(clientID string, request *ProxyRequest) error {
	// Only handle if we have internet
	if !h.mobileApp.HasInternet() {
		return fmt.Errorf("no internet access available")
	}

	// Store the request
	h.requestsMu.Lock()
	h.activeRequests[request.RequestID] = request
	h.requestsMu.Unlock()

	// Make the HTTP request
	go h.executeProxyRequest(clientID, request)

	return nil
}

// executeProxyRequest executes an HTTP request and sends response back through BLE
func (h *BLEProxyHandler) executeProxyRequest(clientID string, request *ProxyRequest) {
	defer func() {
		h.requestsMu.Lock()
		delete(h.activeRequests, request.RequestID)
		h.requestsMu.Unlock()
	}()

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create HTTP request
	var bodyReader io.Reader
	if len(request.Body) > 0 {
		bodyReader = bytes.NewReader(request.Body)
	}

	httpReq, err := http.NewRequest(request.Method, request.URL, bodyReader)
	if err != nil {
		h.sendErrorResponse(clientID, request.RequestID, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Add headers
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(httpReq)
	if err != nil {
		h.sendErrorResponse(clientID, request.RequestID, fmt.Sprintf("Request failed: %v", err))
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.sendErrorResponse(clientID, request.RequestID, fmt.Sprintf("Failed to read response: %v", err))
		return
	}

	// Create response
	response := &ProxyResponse{
		RequestID:  request.RequestID,
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       body,
	}

	// Copy headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	// Send response back through BLE
	h.sendProxyResponse(clientID, response)
}

// sendProxyResponse sends a proxy response through BLE
func (h *BLEProxyHandler) sendProxyResponse(clientID string, response *ProxyResponse) {
	if h.onBLEMessage == nil {
		return
	}

	message := &BLEProxyMessage{
		Type:      "response",
		RequestID: response.RequestID,
		Data:      response,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	h.onBLEMessage(clientID, "internet_proxy", data)
}

// sendErrorResponse sends an error response through BLE
func (h *BLEProxyHandler) sendErrorResponse(clientID, requestID, errorMsg string) {
	response := &ProxyResponse{
		RequestID:  requestID,
		StatusCode: 500,
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       []byte(errorMsg),
	}

	h.sendProxyResponse(clientID, response)
}

// SendProxyRequest sends a proxy request to a BLE peer with internet
func (h *BLEProxyHandler) SendProxyRequest(proxyPeerID, url, method string, headers map[string]string, body []byte) (string, error) {
	if h.onBLEMessage == nil {
		return "", fmt.Errorf("BLE message sender not configured")
	}

	// Generate request ID
	requestID := fmt.Sprintf("%s-%d", h.nodeID, time.Now().UnixNano())

	request := &ProxyRequest{
		RequestID: requestID,
		ClientID:  h.nodeID,
		URL:       url,
		Method:    method,
		Headers:   headers,
		Body:      body,
		CreatedAt: time.Now(),
	}

	message := &BLEProxyMessage{
		Type:      "request",
		RequestID: requestID,
		Data:      request,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request through BLE
	err = h.onBLEMessage(proxyPeerID, "internet_proxy", data)
	if err != nil {
		return "", fmt.Errorf("failed to send BLE message: %w", err)
	}

	return requestID, nil
}

// HandleBLEProxyMessage handles incoming BLE proxy messages
func (h *BLEProxyHandler) HandleBLEProxyMessage(senderID string, data []byte) error {
	var message BLEProxyMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return fmt.Errorf("failed to unmarshal proxy message: %w", err)
	}

	switch message.Type {
	case "request":
		// Handle proxy request
		requestData, err := json.Marshal(message.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal request data: %w", err)
		}

		var request ProxyRequest
		err = json.Unmarshal(requestData, &request)
		if err != nil {
			return fmt.Errorf("failed to unmarshal proxy request: %w", err)
		}

		return h.HandleProxyRequest(senderID, &request)

	case "response":
		// Handle proxy response (for clients waiting for responses)
		// This would be handled by the client-side logic
		// For now, just log it
		fmt.Printf("Received proxy response for request %s\n", message.RequestID)

	default:
		return fmt.Errorf("unknown proxy message type: %s", message.Type)
	}

	return nil
}
