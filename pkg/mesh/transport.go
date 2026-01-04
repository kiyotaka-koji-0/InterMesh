package mesh

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Transport handles TCP connections between peers
type Transport struct {
	nodeID      string
	port        int
	listener    net.Listener
	connections map[string]*Connection
	connMu      sync.RWMutex
	onMessage   func(peerID string, msg *Message)
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool
	mu          sync.Mutex
}

// Connection represents a connection to a peer
type Connection struct {
	PeerID    string
	Conn      net.Conn
	Connected bool
	mu        sync.Mutex
}

// Message represents a message sent between peers
type Message struct {
	Type      string            `json:"type"`    // "data", "route", "proxy_request", etc.
	Source    string            `json:"source"`  // Source node ID
	Dest      string            `json:"dest"`    // Destination node ID
	Payload   []byte            `json:"payload"` // Message payload
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

const (
	DefaultPort    = 9998
	MaxMessageSize = 65536 // 64KB max message size
)

// NewTransport creates a new transport layer
func NewTransport(nodeID string, port int) *Transport {
	ctx, cancel := context.WithCancel(context.Background())
	return &Transport{
		nodeID:      nodeID,
		port:        port,
		connections: make(map[string]*Connection),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// SetMessageHandler sets the callback for received messages
func (t *Transport) SetMessageHandler(handler func(string, *Message)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onMessage = handler
}

// Start starts the transport layer
func (t *Transport) Start() error {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return nil // Already running, not an error
	}
	t.running = true
	// Reset context for restart
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.mu.Unlock()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", t.port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	t.listener = listener

	go t.acceptLoop()

	return nil
}

// Stop stops the transport layer
func (t *Transport) Stop() {
	t.mu.Lock()
	if !t.running {
		t.mu.Unlock()
		return
	}
	t.running = false
	t.mu.Unlock()

	t.cancel()

	if t.listener != nil {
		t.listener.Close()
	}

	// Close all connections
	t.connMu.Lock()
	for _, conn := range t.connections {
		conn.Close()
	}
	t.connections = make(map[string]*Connection)
	t.connMu.Unlock()
}

// ConnectToPeer establishes a connection to a peer
func (t *Transport) ConnectToPeer(peerID, ip string, port int) error {
	// Check if already connected
	t.connMu.RLock()
	if _, exists := t.connections[peerID]; exists {
		t.connMu.RUnlock()
		return nil // Already connected
	}
	t.connMu.RUnlock()

	// Establish TCP connection
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Send handshake
	handshake := Message{
		Type:      "handshake",
		Source:    t.nodeID,
		Dest:      peerID,
		Timestamp: time.Now(),
	}
	if err := t.sendMessage(conn, &handshake); err != nil {
		conn.Close()
		return fmt.Errorf("handshake failed: %w", err)
	}

	// Create connection object
	connection := &Connection{
		PeerID:    peerID,
		Conn:      conn,
		Connected: true,
	}

	t.connMu.Lock()
	t.connections[peerID] = connection
	t.connMu.Unlock()

	// Start reading from this connection
	go t.handleConnection(connection)

	return nil
}

// DisconnectPeer closes the connection to a peer
func (t *Transport) DisconnectPeer(peerID string) {
	t.connMu.Lock()
	conn, exists := t.connections[peerID]
	if exists {
		delete(t.connections, peerID)
	}
	t.connMu.Unlock()

	if exists {
		conn.Close()
	}
}

// SendMessage sends a message to a peer
func (t *Transport) SendMessage(peerID string, msg *Message) error {
	t.connMu.RLock()
	conn, exists := t.connections[peerID]
	t.connMu.RUnlock()

	if !exists {
		return fmt.Errorf("not connected to peer %s", peerID)
	}

	return t.sendMessage(conn.Conn, msg)
}

// BroadcastMessage sends a message to all connected peers
func (t *Transport) BroadcastMessage(msg *Message) {
	t.connMu.RLock()
	connections := make([]*Connection, 0, len(t.connections))
	for _, conn := range t.connections {
		connections = append(connections, conn)
	}
	t.connMu.RUnlock()

	for _, conn := range connections {
		t.sendMessage(conn.Conn, msg)
	}
}

// GetConnectedPeers returns list of connected peer IDs
func (t *Transport) GetConnectedPeers() []string {
	t.connMu.RLock()
	defer t.connMu.RUnlock()

	peers := make([]string, 0, len(t.connections))
	for peerID := range t.connections {
		peers = append(peers, peerID)
	}
	return peers
}

// acceptLoop accepts incoming connections
func (t *Transport) acceptLoop() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		t.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
		conn, err := t.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		go t.handleIncomingConnection(conn)
	}
}

// handleIncomingConnection handles a new incoming connection
func (t *Transport) handleIncomingConnection(conn net.Conn) {
	// Read handshake
	msg, err := t.readMessage(conn)
	if err != nil || msg.Type != "handshake" {
		conn.Close()
		return
	}

	peerID := msg.Source

	// Create connection object
	connection := &Connection{
		PeerID:    peerID,
		Conn:      conn,
		Connected: true,
	}

	t.connMu.Lock()
	t.connections[peerID] = connection
	t.connMu.Unlock()

	// Handle messages from this connection
	t.handleConnection(connection)
}

// handleConnection reads messages from a connection
func (t *Transport) handleConnection(conn *Connection) {
	defer func() {
		conn.Close()
		t.connMu.Lock()
		delete(t.connections, conn.PeerID)
		t.connMu.Unlock()
	}()

	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		conn.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		msg, err := t.readMessage(conn.Conn)
		if err != nil {
			return
		}

		// Handle message
		if t.onMessage != nil {
			t.onMessage(conn.PeerID, msg)
		}
	}
}

// sendMessage sends a message over a connection
func (t *Transport) sendMessage(conn net.Conn, msg *Message) error {
	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Write length prefix (4 bytes)
	length := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	// Write message data
	_, err = conn.Write(data)
	return err
}

// readMessage reads a message from a connection
func (t *Transport) readMessage(conn net.Conn) (*Message, error) {
	// Read length prefix
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if length > MaxMessageSize {
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	// Deserialize message
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// Close closes a connection
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Connected {
		c.Connected = false
		if c.Conn != nil {
			c.Conn.Close()
		}
	}
}
