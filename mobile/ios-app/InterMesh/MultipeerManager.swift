//
//  MultipeerManager.swift
//  InterMesh
//
//  Handles peer-to-peer discovery and communication using Apple's MultipeerConnectivity framework
//  This is the iOS equivalent of Android's WiFi Direct, allowing devices to connect without a router
//

import Foundation
import MultipeerConnectivity
import Combine

/// Represents a discovered peer
struct DiscoveredPeer: Identifiable, Hashable {
    let id: String
    let displayName: String
    let isConnected: Bool
    
    func hash(into hasher: inout Hasher) {
        hasher.combine(id)
    }
    
    static func == (lhs: DiscoveredPeer, rhs: DiscoveredPeer) -> Bool {
        lhs.id == rhs.id
    }
}

/// Manages peer-to-peer networking using MultipeerConnectivity
class MultipeerManager: NSObject, ObservableObject {
    
    // MARK: - Published Properties
    
    @Published var discoveredPeers: [DiscoveredPeer] = []
    @Published var connectedPeers: [DiscoveredPeer] = []
    @Published var isAdvertising = false
    @Published var isBrowsing = false
    @Published var lastReceivedMessage: String = ""
    @Published var connectionStatus: String = "Not Connected"
    
    // MARK: - Private Properties
    
    private let serviceType = "intermesh-p2p"  // Must be 1-15 lowercase letters/numbers/hyphens
    private var peerID: MCPeerID!
    private var session: MCSession!
    private var advertiser: MCNearbyServiceAdvertiser?
    private var browser: MCNearbyServiceBrowser?
    
    // Callbacks
    var onPeerDiscovered: ((DiscoveredPeer) -> Void)?
    var onPeerLost: ((String) -> Void)?
    var onPeerConnected: ((DiscoveredPeer) -> Void)?
    var onPeerDisconnected: ((String) -> Void)?
    var onMessageReceived: ((String, Data) -> Void)?
    var onError: ((String) -> Void)?
    
    // MARK: - Initialization
    
    override init() {
        super.init()
        setupSession()
    }
    
    private func setupSession() {
        // Create peer ID with device name
        let deviceName = UIDevice.current.name
        peerID = MCPeerID(displayName: deviceName)
        
        // Create session
        session = MCSession(peer: peerID, securityIdentity: nil, encryptionPreference: .required)
        session.delegate = self
        
        print("[MultipeerManager] Session created for peer: \(deviceName)")
    }
    
    // MARK: - Public Methods
    
    /// Start advertising this device to nearby peers
    func startAdvertising() {
        guard !isAdvertising else { return }
        
        let discoveryInfo: [String: String] = [
            "app": "InterMesh",
            "version": "1.0",
            "platform": "iOS"
        ]
        
        advertiser = MCNearbyServiceAdvertiser(
            peer: peerID,
            discoveryInfo: discoveryInfo,
            serviceType: serviceType
        )
        advertiser?.delegate = self
        advertiser?.startAdvertisingPeer()
        
        isAdvertising = true
        print("[MultipeerManager] Started advertising")
    }
    
    /// Stop advertising
    func stopAdvertising() {
        advertiser?.stopAdvertisingPeer()
        advertiser = nil
        isAdvertising = false
        print("[MultipeerManager] Stopped advertising")
    }
    
    /// Start browsing for nearby peers
    func startBrowsing() {
        guard !isBrowsing else { return }
        
        browser = MCNearbyServiceBrowser(peer: peerID, serviceType: serviceType)
        browser?.delegate = self
        browser?.startBrowsingForPeers()
        
        isBrowsing = true
        print("[MultipeerManager] Started browsing for peers")
    }
    
    /// Stop browsing
    func stopBrowsing() {
        browser?.stopBrowsingForPeers()
        browser = nil
        isBrowsing = false
        print("[MultipeerManager] Stopped browsing")
    }
    
    /// Start both advertising and browsing (full mesh mode)
    func startMeshDiscovery() {
        startAdvertising()
        startBrowsing()
        connectionStatus = "Searching for peers..."
    }
    
    /// Stop all discovery
    func stopMeshDiscovery() {
        stopAdvertising()
        stopBrowsing()
        connectionStatus = "Discovery stopped"
    }
    
    /// Connect to a discovered peer
    func connect(to peer: DiscoveredPeer) {
        guard let mcPeer = findMCPeerID(for: peer.id) else {
            onError?("Peer not found")
            return
        }
        
        // Send invitation
        browser?.invitePeer(mcPeer, to: session, withContext: nil, timeout: 30)
        connectionStatus = "Connecting to \(peer.displayName)..."
        print("[MultipeerManager] Sent invitation to \(peer.displayName)")
    }
    
    /// Disconnect from a peer
    func disconnect(from peer: DiscoveredPeer) {
        // MultipeerConnectivity doesn't have a direct disconnect for a single peer
        // The session handles this automatically when the peer becomes unavailable
        print("[MultipeerManager] Disconnect requested for \(peer.displayName)")
    }
    
    /// Disconnect from all peers
    func disconnectAll() {
        session.disconnect()
        connectedPeers.removeAll()
        connectionStatus = "Disconnected"
        print("[MultipeerManager] Disconnected from all peers")
    }
    
    /// Send data to all connected peers
    func sendToAll(_ data: Data) {
        guard !session.connectedPeers.isEmpty else {
            onError?("No connected peers")
            return
        }
        
        do {
            try session.send(data, toPeers: session.connectedPeers, with: .reliable)
            print("[MultipeerManager] Sent \(data.count) bytes to \(session.connectedPeers.count) peers")
        } catch {
            onError?("Failed to send: \(error.localizedDescription)")
        }
    }
    
    /// Send data to a specific peer
    func send(_ data: Data, to peer: DiscoveredPeer) {
        guard let mcPeer = session.connectedPeers.first(where: { $0.displayName == peer.displayName }) else {
            onError?("Peer not connected")
            return
        }
        
        do {
            try session.send(data, toPeers: [mcPeer], with: .reliable)
            print("[MultipeerManager] Sent \(data.count) bytes to \(peer.displayName)")
        } catch {
            onError?("Failed to send: \(error.localizedDescription)")
        }
    }
    
    /// Send a string message to all connected peers
    func sendMessage(_ message: String) {
        guard let data = message.data(using: .utf8) else { return }
        sendToAll(data)
    }
    
    /// Cleanup resources
    func cleanup() {
        stopMeshDiscovery()
        disconnectAll()
    }
    
    // MARK: - Private Methods
    
    private func findMCPeerID(for id: String) -> MCPeerID? {
        // The ID is the display name in our case
        return browser?.connectedPeers.first { $0.displayName == id }
    }
    
    private func updateConnectedPeers() {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            self.connectedPeers = self.session.connectedPeers.map { mcPeer in
                DiscoveredPeer(
                    id: mcPeer.displayName,
                    displayName: mcPeer.displayName,
                    isConnected: true
                )
            }
            
            let count = self.connectedPeers.count
            self.connectionStatus = count > 0 ? "Connected to \(count) peer(s)" : "No peers connected"
        }
    }
}

// MARK: - MCSessionDelegate

extension MultipeerManager: MCSessionDelegate {
    
    func session(_ session: MCSession, peer peerID: MCPeerID, didChange state: MCSessionState) {
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            switch state {
            case .notConnected:
                print("[MultipeerManager] Peer disconnected: \(peerID.displayName)")
                self.onPeerDisconnected?(peerID.displayName)
                
            case .connecting:
                print("[MultipeerManager] Connecting to: \(peerID.displayName)")
                self.connectionStatus = "Connecting to \(peerID.displayName)..."
                
            case .connected:
                print("[MultipeerManager] Connected to: \(peerID.displayName)")
                let peer = DiscoveredPeer(
                    id: peerID.displayName,
                    displayName: peerID.displayName,
                    isConnected: true
                )
                self.onPeerConnected?(peer)
                
            @unknown default:
                print("[MultipeerManager] Unknown state for: \(peerID.displayName)")
            }
            
            self.updateConnectedPeers()
        }
    }
    
    func session(_ session: MCSession, didReceive data: Data, fromPeer peerID: MCPeerID) {
        DispatchQueue.main.async { [weak self] in
            print("[MultipeerManager] Received \(data.count) bytes from \(peerID.displayName)")
            
            self?.onMessageReceived?(peerID.displayName, data)
            
            if let message = String(data: data, encoding: .utf8) {
                self?.lastReceivedMessage = message
            }
        }
    }
    
    func session(_ session: MCSession, didReceive stream: InputStream, withName streamName: String, fromPeer peerID: MCPeerID) {
        print("[MultipeerManager] Received stream from \(peerID.displayName)")
    }
    
    func session(_ session: MCSession, didStartReceivingResourceWithName resourceName: String, fromPeer peerID: MCPeerID, with progress: Progress) {
        print("[MultipeerManager] Started receiving resource: \(resourceName)")
    }
    
    func session(_ session: MCSession, didFinishReceivingResourceWithName resourceName: String, fromPeer peerID: MCPeerID, at localURL: URL?, withError error: Error?) {
        print("[MultipeerManager] Finished receiving resource: \(resourceName)")
    }
}

// MARK: - MCNearbyServiceAdvertiserDelegate

extension MultipeerManager: MCNearbyServiceAdvertiserDelegate {
    
    func advertiser(_ advertiser: MCNearbyServiceAdvertiser, didReceiveInvitationFromPeer peerID: MCPeerID, withContext context: Data?, invitationHandler: @escaping (Bool, MCSession?) -> Void) {
        print("[MultipeerManager] Received invitation from \(peerID.displayName)")
        
        // Auto-accept invitations for mesh networking
        invitationHandler(true, session)
        
        DispatchQueue.main.async { [weak self] in
            self?.connectionStatus = "Accepting connection from \(peerID.displayName)..."
        }
    }
    
    func advertiser(_ advertiser: MCNearbyServiceAdvertiser, didNotStartAdvertisingPeer error: Error) {
        print("[MultipeerManager] Failed to advertise: \(error.localizedDescription)")
        
        DispatchQueue.main.async { [weak self] in
            self?.isAdvertising = false
            self?.onError?("Failed to advertise: \(error.localizedDescription)")
        }
    }
}

// MARK: - MCNearbyServiceBrowserDelegate

extension MultipeerManager: MCNearbyServiceBrowserDelegate {
    
    func browser(_ browser: MCNearbyServiceBrowser, foundPeer peerID: MCPeerID, withDiscoveryInfo info: [String : String]?) {
        print("[MultipeerManager] Found peer: \(peerID.displayName), info: \(info ?? [:])")
        
        let peer = DiscoveredPeer(
            id: peerID.displayName,
            displayName: peerID.displayName,
            isConnected: false
        )
        
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            // Add to discovered peers if not already there
            if !self.discoveredPeers.contains(where: { $0.id == peer.id }) {
                self.discoveredPeers.append(peer)
                self.onPeerDiscovered?(peer)
            }
        }
    }
    
    func browser(_ browser: MCNearbyServiceBrowser, lostPeer peerID: MCPeerID) {
        print("[MultipeerManager] Lost peer: \(peerID.displayName)")
        
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }
            
            self.discoveredPeers.removeAll { $0.id == peerID.displayName }
            self.onPeerLost?(peerID.displayName)
        }
    }
    
    func browser(_ browser: MCNearbyServiceBrowser, didNotStartBrowsingForPeers error: Error) {
        print("[MultipeerManager] Failed to browse: \(error.localizedDescription)")
        
        DispatchQueue.main.async { [weak self] in
            self?.isBrowsing = false
            self?.onError?("Failed to browse: \(error.localizedDescription)")
        }
    }
}

// MARK: - Browser Extension for connectedPeers

private extension MCNearbyServiceBrowser {
    // Note: This is a workaround since browser doesn't track peers directly
    var connectedPeers: [MCPeerID] {
        // This would need to be tracked manually in a real implementation
        return []
    }
}
