package com.intermesh.app

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.util.Log
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import com.google.android.material.button.MaterialButton
import com.google.android.material.switchmaterial.SwitchMaterial
import android.widget.TextView
import intermesh.Intermesh
import intermesh.MobileApp
import intermesh.MobileListenerProxy
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
    
    private var isConnected = false
    
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
        
        deviceIdText.text = "Device ID: $deviceId"
    }
    
    private fun setupListeners() {
        connectButton.setOnClickListener {
            if (isConnected) {
                disconnectFromMesh()
            } else {
                connectToMesh()
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
            mobileApp = Intermesh.newMobileApp(deviceId, "Android Device")
            
            // Register listener for connection events
            val listener = object : MobileListenerProxy() {
                override fun onConnectionStatusChanged(connected: Boolean) {
                    runOnUiThread {
                        updateConnectionStatus(connected)
                    }
                }
                
                override fun onPeerDiscovered(peerId: String) {
                    runOnUiThread {
                        Log.d(TAG, "Peer discovered: $peerId")
                        updateStats()
                    }
                }
                
                override fun onPeerLost(peerId: String) {
                    runOnUiThread {
                        Log.d(TAG, "Peer lost: $peerId")
                        updateStats()
                    }
                }
                
                override fun onInternetAvailable() {
                    runOnUiThread {
                        showMessage("Internet connection available!")
                    }
                }
            }
            
            mobileApp.registerListener(listener)
            
            Log.d(TAG, "Mesh initialized successfully")
            showMessage("Mesh system ready")
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize mesh: ${e.message}", e)
            showMessage("Failed to initialize: ${e.message}")
        }
    }
    
    private fun connectToMesh() {
        try {
            mobileApp.start()
            mobileApp.connectToNetwork()
            
            isConnected = true
            updateConnectionStatus(true)
            showMessage("Connecting to mesh network...")
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to connect: ${e.message}", e)
            showMessage("Connection failed: ${e.message}")
        }
    }
    
    private fun disconnectFromMesh() {
        try {
            mobileApp.stop()
            
            isConnected = false
            updateConnectionStatus(false)
            showMessage("Disconnected from mesh")
            
        } catch (e: Exception) {
            Log.e(TAG, "Failed to disconnect: ${e.message}", e)
        }
    }
    
    private fun enableInternetSharing() {
        try {
            val success = mobileApp.enableInternetSharing()
            if (success) {
                showMessage("Internet sharing enabled")
            } else {
                showMessage("Failed to enable sharing")
                sharingSwitch.isChecked = false
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to enable sharing: ${e.message}", e)
            showMessage("Error: ${e.message}")
            sharingSwitch.isChecked = false
        }
    }
    
    private fun requestInternetAccess() {
        try {
            val success = mobileApp.requestInternetAccess()
            if (success) {
                showMessage("Internet access request sent")
            } else {
                showMessage("No proxies available")
            }
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
            
            peersCountText.text = stats.connectedPeers.toString()
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
        if (::mobileApp.isInitialized && isConnected) {
            mobileApp.stop()
        }
    }
}
