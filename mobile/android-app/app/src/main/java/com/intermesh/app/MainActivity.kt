package com.intermesh.app

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.util.Log
import android.widget.Toast
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import com.google.android.material.button.MaterialButton
import com.google.android.material.switchmaterial.SwitchMaterial
import android.widget.TextView
import intermesh.Intermesh
import intermesh.MobileApp
import java.util.UUID

class MainActivity : AppCompatActivity() {
    
    private lateinit var mobileApp: MobileApp
    private val deviceId = "android-${UUID.randomUUID()}"
    
    private lateinit var statusText: TextView
    private lateinit var deviceIdText: TextView
    private lateinit var connectButton: MaterialButton
    private lateinit var peersCountText: TextView
    private lateinit var proxiesCountText: TextView
    private lateinit var sharingSwitch: SwitchMaterial
    private lateinit var requestInternetButton: MaterialButton
    private lateinit var messageText: TextView
    
    // WiFi Direct manager for peer-to-peer discovery
    private lateinit var wifiDirectManager: WifiDirectManager
    private var isWifiDirectEnabled = false
    
    // BLE manager for cross-platform connectivity (iOS <-> Android)
    private lateinit var bleManager: BLEManager
    private var isBleEnabled = false
    
    private var isConnected = false
    private val mainHandler = Handler(Looper.getMainLooper())
    
    // Track discovered P2P peers separately
    private val p2pPeers = mutableMapOf<String, WifiDirectManager.PeerInfo>()
    private val blePeers = mutableMapOf<String, BLEManager.BLEPeer>()
    
    companion object {
        private const val TAG = "InterMesh"
        private const val PERMISSION_REQUEST_CODE = 1001
    }
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        
        // Initialize views
        initViews()
        
        // Check permissions
        if (checkPermissions()) {
            initializeMesh()
            initializeWifiDirect()
            initializeBLE()
        } else {
            requestPermissions()
        }
        
        // Set up click listeners
        setupListeners()
    }
    
    private fun initViews() {
        statusText = findViewById(R.id.statusText)
        deviceIdText = findViewById(R.id.deviceIdText)
        connectButton = findViewById(R.id.connectButton)
        peersCountText = findViewById(R.id.peersCountText)
        proxiesCountText = findViewById(R.id.proxiesCountText)
        sharingSwitch = findViewById(R.id.sharingSwitch)
        requestInternetButton = findViewById(R.id.requestInternetButton)
        messageText = findViewById(R.id.messageText)
        
        deviceIdText.text = "Device ID: ${deviceId.takeLast(12)}"
    }
    
    private fun setupListeners() {
        connectButton.setOnClickListener {
            if (isConnected) {
                disconnectFromMesh()
            } else {
                connectToMesh()
            }
        }
        
        // Long press to show peer selection dialog
        connectButton.setOnLongClickListener {
            if (isConnected && isWifiDirectEnabled && p2pPeers.isNotEmpty()) {
                showPeerSelectionDialog()
                true
            } else {
                false
            }
        }
        
        sharingSwitch.setOnCheckedChangeListener { _, isChecked ->
            if (isChecked) {
                enableInternetSharing()
            } else {
                // Disable sharing logic would go here
                showMessage("Internet sharing disabled")
            }
        }
        
        requestInternetButton.setOnClickListener {
            requestInternetAccess()
        }
        
        // Make peers count clickable to show peer list
        peersCountText.setOnClickListener {
            if (p2pPeers.isNotEmpty()) {
                showPeerSelectionDialog()
            } else {
                showMessage("No peers discovered yet. Make sure WiFi is on and another device is nearby.")
            }
        }
    }
    
    private fun checkPermissions(): Boolean {
        val permissions = mutableListOf(
            Manifest.permission.ACCESS_FINE_LOCATION,
            Manifest.permission.ACCESS_COARSE_LOCATION,
            Manifest.permission.ACCESS_WIFI_STATE,
            Manifest.permission.CHANGE_WIFI_STATE,
            Manifest.permission.INTERNET
        )
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            permissions.add(Manifest.permission.NEARBY_WIFI_DEVICES)
            permissions.add(Manifest.permission.BLUETOOTH_SCAN)
            permissions.add(Manifest.permission.BLUETOOTH_ADVERTISE)
            permissions.add(Manifest.permission.BLUETOOTH_CONNECT)
        } else {
            permissions.add(Manifest.permission.BLUETOOTH)
            permissions.add(Manifest.permission.BLUETOOTH_ADMIN)
        }
        
        return permissions.all {
            ContextCompat.checkSelfPermission(this, it) == PackageManager.PERMISSION_GRANTED
        }
    }
    
    private fun requestPermissions() {
        val permissions = mutableListOf(
            Manifest.permission.ACCESS_FINE_LOCATION,
            Manifest.permission.ACCESS_COARSE_LOCATION,
            Manifest.permission.ACCESS_WIFI_STATE,
            Manifest.permission.CHANGE_WIFI_STATE,
            Manifest.permission.INTERNET
        )
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            permissions.add(Manifest.permission.NEARBY_WIFI_DEVICES)
            permissions.add(Manifest.permission.BLUETOOTH_SCAN)
            permissions.add(Manifest.permission.BLUETOOTH_ADVERTISE)
            permissions.add(Manifest.permission.BLUETOOTH_CONNECT)
        } else {
            permissions.add(Manifest.permission.BLUETOOTH)
            permissions.add(Manifest.permission.BLUETOOTH_ADMIN)
        }
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            permissions.add(Manifest.permission.NEARBY_WIFI_DEVICES)
        }
        
        ActivityCompat.requestPermissions(
            this,
            permissions.toTypedArray(),
            PERMISSION_REQUEST_CODE
        )
    }
    
    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        
        if (requestCode == PERMISSION_REQUEST_CODE) {
            if (grantResults.all { it == PackageManager.PERMISSION_GRANTED }) {
                initializeMesh()
                initializeWifiDirect()
                initializeBLE()
            } else {
                Toast.makeText(
                    this,
                    "Permissions required for mesh networking",
                    Toast.LENGTH_LONG
                ).show()
            }
        }
    }
    
    private fun initializeMesh() {
        try {
            // Create mobile app instance
            mobileApp = Intermesh.newMobileApp(deviceId, "Android Device", "192.168.1.100", "00:00:00:00:00:00")
            
            Log.d(TAG, "Mesh initialized successfully")
            showMessage("Mesh system ready")
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize mesh: ${e.message}", e)
            showMessage("Failed to initialize: ${e.message}")
        }
    }
    
    private fun initializeWifiDirect() {
        try {
            wifiDirectManager = WifiDirectManager(this)
            
            if (!wifiDirectManager.initialize()) {
                Log.w(TAG, "WiFi Direct not supported on this device")
                showMessage("WiFi Direct not available")
                return
            }
            
            isWifiDirectEnabled = true
            
            // Set up WiFi Direct callbacks
            wifiDirectManager.onPeerDiscovered = { peer ->
                mainHandler.post {
                    p2pPeers[peer.deviceAddress] = peer
                    Log.d(TAG, "P2P Peer found: ${peer.deviceName}")
                    updatePeerCount()
                    showMessage("Found peer: ${peer.deviceName}")
                }
            }
            
            wifiDirectManager.onPeerLost = { address ->
                mainHandler.post {
                    val peer = p2pPeers.remove(address)
                    Log.d(TAG, "P2P Peer lost: ${peer?.deviceName ?: address}")
                    updatePeerCount()
                }
            }
            
            wifiDirectManager.onConnected = { address, isGroupOwner ->
                mainHandler.post {
                    val role = if (isGroupOwner) "Group Owner" else "Client"
                    Log.d(TAG, "P2P Connected as $role to $address")
                    showMessage("P2P Connected! ($role)")
                    updateConnectionStatus(true)
                }
            }
            
            wifiDirectManager.onDisconnected = {
                mainHandler.post {
                    Log.d(TAG, "P2P Disconnected")
                    showMessage("P2P Disconnected")
                }
            }
            
            wifiDirectManager.onMessageReceived = { from, data ->
                mainHandler.post {
                    Log.d(TAG, "Received ${data.size} bytes from $from")
                    showMessage("Received data from peer")
                }
            }
            
            wifiDirectManager.onError = { error ->
                mainHandler.post {
                    Log.e(TAG, "WiFi Direct error: $error")
                    showMessage("P2P Error: $error")
                }
            }
            
            Log.d(TAG, "WiFi Direct initialized")
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize WiFi Direct: ${e.message}", e)
            isWifiDirectEnabled = false
        }
    }
    
    private fun initializeBLE() {
        try {
            bleManager = BLEManager(this)
            
            if (!bleManager.initialize(deviceId, "Android-${deviceId.takeLast(6)}")) {
                Log.w(TAG, "BLE not available on this device")
                return
            }
            
            isBleEnabled = true
            
            // Set up BLE callbacks
            bleManager.onPeerDiscovered = { peer ->
                mainHandler.post {
                    blePeers[peer.address] = peer
                    Log.d(TAG, "BLE Peer found: ${peer.name} (${peer.platform})")
                    updatePeerCount()
                    showMessage("BLE: Found ${peer.name} (${peer.platform})")
                }
            }
            
            bleManager.onPeerConnected = { peer ->
                mainHandler.post {
                    blePeers[peer.address] = peer
                    Log.d(TAG, "BLE Connected to: ${peer.name} (${peer.platform})")
                    showMessage("BLE Connected: ${peer.name} (${peer.platform})")
                    updatePeerCount()
                }
            }
            
            bleManager.onPeerDisconnected = { address ->
                mainHandler.post {
                    val peer = blePeers.remove(address)
                    Log.d(TAG, "BLE Peer disconnected: ${peer?.name ?: address}")
                    updatePeerCount()
                }
            }
            
            bleManager.onMessageReceived = { from, data ->
                mainHandler.post {
                    val message = String(data, Charsets.UTF_8)
                    Log.d(TAG, "BLE Message from $from: $message")
                    showMessage("BLE: $message")
                }
            }
            
            bleManager.onError = { error ->
                mainHandler.post {
                    Log.e(TAG, "BLE error: $error")
                    showMessage("BLE Error: $error")
                }
            }
            
            Log.d(TAG, "BLE Manager initialized")
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize BLE: ${e.message}", e)
            isBleEnabled = false
        }
    }
    
    private fun updatePeerCount() {
        val meshPeers = if (::mobileApp.isInitialized && isConnected) {
            try {
                mobileApp.getNetworkStats().peerCount.toInt()
            } catch (e: Exception) { 0 }
        } else 0
        
        // Combine WiFi Direct + BLE peers
        val totalPeers = meshPeers + p2pPeers.size + blePeers.size
        peersCountText.text = totalPeers.toString()
        
        // Update proxies (P2P peers with internet)
        val proxies = if (::mobileApp.isInitialized && isConnected) {
            try {
                mobileApp.getNetworkStats().availableProxies.toInt()
            } catch (e: Exception) { 0 }
        } else 0
        proxiesCountText.text = proxies.toString()
    }
    
    private fun connectToMesh() {
        try {
            // Show connecting state first
            statusText.text = "Connecting..."
            statusText.setTextColor(getColor(R.color.primary))
            connectButton.isEnabled = false
            showMessage("Starting mesh network...")
            
            // Start traditional mesh networking
            mobileApp.start()
            mobileApp.connectToNetwork()
            
            // Also start WiFi Direct peer discovery
            if (isWifiDirectEnabled) {
                wifiDirectManager.startDiscovery()
                Log.d(TAG, "WiFi Direct discovery started")
            }
            
            // Start BLE for cross-platform connectivity (iOS <-> Android)
            if (isBleEnabled) {
                bleManager.start()
                Log.d(TAG, "BLE started for cross-platform discovery")
            }
            
            // Now connected
            isConnected = true
            connectButton.isEnabled = true
            updateConnectionStatus(true)
            
            val modes = mutableListOf("Mesh")
            if (isWifiDirectEnabled) modes.add("WiFi Direct")
            if (isBleEnabled) modes.add("BLE")
            showMessage("Connected! (${modes.joinToString(" + ")}) Searching for peers...")
            
            // Start periodic stats update
            startStatsUpdate()
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to connect: ${e.message}", e)
            connectButton.isEnabled = true
            updateConnectionStatus(false)
            showMessage("Connection failed: ${e.message}")
        }
    }
    
    private fun disconnectFromMesh() {
        try {
            // Stop BLE
            if (isBleEnabled && ::bleManager.isInitialized) {
                bleManager.stop()
            }
            
            // Stop WiFi Direct
            if (isWifiDirectEnabled) {
                wifiDirectManager.stopDiscovery()
                wifiDirectManager.disconnect()
            }
            
            // Stop mesh
            mobileApp.stop()
            
            isConnected = false
            p2pPeers.clear()
            blePeers.clear()
            updateConnectionStatus(false)
            showMessage("Disconnected from mesh")
            
            // Stop stats update
            stopStatsUpdate()
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to disconnect: ${e.message}", e)
        }
    }
    
    private var statsUpdateRunnable: Runnable? = null
    
    private fun startStatsUpdate() {
        statsUpdateRunnable = object : Runnable {
            override fun run() {
                if (isConnected) {
                    updatePeerCount()
                    mainHandler.postDelayed(this, 2000) // Update every 2 seconds
                }
            }
        }
        mainHandler.post(statsUpdateRunnable!!)
    }
    
    private fun stopStatsUpdate() {
        statsUpdateRunnable?.let { mainHandler.removeCallbacks(it) }
        statsUpdateRunnable = null
    }
    
    private fun showPeerSelectionDialog() {
        // Combine WiFi Direct and BLE peers
        val allPeers = mutableListOf<Pair<String, () -> Unit>>()
        
        // Add WiFi Direct peers
        p2pPeers.values.forEach { peer ->
            allPeers.add("📶 ${peer.deviceName} (WiFi Direct)" to {
                showMessage("Connecting to ${peer.deviceName} via WiFi Direct...")
                wifiDirectManager.connectToPeer(peer.deviceAddress)
            })
        }
        
        // Add BLE peers
        blePeers.values.forEach { peer ->
            val platformIcon = if (peer.platform == "ios") "🍎" else "🤖"
            allPeers.add("$platformIcon ${peer.name} (BLE - ${peer.platform})" to {
                showMessage("Connecting to ${peer.name} via BLE...")
                bleManager.connectToPeer(peer)
            })
        }
        
        if (allPeers.isEmpty()) {
            showMessage("No peers found yet. Make sure other devices are nearby with InterMesh running.")
            return
        }
        
        val peerNames = allPeers.map { it.first }.toTypedArray()
        
        AlertDialog.Builder(this)
            .setTitle("Connect to Peer")
            .setItems(peerNames) { _, which ->
                allPeers[which].second()
            }
            .setNegativeButton("Cancel", null)
            .show()
    }
    
    private fun sendTestMessage() {
        // Send to both WiFi Direct and BLE peers
        val testMessage = "Hello from Android ${deviceId.takeLast(6)}!"
        
        if (isBleEnabled && bleManager.getConnectedPeerCount() > 0) {
            bleManager.sendTextMessage(testMessage)
            showMessage("Test message sent via BLE")
        }
        
        if (isWifiDirectEnabled && wifiDirectManager.isConnectedToPeer()) {
            wifiDirectManager.sendData(testMessage.toByteArray())
            showMessage("Test message sent via WiFi Direct")
        }
    }
    
    private fun enableInternetSharing() {
        try {
            mobileApp.enableInternetSharing()
            showMessage("Internet sharing enabled")
        } catch (e: Exception) {
            showMessage("Failed to enable sharing: ${e.message}")
            sharingSwitch.isChecked = false
            Log.e(TAG, "Failed to enable sharing: ${e.message}", e)
        }
    }
    
    private fun requestInternetAccess() {
        try {
            val message = mobileApp.requestInternetAccess()
            showMessage(message)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to request internet: ${e.message}", e)
            showMessage("Error: ${e.message}")
        }
    }
    
    private fun updateConnectionStatus(connected: Boolean) {
        isConnected = connected
        
        if (connected) {
            statusText.text = "Connected"
            statusText.setTextColor(getColor(R.color.primary))
            connectButton.text = "Disconnect"
        } else {
            statusText.text = "Disconnected"
            statusText.setTextColor(getColor(R.color.textSecondary))
            connectButton.text = "Connect to Mesh"
            peersCountText.text = "0"
            proxiesCountText.text = "0"
        }
        
        updateStats()
    }
    
    private fun updateStats() {
        if (!isConnected) return
        
        try {
            val stats = mobileApp.getNetworkStats()
            
            // Combine mesh peers with P2P peers (WiFi Direct + BLE)
            val totalPeers = stats.peerCount + p2pPeers.size + blePeers.size
            peersCountText.text = totalPeers.toString()
            proxiesCountText.text = stats.availableProxies.toString()
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to get stats: ${e.message}", e)
        }
    }
    
    private fun showMessage(message: String) {
        messageText.text = message
        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
    }
    
    override fun onDestroy() {
        super.onDestroy()
        
        // Stop stats updates
        stopStatsUpdate()
        
        // Clean up BLE
        if (::bleManager.isInitialized) {
            bleManager.stop()
        }
        
        // Clean up WiFi Direct
        if (::wifiDirectManager.isInitialized) {
            wifiDirectManager.cleanup()
        }
        
        // Clean up mesh
        if (::mobileApp.isInitialized && isConnected) {
            mobileApp.stop()
        }
    }
}