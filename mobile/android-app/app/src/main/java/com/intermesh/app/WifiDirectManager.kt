package com.intermesh.app

import android.annotation.SuppressLint
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.net.NetworkInfo
import android.net.wifi.WpsInfo
import android.net.wifi.p2p.*
import android.os.Build
import android.util.Log
import java.net.InetAddress
import java.net.ServerSocket
import java.net.Socket
import java.util.concurrent.ConcurrentHashMap
import kotlin.concurrent.thread

/**
 * WiFi Direct Manager for peer-to-peer mesh networking
 * Allows devices to discover and connect to each other without a WiFi router
 */
class WifiDirectManager(private val context: Context) {
    
    companion object {
        private const val TAG = "WifiDirectManager"
        private const val MESH_SERVICE_PORT = 9998
    }
    
    private var wifiP2pManager: WifiP2pManager? = null
    private var channel: WifiP2pManager.Channel? = null
    private var receiver: WifiDirectReceiver? = null
    
    private val discoveredPeers = ConcurrentHashMap<String, WifiP2pDevice>()
    private var isWifiP2pEnabled = false
    private var isDiscovering = false
    private var isConnected = false
    private var groupOwnerAddress: InetAddress? = null
    private var isGroupOwner = false
    
    // Callbacks
    var onPeerDiscovered: ((PeerInfo) -> Unit)? = null
    var onPeerLost: ((String) -> Unit)? = null
    var onConnected: ((String, Boolean) -> Unit)? = null  // address, isGroupOwner
    var onDisconnected: (() -> Unit)? = null
    var onMessageReceived: ((String, ByteArray) -> Unit)? = null
    var onError: ((String) -> Unit)? = null
    
    // Server socket for receiving connections (when group owner)
    private var serverSocket: ServerSocket? = null
    private var isServerRunning = false
    
    data class PeerInfo(
        val deviceAddress: String,
        val deviceName: String,
        val status: Int,
        val isGroupOwner: Boolean
    )
    
    /**
     * Initialize the WiFi Direct manager
     */
    fun initialize(): Boolean {
        wifiP2pManager = context.getSystemService(Context.WIFI_P2P_SERVICE) as? WifiP2pManager
        if (wifiP2pManager == null) {
            Log.e(TAG, "WiFi P2P is not supported on this device")
            return false
        }
        
        channel = wifiP2pManager?.initialize(context, context.mainLooper) { 
            Log.w(TAG, "WiFi P2P channel disconnected")
            cleanup()
        }
        
        if (channel == null) {
            Log.e(TAG, "Failed to initialize WiFi P2P channel")
            return false
        }
        
        // Register broadcast receiver
        receiver = WifiDirectReceiver()
        val intentFilter = IntentFilter().apply {
            addAction(WifiP2pManager.WIFI_P2P_STATE_CHANGED_ACTION)
            addAction(WifiP2pManager.WIFI_P2P_PEERS_CHANGED_ACTION)
            addAction(WifiP2pManager.WIFI_P2P_CONNECTION_CHANGED_ACTION)
            addAction(WifiP2pManager.WIFI_P2P_THIS_DEVICE_CHANGED_ACTION)
        }
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            context.registerReceiver(receiver, intentFilter, Context.RECEIVER_NOT_EXPORTED)
        } else {
            context.registerReceiver(receiver, intentFilter)
        }
        
        Log.d(TAG, "WiFi Direct manager initialized")
        return true
    }
    
    /**
     * Start discovering nearby WiFi Direct peers
     */
    @SuppressLint("MissingPermission")
    fun startDiscovery() {
        if (!isWifiP2pEnabled) {
            onError?.invoke("WiFi Direct is not enabled")
            return
        }
        
        if (isDiscovering) {
            Log.d(TAG, "Already discovering peers")
            return
        }
        
        wifiP2pManager?.discoverPeers(channel, object : WifiP2pManager.ActionListener {
            override fun onSuccess() {
                isDiscovering = true
                Log.d(TAG, "Peer discovery started")
            }
            
            override fun onFailure(reason: Int) {
                isDiscovering = false
                val errorMsg = getErrorMessage(reason)
                Log.e(TAG, "Peer discovery failed: $errorMsg")
                onError?.invoke("Discovery failed: $errorMsg")
            }
        })
    }
    
    /**
     * Stop discovering peers
     */
    @SuppressLint("MissingPermission")
    fun stopDiscovery() {
        if (!isDiscovering) return
        
        wifiP2pManager?.stopPeerDiscovery(channel, object : WifiP2pManager.ActionListener {
            override fun onSuccess() {
                isDiscovering = false
                Log.d(TAG, "Peer discovery stopped")
            }
            
            override fun onFailure(reason: Int) {
                Log.e(TAG, "Failed to stop discovery: ${getErrorMessage(reason)}")
            }
        })
    }
    
    /**
     * Connect to a discovered peer
     */
    @SuppressLint("MissingPermission")
    fun connectToPeer(deviceAddress: String) {
        val config = WifiP2pConfig().apply {
            this.deviceAddress = deviceAddress
            wps.setup = WpsInfo.PBC  // Push Button Configuration
            groupOwnerIntent = 0  // Prefer to be client (let other device be group owner)
        }
        
        wifiP2pManager?.connect(channel, config, object : WifiP2pManager.ActionListener {
            override fun onSuccess() {
                Log.d(TAG, "Connection initiated to $deviceAddress")
            }
            
            override fun onFailure(reason: Int) {
                val errorMsg = getErrorMessage(reason)
                Log.e(TAG, "Connection failed: $errorMsg")
                onError?.invoke("Connection failed: $errorMsg")
            }
        })
    }
    
    /**
     * Disconnect from current peer
     */
    @SuppressLint("MissingPermission")
    fun disconnect() {
        wifiP2pManager?.removeGroup(channel, object : WifiP2pManager.ActionListener {
            override fun onSuccess() {
                Log.d(TAG, "Disconnected from group")
                isConnected = false
                stopServer()
            }
            
            override fun onFailure(reason: Int) {
                Log.e(TAG, "Disconnect failed: ${getErrorMessage(reason)}")
            }
        })
    }
    
    /**
     * Send data to a connected peer
     */
    fun sendData(data: ByteArray) {
        if (!isConnected) {
            onError?.invoke("Not connected to any peer")
            return
        }
        
        val targetAddress = groupOwnerAddress
        if (targetAddress == null) {
            onError?.invoke("No peer address available")
            return
        }
        
        thread {
            try {
                Socket(targetAddress, MESH_SERVICE_PORT).use { socket ->
                    socket.getOutputStream().write(data)
                    Log.d(TAG, "Sent ${data.size} bytes to peer")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Failed to send data: ${e.message}")
                onError?.invoke("Send failed: ${e.message}")
            }
        }
    }
    
    /**
     * Start server socket for receiving connections (when group owner)
     */
    private fun startServer() {
        if (isServerRunning) return
        
        thread {
            try {
                serverSocket = ServerSocket(MESH_SERVICE_PORT)
                isServerRunning = true
                Log.d(TAG, "Server started on port $MESH_SERVICE_PORT")
                
                while (isServerRunning) {
                    try {
                        val client = serverSocket?.accept() ?: break
                        thread {
                            handleClient(client)
                        }
                    } catch (e: Exception) {
                        if (isServerRunning) {
                            Log.e(TAG, "Server accept error: ${e.message}")
                        }
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "Failed to start server: ${e.message}")
            }
        }
    }
    
    private fun handleClient(socket: Socket) {
        try {
            val clientAddress = socket.inetAddress.hostAddress ?: "unknown"
            val data = socket.getInputStream().readBytes()
            Log.d(TAG, "Received ${data.size} bytes from $clientAddress")
            onMessageReceived?.invoke(clientAddress, data)
        } catch (e: Exception) {
            Log.e(TAG, "Error handling client: ${e.message}")
        } finally {
            socket.close()
        }
    }
    
    private fun stopServer() {
        isServerRunning = false
        try {
            serverSocket?.close()
            serverSocket = null
        } catch (e: Exception) {
            Log.e(TAG, "Error stopping server: ${e.message}")
        }
    }
    
    /**
     * Get list of discovered peers
     */
    fun getDiscoveredPeers(): List<PeerInfo> {
        return discoveredPeers.values.map { device ->
            PeerInfo(
                deviceAddress = device.deviceAddress,
                deviceName = device.deviceName ?: "Unknown",
                status = device.status,
                isGroupOwner = device.isGroupOwner
            )
        }
    }
    
    /**
     * Check if WiFi Direct is enabled
     */
    fun isEnabled(): Boolean = isWifiP2pEnabled
    
    /**
     * Check if currently connected to a peer
     */
    fun isConnectedToPeer(): Boolean = isConnected
    
    /**
     * Clean up resources
     */
    fun cleanup() {
        stopDiscovery()
        disconnect()
        stopServer()
        
        try {
            receiver?.let { context.unregisterReceiver(it) }
        } catch (e: Exception) {
            Log.e(TAG, "Error unregistering receiver: ${e.message}")
        }
        
        discoveredPeers.clear()
        receiver = null
        channel = null
        wifiP2pManager = null
    }
    
    private fun getErrorMessage(reason: Int): String {
        return when (reason) {
            WifiP2pManager.ERROR -> "Internal error"
            WifiP2pManager.P2P_UNSUPPORTED -> "P2P not supported"
            WifiP2pManager.BUSY -> "System busy"
            WifiP2pManager.NO_SERVICE_REQUESTS -> "No service requests"
            else -> "Unknown error ($reason)"
        }
    }
    
    /**
     * Broadcast receiver for WiFi Direct events
     */
    private inner class WifiDirectReceiver : BroadcastReceiver() {
        
        @SuppressLint("MissingPermission")
        override fun onReceive(context: Context, intent: Intent) {
            when (intent.action) {
                WifiP2pManager.WIFI_P2P_STATE_CHANGED_ACTION -> {
                    val state = intent.getIntExtra(WifiP2pManager.EXTRA_WIFI_STATE, -1)
                    isWifiP2pEnabled = state == WifiP2pManager.WIFI_P2P_STATE_ENABLED
                    Log.d(TAG, "WiFi P2P state: ${if (isWifiP2pEnabled) "enabled" else "disabled"}")
                    
                    if (!isWifiP2pEnabled) {
                        discoveredPeers.clear()
                        onError?.invoke("WiFi Direct is disabled")
                    }
                }
                
                WifiP2pManager.WIFI_P2P_PEERS_CHANGED_ACTION -> {
                    // Request updated peer list
                    wifiP2pManager?.requestPeers(channel) { peers ->
                        handlePeersChanged(peers)
                    }
                }
                
                WifiP2pManager.WIFI_P2P_CONNECTION_CHANGED_ACTION -> {
                    val networkInfo = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                        intent.getParcelableExtra(WifiP2pManager.EXTRA_NETWORK_INFO, NetworkInfo::class.java)
                    } else {
                        @Suppress("DEPRECATION")
                        intent.getParcelableExtra(WifiP2pManager.EXTRA_NETWORK_INFO)
                    }
                    
                    if (networkInfo?.isConnected == true) {
                        // Request connection info
                        wifiP2pManager?.requestConnectionInfo(channel) { info ->
                            handleConnectionInfo(info)
                        }
                    } else {
                        isConnected = false
                        groupOwnerAddress = null
                        stopServer()
                        onDisconnected?.invoke()
                        Log.d(TAG, "Disconnected from peer")
                    }
                }
                
                WifiP2pManager.WIFI_P2P_THIS_DEVICE_CHANGED_ACTION -> {
                    val device = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                        intent.getParcelableExtra(WifiP2pManager.EXTRA_WIFI_P2P_DEVICE, WifiP2pDevice::class.java)
                    } else {
                        @Suppress("DEPRECATION")
                        intent.getParcelableExtra(WifiP2pManager.EXTRA_WIFI_P2P_DEVICE)
                    }
                    Log.d(TAG, "This device: ${device?.deviceName} (${device?.deviceAddress})")
                }
            }
        }
        
        private fun handlePeersChanged(peers: WifiP2pDeviceList) {
            val currentPeers = peers.deviceList.associateBy { it.deviceAddress }
            
            // Find new peers
            currentPeers.forEach { (address, device) ->
                if (!discoveredPeers.containsKey(address)) {
                    discoveredPeers[address] = device
                    val peerInfo = PeerInfo(
                        deviceAddress = device.deviceAddress,
                        deviceName = device.deviceName ?: "Unknown",
                        status = device.status,
                        isGroupOwner = device.isGroupOwner
                    )
                    Log.d(TAG, "Discovered peer: ${device.deviceName} (${device.deviceAddress})")
                    onPeerDiscovered?.invoke(peerInfo)
                }
            }
            
            // Find lost peers
            val lostPeers = discoveredPeers.keys - currentPeers.keys
            lostPeers.forEach { address ->
                discoveredPeers.remove(address)
                Log.d(TAG, "Lost peer: $address")
                onPeerLost?.invoke(address)
            }
            
            Log.d(TAG, "Total peers: ${discoveredPeers.size}")
        }
        
        private fun handleConnectionInfo(info: WifiP2pInfo) {
            isConnected = info.groupFormed
            isGroupOwner = info.isGroupOwner
            groupOwnerAddress = info.groupOwnerAddress
            
            if (info.groupFormed) {
                val ownerAddr = info.groupOwnerAddress?.hostAddress ?: "unknown"
                Log.d(TAG, "Connected! Group owner: $ownerAddr, isOwner: $isGroupOwner")
                
                if (isGroupOwner) {
                    // Start server to accept connections
                    startServer()
                }
                
                onConnected?.invoke(ownerAddr, isGroupOwner)
            }
        }
    }
}
