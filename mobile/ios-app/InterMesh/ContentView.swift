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
                        
                        // P2P Status indicator
                        if multipeerManager.isBrowsing {
                            HStack(spacing: 4) {
                                Circle()
                                    .fill(Color.green)
                                    .frame(width: 8, height: 8)
                                Text("P2P")
                                    .font(.caption)
                                    .foregroundColor(.green)
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
                            Text(multipeerManager.connectionStatus)
                                .foregroundColor(multipeerManager.connectedPeers.isEmpty ? .orange : .green)
                                .fontWeight(.medium)
                                .font(.caption)
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
                            
                            // P2P Peers (tappable)
                            Button(action: { showPeerList = true }) {
                                VStack(alignment: .leading) {
                                    Text("\(multipeerManager.discoveredPeers.count)")
                                        .font(.system(size: 32, weight: .bold))
                                        .foregroundColor(.purple)
                                    Text("P2P Peers")
                                        .font(.caption)
                                        .foregroundColor(.gray)
                                }
                            }
                            .disabled(multipeerManager.discoveredPeers.isEmpty)
                            
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
                        if !multipeerManager.connectedPeers.isEmpty {
                            Divider()
                            HStack {
                                Image(systemName: "wifi")
                                    .foregroundColor(.green)
                                Text("Connected via P2P: \(multipeerManager.connectedPeers.map { $0.displayName }.joined(separator: ", "))")
                                    .font(.caption)
                                    .foregroundColor(.green)
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
                            
                            // Also toggle P2P discovery
                            if meshManager.isConnected {
                                multipeerManager.startMeshDiscovery()
                            } else {
                                multipeerManager.stopMeshDiscovery()
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
                            meshManager.requestInternetAccess()
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
                        .disabled(!meshManager.isConnected && multipeerManager.connectedPeers.isEmpty)
                        .opacity((meshManager.isConnected || !multipeerManager.connectedPeers.isEmpty) ? 1.0 : 0.5)
                        
                        // Send Test Message (for P2P testing)
                        if !multipeerManager.connectedPeers.isEmpty {
                            Button(action: {
                                multipeerManager.sendMessage("Hello from \(UIDevice.current.name)!")
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
                    
                    // Last received P2P message
                    if !multipeerManager.lastReceivedMessage.isEmpty {
                        VStack(alignment: .leading, spacing: 5) {
                            Text("Last P2P Message:")
                                .font(.caption)
                                .foregroundColor(.gray)
                            Text(multipeerManager.lastReceivedMessage)
                                .font(.caption)
                                .foregroundColor(.purple)
                                .padding(8)
                                .background(Color.purple.opacity(0.1))
                                .cornerRadius(8)
                        }
                        .padding(.horizontal)
                    }
                    
                    Spacer(minLength: 30)
                }
            }
            .navigationBarHidden(true)
        }
        .sheet(isPresented: $showPeerList) {
            PeerListView(multipeerManager: multipeerManager)
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
        }
        .onDisappear {
            multipeerManager.cleanup()
        }
    }
    
    private func setupMultipeerCallbacks() {
        multipeerManager.onPeerDiscovered = { peer in
            meshManager.statusMessage = "Found P2P peer: \(peer.displayName)"
        }
        
        multipeerManager.onPeerConnected = { peer in
            meshManager.statusMessage = "Connected to P2P peer: \(peer.displayName)"
        }
        
        multipeerManager.onPeerDisconnected = { name in
            meshManager.statusMessage = "P2P peer disconnected: \(name)"
        }
        
        multipeerManager.onMessageReceived = { from, data in
            if let message = String(data: data, encoding: .utf8) {
                meshManager.statusMessage = "Message from \(from): \(message)"
            }
        }
        
        multipeerManager.onError = { error in
            meshManager.errorMessage = error
            meshManager.showError = true
        }
    }
}

// MARK: - Peer List View

struct PeerListView: View {
    @ObservedObject var multipeerManager: MultipeerManager
    @Environment(\.dismiss) var dismiss
    
    var body: some View {
        NavigationView {
            List {
                if multipeerManager.discoveredPeers.isEmpty {
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
                } else {
                    Section(header: Text("Discovered Peers")) {
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
                                    .tint(.blue)
                                }
                            }
                        }
                    }
                }
                
                if !multipeerManager.connectedPeers.isEmpty {
                    Section(header: Text("Connected Peers")) {
                        ForEach(multipeerManager.connectedPeers) { peer in
                            HStack {
                                Image(systemName: "wifi")
                                    .foregroundColor(.green)
                                Text(peer.displayName)
                                    .fontWeight(.medium)
                                Spacer()
                                Text("Connected")
                                    .font(.caption)
                                    .foregroundColor(.green)
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
    
    private var mobileApp: IntermeshMobileApp?
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
                app.connectToNetwork()
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
        
        let result = app.requestInternetAccess()
        if result.hasPrefix("Error") || result.hasPrefix("error") {
            errorMessage = result
            showError = true
        } else {
            successMessage = result
            showSuccess = true
            statusMessage = "Connected to internet proxy"
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
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
