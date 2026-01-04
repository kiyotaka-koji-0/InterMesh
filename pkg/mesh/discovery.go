package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Discovery handles peer discovery using UDP multicast
type Discovery struct {
	nodeID         string
	nodeName       string
	hasInternet    bool
	port           int
	multicastAddr  string
	conn           *net.UDPConn
	peerDiscovered func(peer *DiscoveredPeer)
	peerLost       func(peerID string)
	peers          map[string]*DiscoveredPeer
	peersMu        sync.RWMutex
	running        bool
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.Mutex
}

// DiscoveredPeer represents a discovered peer on the network
type DiscoveredPeer struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IP          string    `json:"ip"`
	Port        int       `json:"port"`
	HasInternet bool      `json:"has_internet"`
	LastSeen    time.Time `json:"last_seen"`
	MAC         string    `json:"mac"`
}

// AnnounceMessage is broadcast to discover peers
type AnnounceMessage struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Port        int    `json:"port"`
	HasInternet bool   `json:"has_internet"`
	MessageType string `json:"type"` // "announce" or "goodbye"
}

const (
	MulticastGroup   = "224.0.0.250:9999"
	AnnounceInterval = 5 * time.Second
	PeerTimeout      = 15 * time.Second
)

// NewDiscovery creates a new discovery service
func NewDiscovery(nodeID, nodeName string, port int, hasInternet bool) *Discovery {
	ctx, cancel := context.WithCancel(context.Background())
	return &Discovery{
		nodeID:        nodeID,
		nodeName:      nodeName,
		port:          port,
		hasInternet:   hasInternet,
		multicastAddr: MulticastGroup,
		peers:         make(map[string]*DiscoveredPeer),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// SetCallbacks sets the discovery callbacks
func (d *Discovery) SetCallbacks(discovered func(*DiscoveredPeer), lost func(string)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.peerDiscovered = discovered
	d.peerLost = lost
}

// UpdateInternetStatus updates the internet availability status
func (d *Discovery) UpdateInternetStatus(hasInternet bool) {
	d.mu.Lock()
	d.hasInternet = hasInternet
	d.mu.Unlock()
}

// Start begins the discovery process
func (d *Discovery) Start() error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return nil // Already running, not an error
	}
	// Reset context for restart capability
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.running = true
	d.mu.Unlock()

	// Setup multicast listener
	addr, err := net.ResolveUDPAddr("udp", d.multicastAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to listen on multicast: %w", err)
	}
	conn.SetReadBuffer(4096)
	d.conn = conn

	// Start announcement broadcast
	go d.announceLoop()

	// Start listening for announcements
	go d.listenLoop()

	// Start peer timeout checker
	go d.timeoutLoop()

	return nil
}

// Stop stops the discovery process
func (d *Discovery) Stop() {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return
	}
	d.running = false
	d.mu.Unlock()

	// Send goodbye message
	d.sendGoodbye()

	// Cancel context and close connection
	d.cancel()
	if d.conn != nil {
		d.conn.Close()
	}

	// Reset context for restart capability
	d.ctx, d.cancel = context.WithCancel(context.Background())
}

// GetPeers returns all discovered peers
func (d *Discovery) GetPeers() []*DiscoveredPeer {
	d.peersMu.RLock()
	defer d.peersMu.RUnlock()

	peers := make([]*DiscoveredPeer, 0, len(d.peers))
	for _, peer := range d.peers {
		peers = append(peers, peer)
	}
	return peers
}

// announceLoop periodically broadcasts presence
func (d *Discovery) announceLoop() {
	ticker := time.NewTicker(AnnounceInterval)
	defer ticker.Stop()

	// Send initial announcement immediately
	d.sendAnnounce()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.sendAnnounce()
		}
	}
}

// sendAnnounce sends an announcement message
func (d *Discovery) sendAnnounce() {
	d.mu.Lock()
	msg := AnnounceMessage{
		ID:          d.nodeID,
		Name:        d.nodeName,
		Port:        d.port,
		HasInternet: d.hasInternet,
		MessageType: "announce",
	}
	d.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	addr, _ := net.ResolveUDPAddr("udp", d.multicastAddr)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.Write(data)
}

// sendGoodbye sends a goodbye message before stopping
func (d *Discovery) sendGoodbye() {
	d.mu.Lock()
	msg := AnnounceMessage{
		ID:          d.nodeID,
		Name:        d.nodeName,
		Port:        d.port,
		MessageType: "goodbye",
	}
	d.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	addr, _ := net.ResolveUDPAddr("udp", d.multicastAddr)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.Write(data)
}

// listenLoop listens for announcements from other peers
func (d *Discovery) listenLoop() {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-d.ctx.Done():
			return
		default:
		}

		d.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, remoteAddr, err := d.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		var msg AnnounceMessage
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			continue
		}

		// Ignore our own announcements
		if msg.ID == d.nodeID {
			continue
		}

		if msg.MessageType == "goodbye" {
			d.handlePeerGoodbye(msg.ID)
		} else {
			d.handlePeerAnnounce(&msg, remoteAddr.IP.String())
		}
	}
}

// handlePeerAnnounce processes a peer announcement
func (d *Discovery) handlePeerAnnounce(msg *AnnounceMessage, ip string) {
	d.peersMu.Lock()
	existing, found := d.peers[msg.ID]

	peer := &DiscoveredPeer{
		ID:          msg.ID,
		Name:        msg.Name,
		IP:          ip,
		Port:        msg.Port,
		HasInternet: msg.HasInternet,
		LastSeen:    time.Now(),
		MAC:         "", // MAC address would need ARP lookup
	}

	d.peers[msg.ID] = peer
	d.peersMu.Unlock()

	// Notify if this is a new peer
	if !found && d.peerDiscovered != nil {
		d.peerDiscovered(peer)
	} else if found && existing.HasInternet != peer.HasInternet {
		// Internet status changed
		if d.peerDiscovered != nil {
			d.peerDiscovered(peer)
		}
	}
}

// handlePeerGoodbye processes a peer goodbye message
func (d *Discovery) handlePeerGoodbye(peerID string) {
	d.peersMu.Lock()
	delete(d.peers, peerID)
	d.peersMu.Unlock()

	if d.peerLost != nil {
		d.peerLost(peerID)
	}
}

// timeoutLoop checks for peers that haven't been seen recently
func (d *Discovery) timeoutLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.checkPeerTimeouts()
		}
	}
}

// checkPeerTimeouts removes peers that haven't announced recently
func (d *Discovery) checkPeerTimeouts() {
	now := time.Now()

	d.peersMu.Lock()
	for id, peer := range d.peers {
		if now.Sub(peer.LastSeen) > PeerTimeout {
			delete(d.peers, id)
			if d.peerLost != nil {
				go d.peerLost(id)
			}
		}
	}
	d.peersMu.Unlock()
}
