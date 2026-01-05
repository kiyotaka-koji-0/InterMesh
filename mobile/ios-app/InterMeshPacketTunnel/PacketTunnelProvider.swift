import NetworkExtension
import Intermesh

class PacketTunnelProvider: NEPacketTunnelProvider {

    override func startTunnel(options: [String : NSObject]?, completionHandler: @escaping (Error?) -> Void) {
        // 1. Create tunnel settings
        let settings = NEPacketTunnelNetworkSettings(tunnelRemoteAddress: "127.0.0.1")
        
        // 2. Configure IPv4 settings (local virtual network)
        settings.ipv4Settings = NEIPv4Settings(addresses: ["10.0.0.2"], subnetMasks: ["255.255.255.0"])
        settings.ipv4Settings?.includedRoutes = [NEIPv4Route.default()]
        
        // 3. Configure DNS
        settings.dnsSettings = NEDNSSettings(servers: ["8.8.8.8", "8.8.4.4"])
        
        // 4. Configure HTTP Proxy for the Tunnel
        // This tells iOS to route HTTP traffic through our local InterMesh proxy
        let proxySettings = NEProxySettings()
        proxySettings.httpServer = NEProxyServer(address: "127.0.0.1", port: 8080)
        proxySettings.httpsServer = NEProxyServer(address: "127.0.0.1", port: 8080)
        proxySettings.autoProxyConfigurationEnabled = false
        proxySettings.httpEnabled = true
        proxySettings.httpsEnabled = true
        proxySettings.excludeSimpleHostnames = true
        proxySettings.exceptionList = ["localhost", "127.0.0.1"]
        
        settings.proxySettings = proxySettings
        
        // 5. Apply settings
        setTunnelNetworkSettings(settings) { error in
            if let error = error {
                completionHandler(error)
                return
            }
            completionHandler(nil)
        }
    }
    
    override func stopTunnel(with reason: NEProviderStopReason, completionHandler: @escaping () -> Void) {
        completionHandler()
    }
    
    override func handleAppMessage(_ messageData: Data, completionHandler: ((Data?) -> Void)? = nil) {
        if let handler = completionHandler {
            handler(messageData)
        }
    }
    
    override func sleep(completionHandler: @escaping () -> Void) {
        completionHandler()
    }
    
    override func wake() {
        // Restore connection if needed
    }
}
