package intermesh

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kiyotaka-koji-0/intermesh/pkg/mesh"
)

func TestBLERelayIntegration(t *testing.T) {
	// 1. Setup "Internet" server (A's destination)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello from Internet")
	}))
	defer ts.Close()

	// 2. Setup Node A (The Internet Exit)
	appA := NewMobileApp("node-A", "Device A", "127.0.0.1", "00:00:00:00:00:01")
	appA.app.Node.SetInternetStatus(true)
	appA.app.EnableInternetSharing()
	appA.app.Start()
	defer appA.Stop()

	// 3. Setup Node B (The Relay)
	appB := NewMobileApp("node-B", "Device B", "127.0.0.1", "00:00:00:00:00:02")
	appB.app.Start()
	defer appB.Stop()

	// Simulating Discovery: Manually add A to B's discovery list
	// Since we are in separate packages, we'll use a UDP packet to trigger B's discovery
	announceMsg := mesh.AnnounceMessage{
		ID:          "node-A",
		Name:        "Device A",
		Port:        9998,
		HasInternet: true,
		MessageType: "announce",
	}
	data, _ := json.Marshal(announceMsg)

	// Send to multicast port
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9999")
	conn, _ := net.DialUDP("udp", nil, addr)
	conn.Write(data)
	conn.Close()

	// Give discovery a moment to process
	time.Sleep(200 * time.Millisecond)

	// Bridging B's InternetClient to A's InternetProxy
	success := appB.app.RequestInternetAccess()
	if !success {
		// Fallback: Manually connect if discovery/transport handshake is too slow for the test
		appB.app.InternetClient.ConnectToProxy("node-A", "127.0.0.1", 9997)
	}

	// 4. Setup simulated BLE "Client C" and its interaction with B
	// Mock C's request
	request := &ProxyRequest{
		RequestID: "req-1",
		ClientID:  "node-C",
		URL:       ts.URL,
		Method:    "GET",
		Headers:   make(map[string]string),
	}

	bleMsg := &BLEProxyMessage{
		Type:      "request",
		RequestID: "req-1",
		Data:      request,
	}

	bleData, _ := json.Marshal(bleMsg)

	// Setup a way to capture the response from B to C
	var capturedResponse []byte
	appB.SetBLEMessageSender(func(peerID string, messageType string, data []byte) error {
		if peerID == "node-C" {
			capturedResponse = data
		}
		return nil
	})

	// 5. Execute: Simulate B receiving the BLE request from C
	err := appB.HandleBLEProxyMessage("node-C", bleData)
	if err != nil {
		t.Fatalf("HandleBLEProxyMessage failed: %v", err)
	}

	// Wait for async relay to complete
	time.Sleep(1 * time.Second)

	// 6. Verify: Did B relay to A and send response back to C?
	if capturedResponse == nil {
		t.Fatal("No response captured from Relay B to Client C")
	}

	var responseMsg BLEProxyMessage
	err = json.Unmarshal(capturedResponse, &responseMsg)
	if err != nil {
		t.Fatalf("Failed to unmarshal captured response: %v", err)
	}

	if responseMsg.Type != "response" {
		t.Errorf("Expected response type, got %s", responseMsg.Type)
	}

	// Verify the body of the response
	// Note: responseMsg.Data is interface{}, need to handle it
	respBytes, _ := json.Marshal(responseMsg.Data)
	var proxyResp ProxyResponse
	json.Unmarshal(respBytes, &proxyResp)

	if string(proxyResp.Body) != "Hello from Internet" {
		t.Errorf("Expected 'Hello from Internet', got '%s'", string(proxyResp.Body))
	}

	if proxyResp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", proxyResp.StatusCode)
	}
}
