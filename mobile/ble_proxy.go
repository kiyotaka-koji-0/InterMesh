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
	nodeID           string
	activeRequests   map[string]*ProxyRequest
	requestsMu       sync.RWMutex
	onBLEMessage     func(peerID string, messageType string, data []byte) error
	mobileApp        *MobileApp
	pendingResponses map[string]chan *ProxyResponse
	responsesMu      sync.RWMutex
}

// ProxyRequest represents an ongoing internet request
type ProxyRequest struct {
	RequestID string            `json:"request_id"`
	ClientID  string            `json:"client_id"`
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body"`
	CreatedAt time.Time         `json:"created_at"`
}

// ProxyResponse represents a response to a proxy request
type ProxyResponse struct {
	RequestID  string            `json:"request_id"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Error      string            `json:"error,omitempty"`
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
		nodeID:           nodeID,
		activeRequests:   make(map[string]*ProxyRequest),
		mobileApp:        mobileApp,
		pendingResponses: make(map[string]chan *ProxyResponse),
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
		responseData, err := json.Marshal(message.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal response data: %w", err)
		}

		var response ProxyResponse
		err = json.Unmarshal(responseData, &response)
		if err != nil {
			return fmt.Errorf("failed to unmarshal proxy response: %w", err)
		}

		// Send to waiting goroutine
		h.responsesMu.RLock()
		if ch, ok := h.pendingResponses[message.RequestID]; ok {
			select {
			case ch <- &response:
			default:
			}
		}
		h.responsesMu.RUnlock()

	default:
		return fmt.Errorf("unknown proxy message type: %s", message.Type)
	}

	return nil
}

// ExecuteProxyRequestSync executes an HTTP request synchronously (for devices with internet)
// This is called by devices that receive a proxy request and have internet access
func (h *BLEProxyHandler) ExecuteProxyRequestSync(requestJSON string) (string, error) {
	var request ProxyRequest
	err := json.Unmarshal([]byte(requestJSON), &request)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Execute the HTTP request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var bodyReader io.Reader
	if len(request.Body) > 0 {
		bodyReader = bytes.NewReader(request.Body)
	}

	httpReq, err := http.NewRequest(request.Method, request.URL, bodyReader)
	if err != nil {
		response := &ProxyResponse{
			RequestID:  request.RequestID,
			StatusCode: 500,
			Error:      fmt.Sprintf("Invalid request: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	// Add headers
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(httpReq)
	if err != nil {
		response := &ProxyResponse{
			RequestID:  request.RequestID,
			StatusCode: 500,
			Error:      fmt.Sprintf("Request failed: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response := &ProxyResponse{
			RequestID:  request.RequestID,
			StatusCode: 500,
			Error:      fmt.Sprintf("Failed to read response: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
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

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(responseJSON), nil
}

// CreateProxyRequest creates a JSON proxy request for sending via BLE
func (h *BLEProxyHandler) CreateProxyRequest(url, method string) (string, error) {
	requestID := fmt.Sprintf("%s-%d", h.nodeID, time.Now().UnixNano())

	request := &ProxyRequest{
		RequestID: requestID,
		ClientID:  h.nodeID,
		URL:       url,
		Method:    method,
		Headers:   make(map[string]string),
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

	return string(data), nil
}

// parseProxyResponse parses a JSON proxy response (internal use)
func (h *BLEProxyHandler) parseProxyResponse(responseJSON string) (int, string, error) {
	var response ProxyResponse
	err := json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		return 0, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != "" {
		return response.StatusCode, response.Error, nil
	}

	return response.StatusCode, string(response.Body), nil
}

// ParseProxyResponseBody parses a JSON proxy response and returns just the body
func (h *BLEProxyHandler) ParseProxyResponseBody(responseJSON string) string {
	_, body, err := h.parseProxyResponse(responseJSON)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return body
}
