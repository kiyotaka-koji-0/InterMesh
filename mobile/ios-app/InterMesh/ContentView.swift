//
//  ContentView.swift
//  InterMesh
//
//  Created by InterMesh Team
//

import SwiftUI
import Intermesh

struct ContentView: View {
    @StateObject private var meshManager = MeshManager()
    @StateObject private var multipeerManager = MultipeerManager()
    @StateObject private var bleManager = BLEManager()
    @State private var showPeerList = false
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(spacing: 20) {
                    // Header
                    HStack {
                        Text("InterMesh")
                            .font(.largeTitle)
                            .fontWeight(.bold)
                        
                        Spacer()
                        
                        // P2P Status indicators
                        HStack(spacing: 8) {
                            if multipeerManager.isBrowsing {
                                HStack(spacing: 4) {
                                    Circle()
                                        .fill(Color.green)
                                        .frame(width: 8, height: 8)
                                    Text("MC")
                                        .font(.caption)
                                        .foregroundColor(.green)
                                }
                            }
                            
                            if bleManager.isScanning || bleManager.isAdvertising {
                                HStack(spacing: 4) {
                                    Circle()
                                        .fill(Color.blue)
                                        .frame(width: 8, height: 8)
                                    Text("BLE")
                                        .font(.caption)
                                        .foregroundColor(.blue)
                                }
                            }
                        }
                    }
                    .padding(.top, 20)
                    .padding(.horizontal)
                    
                    // Device Info Card
                    VStack(alignment: .leading, spacing: 10) {
                        Text("Device Information")
                            .font(.headline)
                        
                        HStack {
                            Text("Device ID:")
                                .foregroundColor(.gray)
                            Spacer()
                            Text(String(meshManager.deviceID.suffix(12)))
                                .fontWeight(.medium)
                                .font(.system(.body, design: .monospaced))
                        }
                        
                        HStack {
                            Text("Status:")
                                .foregroundColor(.gray)
                            Spacer()
                            Text(meshManager.isConnected ? "Connected" : "Disconnected")
                                .foregroundColor(meshManager.isConnected ? .green : .red)
                                .fontWeight(.medium)
                        }
                        
                        HStack {
                            Text("P2P Status:")
                                .foregroundColor(.gray)
                            Spacer()
                            VStack(alignment: .trailing) {
                                Text("MC: \(multipeerManager.connectionStatus)")
                                    .font(.caption)
                                Text("BLE: \(bleManager.connectionStatus)")
                                    .font(.caption)
                            }
                            .foregroundColor(.orange)
                        }
                    }
                    .padding()
                    .background(Color(.systemGray6))
                    .cornerRadius(12)
                    .padding(.horizontal)
                    
                    // Statistics Card
                    VStack(alignment: .leading, spacing: 10) {
                        Text("Network Statistics")
                            .font(.headline)
                        
                        HStack {
                            // Mesh Peers
                            VStack(alignment: .leading) {
                                Text("\(meshManager.peerCount)")
                                    .font(.system(size: 32, weight: .bold))
                                    .foregroundColor(.blue)
                                Text("Mesh Peers")
                                    .font(.caption)
                                    .foregroundColor(.gray)
                            }
                            
                            Spacer()
                            
                            // P2P Peers (tappable) - Combined MC + BLE
                            Button(action: { showPeerList = true }) {
                                VStack(alignment: .leading) {
                                    let totalP2P = multipeerManager.discoveredPeers.count + bleManager.discoveredPeers.count
                                    Text("\(totalP2P)")
                                        .font(.system(size: 32, weight: .bold))
                                        .foregroundColor(.purple)
                                    Text("P2P Peers")
                                        .font(.caption)
                                        .foregroundColor(.gray)
                                }
                            }
                            .disabled(multipeerManager.discoveredPeers.isEmpty && bleManager.discoveredPeers.isEmpty)
                            
                            Spacer()
                            
                            // Proxies
                            VStack(alignment: .leading) {
                                Text("\(meshManager.proxyCount)")
                                    .font(.system(size: 32, weight: .bold))
                                    .foregroundColor(.green)
                                Text("Proxies")
                                    .font(.caption)
                                    .foregroundColor(.gray)
                            }
                        }
                        
                        // Connected P2P peers indicator
                        if !multipeerManager.connectedPeers.isEmpty || !bleManager.connectedPeers.isEmpty {
                            Divider()
                            VStack(alignment: .leading, spacing: 4) {
                                if !multipeerManager.connectedPeers.isEmpty {
                                    HStack {
                                        Image(systemName: "wifi")
                                            .foregroundColor(.green)
                                        Text("MC: \(multipeerManager.connectedPeers.map { $0.displayName }.joined(separator: ", "))")
                                            .font(.caption)
                                            .foregroundColor(.green)
                                    }
                                }
                                if !bleManager.connectedPeers.isEmpty {
                                    HStack {
                                        Image(systemName: "wave.3.right")
                                            .foregroundColor(.blue)
                                        Text("BLE: \(bleManager.connectedPeers.map { "\($0.name) (\($0.platform))" }.joined(separator: ", "))")
                                            .font(.caption)
                                            .foregroundColor(.blue)
                                    }
                                }
                            }
                        }
                    }
                    .padding()
                    .background(Color(.systemGray6))
                    .cornerRadius(12)
                    .padding(.horizontal)
                    
                    // Internet Sharing Toggle
                    VStack(spacing: 10) {
                        Toggle("Share Internet", isOn: $meshManager.isSharingInternet)
                            .padding()
                            .background(Color(.systemGray6))
                            .cornerRadius(12)
                            .onChange(of: meshManager.isSharingInternet) { newValue in
                                meshManager.toggleInternetSharing(newValue)
                            }
                    }
                    .padding(.horizontal)
                    
                    Spacer(minLength: 20)
                    
                    // Action Buttons
                    VStack(spacing: 15) {
                        // Connect/Disconnect Button
                        Button(action: {
                            meshManager.toggleConnection()
                            
                            // Also toggle P2P discovery (MultipeerConnectivity + BLE)
                            if meshManager.isConnected {
                                multipeerManager.startMeshDiscovery()
                                bleManager.start() // Start BLE for cross-platform (iOS <-> Android)
                            } else {
                                multipeerManager.stopMeshDiscovery()
                                bleManager.stop()
                            }
                        }) {
                            HStack {
                                Image(systemName: meshManager.isConnected ? "wifi.slash" : "wifi")
                                Text(meshManager.isConnected ? "Disconnect from Mesh" : "Connect to Mesh")
                                    .fontWeight(.semibold)
                            }
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(meshManager.isConnected ? Color.red : Color.blue)
                            .foregroundColor(.white)
                            .cornerRadius(12)
                        }
                        
                        // Request Internet Button
                        Button(action: {
                            meshManager.requestInternetThroughBLEProxy(bleManager: bleManager)
                        }) {
                            HStack {
                                Image(systemName: "globe")
                                Text("Request Internet Access")
                                    .fontWeight(.semibold)
                            }
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(Color.green)
                            .foregroundColor(.white)
                            .cornerRadius(12)
                        }
                        .disabled(!meshManager.isConnected && multipeerManager.connectedPeers.isEmpty && bleManager.connectedPeers.isEmpty)
                        .opacity((meshManager.isConnected || !multipeerManager.connectedPeers.isEmpty || !bleManager.connectedPeers.isEmpty) ? 1.0 : 0.5)
                        
                        // HTTP Proxy Section
                        VStack(alignment: .leading, spacing: 10) {
                            HStack {
                                Image(systemName: "network")
                                    .foregroundColor(.orange)
                                Text("HTTP Proxy")
                                    .font(.headline)
                                Spacer()
                                Circle()
                                    .fill(meshManager.isHTTPProxyRunning ? Color.green : Color.gray)
                                    .frame(width: 10, height: 10)
                            }
                            
                            Text(meshManager.isHTTPProxyRunning 
                                ? "Running on 127.0.0.1:\(meshManager.httpProxyPort)" 
                                : "Not running")
                                .font(.caption)
                                .foregroundColor(meshManager.isHTTPProxyRunning ? .green : .gray)
                            
                            Button(action: {
                                meshManager.toggleHTTPProxy(bleManager: bleManager)
                            }) {
                                HStack {
                                    Image(systemName: meshManager.isHTTPProxyRunning ? "stop.fill" : "play.fill")
                                    Text(meshManager.isHTTPProxyRunning ? "Stop Proxy" : "Start Proxy")
                                        .fontWeight(.semibold)
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(meshManager.isHTTPProxyRunning ? Color.red : Color.orange)
                                .foregroundColor(.white)
                                .cornerRadius(8)
                            }
                            .disabled(bleManager.connectedPeers.isEmpty)
                            .opacity(bleManager.connectedPeers.isEmpty ? 0.5 : 1.0)
                            
                            if meshManager.isHTTPProxyRunning {
                                VStack(alignment: .leading, spacing: 4) {
                                    Text("To use:")
                                        .font(.caption)
                                        .fontWeight(.semibold)
                                    Text("Settings → Wi-Fi → Configure Proxy")
                                        .font(.caption2)
                                        .foregroundColor(.gray)
                                    Text("Server: 127.0.0.1  Port: \(meshManager.httpProxyPort)")
                                        .font(.caption2)
                                        .foregroundColor(.gray)
                                }
                                .padding(.top, 4)
                            }
                        }
                        .padding()
                        .background(Color(.systemGray6))
                        .cornerRadius(12)
                        
                        // Send Test Message (for P2P testing)
                        if !multipeerManager.connectedPeers.isEmpty || !bleManager.connectedPeers.isEmpty {
                            Button(action: {
                                let testMessage = "Hello from \(UIDevice.current.name)!"
                                
                                // Send via MultipeerConnectivity
                                if !multipeerManager.connectedPeers.isEmpty {
                                    multipeerManager.sendMessage(testMessage)
                                }
                                
                                // Send via BLE (to Android devices)
                                if !bleManager.connectedPeers.isEmpty {
                                    bleManager.sendMessage(testMessage)
                                }
                                
                                meshManager.statusMessage = "Test message sent to P2P peers"
                            }) {
                                HStack {
                                    Image(systemName: "paperplane.fill")
                                    Text("Send P2P Test Message")
                                        .fontWeight(.semibold)
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(Color.purple)
                                .foregroundColor(.white)
                                .cornerRadius(12)
                            }
                        }
                    }
                    .padding(.horizontal)
                    .padding(.bottom, 10)
                    
                    // Status Message
                    if !meshManager.statusMessage.isEmpty {
                        Text(meshManager.statusMessage)
                            .font(.caption)
                            .foregroundColor(.gray)
                            .padding(.horizontal)
                            .multilineTextAlignment(.center)
                    }
                    
                    // Last received messages
                    VStack(spacing: 8) {
                        if !multipeerManager.lastReceivedMessage.isEmpty {
                            VStack(alignment: .leading, spacing: 5) {
                                Text("Last MC Message:")
                                    .font(.caption)
                                    .foregroundColor(.gray)
                                Text(multipeerManager.lastReceivedMessage)
                                    .font(.caption)
                                    .foregroundColor(.purple)
                                    .padding(8)
                                    .background(Color.purple.opacity(0.1))
                                    .cornerRadius(8)
                            }
                        }
                        
                        if !bleManager.lastReceivedMessage.isEmpty {
                            VStack(alignment: .leading, spacing: 5) {
                                Text("Last BLE Message:")
                                    .font(.caption)
                                    .foregroundColor(.gray)
                                Text(bleManager.lastReceivedMessage)
                                    .font(.caption)
                                    .foregroundColor(.blue)
                                    .padding(8)
                                    .background(Color.blue.opacity(0.1))
                                    .cornerRadius(8)
                            }
                        }
                    }
                    .padding(.horizontal)
                    
                    Spacer(minLength: 30)
                }
            }
            .navigationBarHidden(true)
        }
        .sheet(isPresented: $showPeerList) {
            PeerListView(multipeerManager: multipeerManager, bleManager: bleManager)
        }
        .alert("Error", isPresented: $meshManager.showError) {
            Button("OK", role: .cancel) { }
        } message: {
            Text(meshManager.errorMessage)
        }
        .alert("Success", isPresented: $meshManager.showSuccess) {
            Button("OK", role: .cancel) { }
        } message: {
            Text(meshManager.successMessage)
        }
        .onAppear {
            setupMultipeerCallbacks()
            setupBLECallbacks()
        }
        .onDisappear {
            multipeerManager.cleanup()
            bleManager.stop()
        }
    }
    
    private func setupMultipeerCallbacks() {
        multipeerManager.onPeerDiscovered = { peer in
            meshManager.statusMessage = "Found MC peer: \(peer.displayName)"
        }
        
        multipeerManager.onPeerConnected = { peer in
            meshManager.statusMessage = "Connected to MC peer: \(peer.displayName)"
        }
        
        multipeerManager.onPeerDisconnected = { name in
            meshManager.statusMessage = "MC peer disconnected: \(name)"
        }
        
        multipeerManager.onMessageReceived = { from, data in
            if let message = String(data: data, encoding: .utf8) {
                meshManager.statusMessage = "MC message from \(from): \(message)"
            }
        }
        
        multipeerManager.onError = { error in
            meshManager.errorMessage = error
            meshManager.showError = true
        }
    }
    
    private func setupBLECallbacks() {
        // Set up internet status callback
        bleManager.getInternetStatus = {
            return meshManager.mobileApp?.hasAnyInternet() ?? false
        }
        
        bleManager.onPeerDiscovered = { peer in
            meshManager.statusMessage = "Found BLE peer: \(peer.name) (\(peer.platform))\(peer.hasInternet ? " [Proxy]" : "")"
        }
        
        bleManager.onPeerConnected = { peer in
            meshManager.statusMessage = "Connected to BLE peer: \(peer.name) (\(peer.platform))"
            
            // Register as proxy if peer has internet
            if peer.hasInternet, let app = meshManager.mobileApp {
                let bleIP = "169.254.1.\(abs(peer.id.hashValue) % 255)"
                app.registerBLEProxy(peer.id, peerIP: bleIP, peerMAC: peer.identifier.uuidString, hasInternet: peer.hasInternet)
                print("Registered BLE proxy: \(peer.name)")
            }
        }
        
        bleManager.onPeerDisconnected = { id in
            meshManager.statusMessage = "BLE peer disconnected: \(id)"
            
            // Unregister proxy
            if let app = meshManager.mobileApp {
                app.unregisterBLEProxy(id)
                print("Unregistered BLE proxy: \(id)")
            }
        }
        
        bleManager.onMessageReceived = { from, data in
            if let message = String(data: data, encoding: .utf8) {
                // Check if this is an HTTP tunnel request (from device using proxy)
                if message.contains("\"method\"") && message.contains("\"url\"") && message.contains("\"id\"") {
                    // This device has internet - execute the tunnel request
                if let app = meshManager.mobileApp, app.hasAnyInternet() {
                        DispatchQueue.global(qos: .userInitiated).async {
                            // Execute the HTTP tunnel request
                            var execError: NSError?
                            let responseJSON = app.executeTunnelRequest(message, error: &execError)
                            
                            if let error = execError {
                                print("Failed to execute tunnel request: \(error)")
                                return
                            }
                                
                            // Send response back via BLE
                            if let responseData = responseJSON.data(using: .utf8),
                               let peer = bleManager.connectedPeers.first(where: { $0.id == from }) {
                                _ = bleManager.sendData(responseData, to: peer.identifier)
                                print("Sent tunnel response to \(from)")
                            }
                        }
                    }
                }
                // Check if this is a tunnel response (from device with internet)
                else if message.contains("\"status_code\"") && message.contains("\"id\"") {
                    // Forward to HTTP proxy handler
                    if let app = meshManager.mobileApp {
                        do {
                            try app.handleTunnelResponse(message)
                            print("Handled tunnel response")
                        } catch {
                            print("Failed to handle tunnel response: \(error)")
                        }
                    }
                }
                // Handle via Go proxy message handler for other cases
                else if message.starts(with: "{") {
                    if let app = meshManager.mobileApp {
                        do {
                            try app.handleBLEProxyMessage(from, data: data)
                            print("Handled BLE proxy message from \(from)")
                        } catch {
                            print("Failed to handle BLE proxy message: \(error)")
                        }
                    }
                } else {
                    // Regular message
                    meshManager.statusMessage = "BLE message from \(from): \(message)"
                }
            }
        }
        
        bleManager.onError = { error in
            if error.contains("powered off") {
                meshManager.errorMessage = "Please enable Bluetooth in Settings to connect with Android devices"
            } else {
                meshManager.errorMessage = error
            }
            meshManager.showError = true
        }
    }
}

// MARK: - Peer List View
struct PeerListView: View {
    @ObservedObject var multipeerManager: MultipeerManager
    @ObservedObject var bleManager: BLEManager
    @Environment(\.dismiss) var dismiss
    
    var body: some View {
        NavigationView {
            List {
                // Empty state
                if multipeerManager.discoveredPeers.isEmpty && bleManager.discoveredPeers.isEmpty {
                    VStack(spacing: 10) {
                        Image(systemName: "wifi.exclamationmark")
                            .font(.largeTitle)
                            .foregroundColor(.gray)
                        Text("No peers discovered")
                            .foregroundColor(.gray)
                        Text("Make sure other devices have InterMesh running and are nearby")
                            .font(.caption)
                            .foregroundColor(.gray)
                            .multilineTextAlignment(.center)
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                }
                
                // MultipeerConnectivity Peers (iOS devices)
                if !multipeerManager.discoveredPeers.isEmpty {
                    Section(header: Text("📱 iOS Peers (MultipeerConnectivity)")) {
                        ForEach(multipeerManager.discoveredPeers) { peer in
                            HStack {
                                VStack(alignment: .leading) {
                                    Text(peer.displayName)
                                        .fontWeight(.medium)
                                    
                                    if multipeerManager.connectedPeers.contains(where: { $0.id == peer.id }) {
                                        Text("Connected")
                                            .font(.caption)
                                            .foregroundColor(.green)
                                    }
                                }
                                
                                Spacer()
                                
                                if multipeerManager.connectedPeers.contains(where: { $0.id == peer.id }) {
                                    Image(systemName: "checkmark.circle.fill")
                                        .foregroundColor(.green)
                                } else {
                                    Button("Connect") {
                                        multipeerManager.connect(to: peer)
                                    }
                                    .buttonStyle(.borderedProminent)
                                    .tint(.purple)
                                }
                            }
                        }
                    }
                }
                
                // BLE Peers (Cross-platform - Android devices)
                if !bleManager.discoveredPeers.isEmpty {
                    Section(header: Text("🤖 Cross-Platform Peers (BLE)")) {
                        ForEach(bleManager.discoveredPeers) { peer in
                            HStack {
                                VStack(alignment: .leading) {
                                    HStack {
                                        Text(peer.name)
                                            .fontWeight(.medium)
                                        
                                        // Platform badge
                                        Text(peer.platform)
                                            .font(.caption2)
                                            .padding(.horizontal, 6)
                                            .padding(.vertical, 2)
                                            .background(peer.platform == "android" ? Color.green.opacity(0.2) : Color.blue.opacity(0.2))
                                            .foregroundColor(peer.platform == "android" ? .green : .blue)
                                            .cornerRadius(4)
                                    }
                                    
                                    if peer.isConnected {
                                        Text("Connected")
                                            .font(.caption)
                                            .foregroundColor(.green)
                                    }
                                }
                                
                                Spacer()
                                
                                if peer.isConnected {
                                    Image(systemName: "checkmark.circle.fill")
                                        .foregroundColor(.green)
                                } else {
                                    Button("Connect") {
                                        bleManager.connect(to: peer)
                                    }
                                    .buttonStyle(.borderedProminent)
                                    .tint(.blue)
                                }
                            }
                        }
                    }
                }
                
                // Connected peers summary
                if !multipeerManager.connectedPeers.isEmpty || !bleManager.connectedPeers.isEmpty {
                    Section(header: Text("✅ Connected")) {
                        ForEach(multipeerManager.connectedPeers) { peer in
                            HStack {
                                Image(systemName: "wifi")
                                    .foregroundColor(.purple)
                                Text(peer.displayName)
                                    .fontWeight(.medium)
                                Spacer()
                                Text("MC")
                                    .font(.caption)
                                    .foregroundColor(.purple)
                            }
                        }
                        
                        ForEach(bleManager.connectedPeers) { peer in
                            HStack {
                                Image(systemName: "wave.3.right")
                                    .foregroundColor(.blue)
                                Text("\(peer.name) (\(peer.platform))")
                                    .fontWeight(.medium)
                                Spacer()
                                Text("BLE")
                                    .font(.caption)
                                    .foregroundColor(.blue)
                            }
                        }
                    }
                }
            }
            .navigationTitle("P2P Peers")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }
}

// MeshManager to handle all mesh networking logic
class MeshManager: ObservableObject {
    @Published var isConnected = false
    @Published var isSharingInternet = false
    @Published var peerCount: Int = 0
    @Published var proxyCount: Int = 0
    @Published var deviceID: String = ""
    @Published var statusMessage: String = ""
    @Published var showError: Bool = false
    @Published var errorMessage: String = ""
    @Published var showSuccess: Bool = false
    @Published var successMessage: String = ""
    @Published var isHTTPProxyRunning = false
    @Published var httpProxyPort: Int = 8080
    
    var mobileApp: IntermeshMobileApp?  // Expose as public property
    private var statsTimer: Timer?
    
    init() {
        setupMeshApp()
    }
    
    deinit {
        statsTimer?.invalidate()
    }
    
    private func setupMeshApp() {
        // Generate a unique device ID
        deviceID = UIDevice.current.identifierForVendor?.uuidString ?? UUID().uuidString
        let deviceName = UIDevice.current.name
        let ipAddress = getIPAddress() ?? "0.0.0.0"
        let macAddress = getMACAddress() ?? "00:00:00:00:00:00"
        
        // Initialize the mobile app
        mobileApp = IntermeshNewMobileApp(deviceID, deviceName, ipAddress, macAddress)
        statusMessage = "Ready to connect"
    }
    
    func toggleConnection() {
        guard let app = mobileApp else { return }
        
        if isConnected {
            // Disconnect
            app.stop()
            isConnected = false
            peerCount = 0
            proxyCount = 0
            statusMessage = "Disconnected from mesh"
            stopUpdatingStats()
        } else {
            // Connect
            do {
                try app.start()
                try app.connectToNetwork()
                isConnected = true
                statusMessage = "Connected! Searching for peers..."
                startUpdatingStats()
            } catch {
                errorMessage = "Failed to connect: \(error.localizedDescription)"
                showError = true
            }
        }
    }
    
    func toggleInternetSharing(_ enabled: Bool) {
        guard let app = mobileApp else { return }
        
        if enabled {
            do {
                try app.enableInternetSharing()
                statusMessage = "Internet sharing enabled"
            } catch {
                errorMessage = "Failed to enable sharing: \(error.localizedDescription)"
                showError = true
                isSharingInternet = false
            }
        } else {
            app.disableInternetSharing()
            statusMessage = "Internet sharing disabled"
        }
    }
    
    func requestInternetAccess() {
        guard let app = mobileApp else { return }
        
        var error: NSError?
        let result = app.requestInternetAccess(&error)
        if let error = error {
            errorMessage = error.localizedDescription
            showError = true
        } else if result.hasPrefix("Error") || result.hasPrefix("error") {
            errorMessage = result
            showError = true
        } else {
            successMessage = result
            showSuccess = true
            statusMessage = "Connected to internet proxy"
        }
    }
    
    func toggleHTTPProxy(bleManager: BLEManager) {
        guard let app = mobileApp else {
            errorMessage = "Mobile app not initialized"
            showError = true
            return
        }
        
        if isHTTPProxyRunning {
            // Stop the proxy
            app.stopHTTPProxy()
            isHTTPProxyRunning = false
            statusMessage = "HTTP Proxy stopped"
        } else {
            // Set up the BLE message sender before starting proxy
            let callback = BLEMessageSender(bleManager: bleManager)
            app.setSimpleBLEMessageSender(callback)
            
            // Start the proxy
            do {
                try app.startHTTPProxy(Int64(httpProxyPort))
                isHTTPProxyRunning = true
                statusMessage = "HTTP Proxy running on port \(httpProxyPort)"
            } catch {
                errorMessage = "Failed to start HTTP proxy: \(error.localizedDescription)"
                showError = true
            }
        }
    }
    
    private func startUpdatingStats() {
        statsTimer?.invalidate()
        statsTimer = Timer.scheduledTimer(withTimeInterval: 2.0, repeats: true) { [weak self] timer in
            guard let self = self, let app = self.mobileApp, self.isConnected else {
                timer.invalidate()
                return
            }
            
            if let stats = app.getNetworkStats() {
                DispatchQueue.main.async {
                    self.peerCount = Int(stats.peerCount)
                    self.proxyCount = Int(stats.availableProxies)
                }
            }
        }
    }
    
    private func stopUpdatingStats() {
        statsTimer?.invalidate()
        statsTimer = nil
    }
    
    private func getIPAddress() -> String? {
        var address: String?
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        
        if getifaddrs(&ifaddr) == 0 {
            var ptr = ifaddr
            while ptr != nil {
                defer { ptr = ptr?.pointee.ifa_next }
                
                guard let interface = ptr?.pointee,
                      let addr = interface.ifa_addr else {
                    continue
                }
                
                let addrFamily = addr.pointee.sa_family
                
                if addrFamily == UInt8(AF_INET) || addrFamily == UInt8(AF_INET6) {
                    guard let namePtr = interface.ifa_name else {
                        continue
                    }
                    let name = String(cString: namePtr)
                    if name == "en0" {
                        var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                        let saLen = socklen_t(addr.pointee.sa_len)
                        getnameinfo(addr,
                                    saLen,
                                    &hostname, socklen_t(hostname.count),
                                    nil, socklen_t(0),
                                    NI_NUMERICHOST)
                        address = String(cString: hostname)
                    }
                }
            }
            freeifaddrs(ifaddr)
        }
        
        return address
    }
    
    private func getMACAddress() -> String? {
        // iOS doesn't allow direct MAC address access for privacy reasons
        return "00:00:00:00:00:00"
    }
    
    func requestInternetThroughBLEProxy(bleManager: BLEManager) {
        guard let app = mobileApp else {
            errorMessage = "Mobile app not initialized"
            return
        }
        
        // Find a BLE peer with internet
        let proxyPeer = bleManager.connectedPeers.first { $0.hasInternet }
        guard let proxy = proxyPeer else {
            errorMessage = "No BLE proxies available"
            return
        }
        
        statusMessage = "Requesting internet through \(proxy.name)..."
        
        // Create and send a proxy request via BLE
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            guard let self = self else { return }
            
            // Create a proxy request JSON
            let testURL = "https://httpbin.org/ip"
            var error: NSError?
            let requestJSON = app.createProxyRequest(testURL, method: "GET", error: &error)
            
            if let error = error {
                DispatchQueue.main.async {
                    self.errorMessage = "Failed to create proxy request: \(error.localizedDescription)"
                }
                return
            }
            
            guard let requestData = requestJSON.data(using: .utf8) else {
                DispatchQueue.main.async {
                    self.errorMessage = "Failed to encode proxy request"
                }
                return
            }
                
            // Send the request via BLE to the proxy peer
            let sent = bleManager.sendData(requestData, to: proxy.identifier)
            
            DispatchQueue.main.async {
                if sent {
                    self.statusMessage = "Request sent to \(proxy.name), waiting for response..."
                } else {
                    self.errorMessage = "Failed to send request via BLE"
                }
            }
        }
    }
}

// BLEMessageSender implements IntermeshBLEMessageCallback for sending messages via BLE
class BLEMessageSender: NSObject, IntermeshBLEMessageCallbackProtocol {
    private weak var bleManager: BLEManager?
    
    init(bleManager: BLEManager) {
        self.bleManager = bleManager
        super.init()
    }
    
    func sendMessage(_ message: String?) -> Bool {
        guard let manager = bleManager, let msg = message else { return false }
        guard let data = msg.data(using: .utf8) else { return false }
        
        // Send to first connected BLE peer with internet
        if let peer = manager.connectedPeers.first(where: { $0.hasInternet }) {
            return manager.sendData(data, to: peer.identifier)
        }
        // If no peer with internet, send to any connected peer
        if let peer = manager.connectedPeers.first {
            return manager.sendData(data, to: peer.identifier)
        }
        return false
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
