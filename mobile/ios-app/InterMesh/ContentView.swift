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
    
    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                // Header
                Text("InterMesh")
                    .font(.largeTitle)
                    .fontWeight(.bold)
                    .padding(.top, 20)
                
                // Device Info Card
                VStack(alignment: .leading, spacing: 10) {
                    Text("Device Information")
                        .font(.headline)
                    
                    HStack {
                        Text("Device ID:")
                            .foregroundColor(.gray)
                        Spacer()
                        Text(meshManager.deviceID)
                            .fontWeight(.medium)
                    }
                    
                    HStack {
                        Text("Status:")
                            .foregroundColor(.gray)
                        Spacer()
                        Text(meshManager.isConnected ? "Connected" : "Disconnected")
                            .foregroundColor(meshManager.isConnected ? .green : .red)
                            .fontWeight(.medium)
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
                        VStack(alignment: .leading) {
                            Text("\(meshManager.peerCount)")
                                .font(.system(size: 32, weight: .bold))
                            Text("Connected Peers")
                                .font(.caption)
                                .foregroundColor(.gray)
                        }
                        
                        Spacer()
                        
                        VStack(alignment: .leading) {
                            Text("\(meshManager.proxyCount)")
                                .font(.system(size: 32, weight: .bold))
                            Text("Available Proxies")
                                .font(.caption)
                                .foregroundColor(.gray)
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
                
                Spacer()
                
                // Action Buttons
                VStack(spacing: 15) {
                    // Connect/Disconnect Button
                    Button(action: {
                        meshManager.toggleConnection()
                    }) {
                        Text(meshManager.isConnected ? "Disconnect from Mesh" : "Connect to Mesh")
                            .fontWeight(.semibold)
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
                        Text("Request Internet Access")
                            .fontWeight(.semibold)
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(Color.green)
                            .foregroundColor(.white)
                            .cornerRadius(12)
                    }
                    .disabled(!meshManager.isConnected)
                    .opacity(meshManager.isConnected ? 1.0 : 0.5)
                }
                .padding(.horizontal)
                .padding(.bottom, 30)
                
                // Status Message
                if !meshManager.statusMessage.isEmpty {
                    Text(meshManager.statusMessage)
                        .font(.caption)
                        .foregroundColor(.gray)
                        .padding(.horizontal)
                        .padding(.bottom, 10)
                }
            }
            .navigationBarHidden(true)
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
    
    init() {
        setupMeshApp()
    }
    
    private func setupMeshApp() {
        // Generate a unique device ID
        deviceID = UIDevice.current.identifierForVendor?.uuidString ?? UUID().uuidString
        let deviceName = UIDevice.current.name
        let ipAddress = getIPAddress() ?? "0.0.0.0"
        let macAddress = getMACAddress() ?? "00:00:00:00:00:00"
        
        // Initialize the mobile app
        do {
            mobileApp = IntermeshNewMobileApp(deviceID, deviceName, ipAddress, macAddress)
            statusMessage = "App initialized"
        } catch {
            errorMessage = "Failed to initialize app: \(error.localizedDescription)"
            showError = true
        }
    }
    
    func toggleConnection() {
        guard let app = mobileApp else { return }
        
        if isConnected {
            // Disconnect
            app.disconnectFromNetwork()
            isConnected = false
            peerCount = 0
            proxyCount = 0
            statusMessage = "Disconnected from mesh network"
        } else {
            // Connect
            do {
                try app.connectToNetwork()
                isConnected = true
                statusMessage = "Connected to mesh network"
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
        
        do {
            let result = try app.requestInternetAccess()
            successMessage = result
            showSuccess = true
            statusMessage = "Connected to internet proxy"
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }
    
    private func startUpdatingStats() {
        Timer.scheduledTimer(withTimeInterval: 2.0, repeats: true) { [weak self] timer in
            guard let self = self, let app = self.mobileApp, self.isConnected else {
                timer.invalidate()
                return
            }
            
            self.peerCount = Int(app.getConnectedPeerCount())
            self.proxyCount = Int(app.getAvailableProxyCount())
        }
    }
    
    private func getIPAddress() -> String? {
        var address: String?
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        
        if getifaddrs(&ifaddr) == 0 {
            var ptr = ifaddr
            while ptr != nil {
                defer { ptr = ptr?.pointee.ifa_next }
                
                let interface = ptr?.pointee
                let addrFamily = interface?.ifa_addr.pointee.sa_family
                
                if addrFamily == UInt8(AF_INET) || addrFamily == UInt8(AF_INET6) {
                    let name = String(cString: (interface?.ifa_name)!)
                    if name == "en0" {
                        var hostname = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                        getnameinfo(interface?.ifa_addr, socklen_t((interface?.ifa_addr.pointee.sa_len)!),
                                  &hostname, socklen_t(hostname.count),
                                  nil, socklen_t(0), NI_NUMERICHOST)
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
        // Return a placeholder or generate a unique identifier
        return "00:00:00:00:00:00"
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
