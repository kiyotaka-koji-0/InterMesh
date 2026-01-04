package com.intermesh.app

import android.Manifest
import android.annotation.SuppressLint
import android.bluetooth.*
import android.bluetooth.le.*
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import android.os.Handler
import android.os.Looper
import android.os.ParcelUuid
import android.util.Log
import androidx.core.content.ContextCompat
import java.nio.charset.StandardCharsets
import java.util.*
import java.util.concurrent.ConcurrentHashMap

/**
 * BLEManager handles Bluetooth Low Energy communication for cross-platform
 * mesh networking between Android and iOS devices.
 * 
 * Uses common UUIDs that match the iOS implementation for interoperability.
 */
class BLEManager(private val context: Context) {
    
    companion object {
        private const val TAG = "BLEManager"
        
        // Common UUIDs shared with iOS - MUST MATCH iOS BLEManager
        val SERVICE_UUID: UUID = UUID.fromString("A1B2C3D4-E5F6-7890-ABCD-EF1234567890")
        val DEVICE_INFO_CHAR_UUID: UUID = UUID.fromString("A1B2C3D4-E5F6-7890-ABCD-EF1234567891")
        val MESSAGE_CHAR_UUID: UUID = UUID.fromString("A1B2C3D4-E5F6-7890-ABCD-EF1234567892")
        val MESH_DATA_CHAR_UUID: UUID = UUID.fromString("A1B2C3D4-E5F6-7890-ABCD-EF1234567893")
        
        // Descriptor UUID for notifications
        val CCCD_UUID: UUID = UUID.fromString("00002902-0000-1000-8000-00805f9b34fb")
        
        private const val SCAN_PERIOD_MS = 10000L // 10 seconds scan period
        private const val MAX_MTU = 512
    }
    
    // Bluetooth components
    private var bluetoothManager: BluetoothManager? = null
    private var bluetoothAdapter: BluetoothAdapter? = null
    private var bleScanner: BluetoothLeScanner? = null
    private var bleAdvertiser: BluetoothLeAdvertiser? = null
    private var gattServer: BluetoothGattServer? = null
    
    // State
    private var isScanning = false
    private var isAdvertising = false
    private var isInitialized = false
    
    // Connected devices
    private val connectedDevices = ConcurrentHashMap<String, BluetoothDevice>()
    private val connectedGatts = ConcurrentHashMap<String, BluetoothGatt>()
    private val discoveredDevices = ConcurrentHashMap<String, BLEPeer>()
    
    // Device info
    private var localDeviceId: String = ""
    private var localDeviceName: String = ""
    
    // Handler for timeouts
    private val handler = Handler(Looper.getMainLooper())
    
    // Callbacks
    var onPeerDiscovered: ((BLEPeer) -> Unit)? = null
    var onPeerConnected: ((BLEPeer) -> Unit)? = null
    var onPeerDisconnected: ((String) -> Unit)? = null
    var onMessageReceived: ((String, ByteArray) -> Unit)? = null
    var onError: ((String) -> Unit)? = null
    var onStateChanged: ((Boolean) -> Unit)? = null
    
    data class BLEPeer(
        val id: String,
        val name: String,
        val address: String,
        val platform: String, // "android" or "ios"
        val rssi: Int = 0
    )
    
    /**
     * Initialize the BLE manager
     */
    fun initialize(deviceId: String, deviceName: String): Boolean {
        localDeviceId = deviceId
        localDeviceName = deviceName
        
        if (!context.packageManager.hasSystemFeature(PackageManager.FEATURE_BLUETOOTH_LE)) {
            Log.e(TAG, "BLE not supported on this device")
            onError?.invoke("BLE not supported on this device")
            return false
        }
        
        bluetoothManager = context.getSystemService(Context.BLUETOOTH_SERVICE) as? BluetoothManager
        bluetoothAdapter = bluetoothManager?.adapter
        
        if (bluetoothAdapter == null) {
            Log.e(TAG, "Bluetooth adapter not available")
            onError?.invoke("Bluetooth not available")
            return false
        }
        
        if (!bluetoothAdapter!!.isEnabled) {
            Log.e(TAG, "Bluetooth is not enabled")
            onError?.invoke("Please enable Bluetooth")
            return false
        }
        
        bleScanner = bluetoothAdapter?.bluetoothLeScanner
        bleAdvertiser = bluetoothAdapter?.bluetoothLeAdvertiser
        
        isInitialized = true
        Log.i(TAG, "BLE Manager initialized successfully")
        return true
    }
    
    /**
     * Start BLE operations (advertising + scanning)
     */
    @SuppressLint("MissingPermission")
    fun start() {
        if (!isInitialized) {
            onError?.invoke("BLE Manager not initialized")
            return
        }
        
        if (!hasRequiredPermissions()) {
            onError?.invoke("Missing Bluetooth permissions")
            return
        }
        
        startGattServer()
        startAdvertising()
        startScanning()
        
        onStateChanged?.invoke(true)
    }
    
    /**
     * Stop all BLE operations
     */
    fun stop() {
        stopScanning()
        stopAdvertising()
        stopGattServer()
        disconnectAll()
        
        onStateChanged?.invoke(false)
    }
    
    /**
     * Check if we have required Bluetooth permissions
     */
    private fun hasRequiredPermissions(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            ContextCompat.checkSelfPermission(context, Manifest.permission.BLUETOOTH_SCAN) == PackageManager.PERMISSION_GRANTED &&
            ContextCompat.checkSelfPermission(context, Manifest.permission.BLUETOOTH_ADVERTISE) == PackageManager.PERMISSION_GRANTED &&
            ContextCompat.checkSelfPermission(context, Manifest.permission.BLUETOOTH_CONNECT) == PackageManager.PERMISSION_GRANTED
        } else {
            ContextCompat.checkSelfPermission(context, Manifest.permission.BLUETOOTH) == PackageManager.PERMISSION_GRANTED &&
            ContextCompat.checkSelfPermission(context, Manifest.permission.BLUETOOTH_ADMIN) == PackageManager.PERMISSION_GRANTED &&
            ContextCompat.checkSelfPermission(context, Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED
        }
    }
    
    // ==================== GATT Server ====================
    
    @SuppressLint("MissingPermission")
    private fun startGattServer() {
        try {
            gattServer = bluetoothManager?.openGattServer(context, gattServerCallback)
            
            val service = BluetoothGattService(SERVICE_UUID, BluetoothGattService.SERVICE_TYPE_PRIMARY)
            
            // Device info characteristic (read)
            val deviceInfoChar = BluetoothGattCharacteristic(
                DEVICE_INFO_CHAR_UUID,
                BluetoothGattCharacteristic.PROPERTY_READ,
                BluetoothGattCharacteristic.PERMISSION_READ
            )
            
            // Message characteristic (write + notify)
            val messageChar = BluetoothGattCharacteristic(
                MESSAGE_CHAR_UUID,
                BluetoothGattCharacteristic.PROPERTY_WRITE or BluetoothGattCharacteristic.PROPERTY_NOTIFY,
                BluetoothGattCharacteristic.PERMISSION_WRITE
            )
            messageChar.addDescriptor(BluetoothGattDescriptor(
                CCCD_UUID,
                BluetoothGattDescriptor.PERMISSION_READ or BluetoothGattDescriptor.PERMISSION_WRITE
            ))
            
            // Mesh data characteristic (write + notify)
            val meshDataChar = BluetoothGattCharacteristic(
                MESH_DATA_CHAR_UUID,
                BluetoothGattCharacteristic.PROPERTY_WRITE or BluetoothGattCharacteristic.PROPERTY_NOTIFY,
                BluetoothGattCharacteristic.PERMISSION_WRITE
            )
            meshDataChar.addDescriptor(BluetoothGattDescriptor(
                CCCD_UUID,
                BluetoothGattDescriptor.PERMISSION_READ or BluetoothGattDescriptor.PERMISSION_WRITE
            ))
            
            service.addCharacteristic(deviceInfoChar)
            service.addCharacteristic(messageChar)
            service.addCharacteristic(meshDataChar)
            
            gattServer?.addService(service)
            Log.i(TAG, "GATT Server started")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start GATT server: ${e.message}")
            onError?.invoke("Failed to start BLE server: ${e.message}")
        }
    }
    
    @SuppressLint("MissingPermission")
    private fun stopGattServer() {
        gattServer?.close()
        gattServer = null
        Log.i(TAG, "GATT Server stopped")
    }
    
    private val gattServerCallback = object : BluetoothGattServerCallback() {
        @SuppressLint("MissingPermission")
        override fun onConnectionStateChange(device: BluetoothDevice, status: Int, newState: Int) {
            when (newState) {
                BluetoothProfile.STATE_CONNECTED -> {
                    Log.i(TAG, "Device connected: ${device.address}")
                    connectedDevices[device.address] = device
                }
                BluetoothProfile.STATE_DISCONNECTED -> {
                    Log.i(TAG, "Device disconnected: ${device.address}")
                    connectedDevices.remove(device.address)
                    handler.post { onPeerDisconnected?.invoke(device.address) }
                }
            }
        }
        
        @SuppressLint("MissingPermission")
        override fun onCharacteristicReadRequest(
            device: BluetoothDevice,
            requestId: Int,
            offset: Int,
            characteristic: BluetoothGattCharacteristic
        ) {
            when (characteristic.uuid) {
                DEVICE_INFO_CHAR_UUID -> {
                    val deviceInfo = "$localDeviceId|$localDeviceName|android"
                    val data = deviceInfo.toByteArray(StandardCharsets.UTF_8)
                    gattServer?.sendResponse(device, requestId, BluetoothGatt.GATT_SUCCESS, offset, 
                        data.copyOfRange(offset, minOf(offset + 20, data.size)))
                }
                else -> {
                    gattServer?.sendResponse(device, requestId, BluetoothGatt.GATT_FAILURE, 0, null)
                }
            }
        }
        
        @SuppressLint("MissingPermission")
        override fun onCharacteristicWriteRequest(
            device: BluetoothDevice,
            requestId: Int,
            characteristic: BluetoothGattCharacteristic,
            preparedWrite: Boolean,
            responseNeeded: Boolean,
            offset: Int,
            value: ByteArray
        ) {
            when (characteristic.uuid) {
                MESSAGE_CHAR_UUID -> {
                    Log.d(TAG, "Received message from ${device.address}")
                    handler.post { onMessageReceived?.invoke(device.address, value) }
                    if (responseNeeded) {
                        gattServer?.sendResponse(device, requestId, BluetoothGatt.GATT_SUCCESS, 0, null)
                    }
                }
                MESH_DATA_CHAR_UUID -> {
                    Log.d(TAG, "Received mesh data from ${device.address}")
                    handler.post { onMessageReceived?.invoke(device.address, value) }
                    if (responseNeeded) {
                        gattServer?.sendResponse(device, requestId, BluetoothGatt.GATT_SUCCESS, 0, null)
                    }
                }
                else -> {
                    if (responseNeeded) {
                        gattServer?.sendResponse(device, requestId, BluetoothGatt.GATT_FAILURE, 0, null)
                    }
                }
            }
        }
        
        @SuppressLint("MissingPermission")
        override fun onDescriptorWriteRequest(
            device: BluetoothDevice,
            requestId: Int,
            descriptor: BluetoothGattDescriptor,
            preparedWrite: Boolean,
            responseNeeded: Boolean,
            offset: Int,
            value: ByteArray
        ) {
            if (responseNeeded) {
                gattServer?.sendResponse(device, requestId, BluetoothGatt.GATT_SUCCESS, 0, null)
            }
        }
    }
    
    // ==================== Advertising ====================
    
    @SuppressLint("MissingPermission")
    private fun startAdvertising() {
        if (bleAdvertiser == null) {
            Log.e(TAG, "BLE Advertiser not available")
            return
        }
        
        val settings = AdvertiseSettings.Builder()
            .setAdvertiseMode(AdvertiseSettings.ADVERTISE_MODE_LOW_LATENCY)
            .setConnectable(true)
            .setTimeout(0)
            .setTxPowerLevel(AdvertiseSettings.ADVERTISE_TX_POWER_HIGH)
            .build()
        
        val data = AdvertiseData.Builder()
            .setIncludeDeviceName(false)
            .addServiceUuid(ParcelUuid(SERVICE_UUID))
            .build()
        
        val scanResponse = AdvertiseData.Builder()
            .setIncludeDeviceName(true)
            .build()
        
        try {
            bleAdvertiser?.startAdvertising(settings, data, scanResponse, advertiseCallback)
            isAdvertising = true
            Log.i(TAG, "BLE Advertising started")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start advertising: ${e.message}")
            onError?.invoke("Failed to start BLE advertising")
        }
    }
    
    @SuppressLint("MissingPermission")
    private fun stopAdvertising() {
        if (isAdvertising) {
            bleAdvertiser?.stopAdvertising(advertiseCallback)
            isAdvertising = false
            Log.i(TAG, "BLE Advertising stopped")
        }
    }
    
    private val advertiseCallback = object : AdvertiseCallback() {
        override fun onStartSuccess(settingsInEffect: AdvertiseSettings) {
            Log.i(TAG, "Advertising started successfully")
        }
        
        override fun onStartFailure(errorCode: Int) {
            Log.e(TAG, "Advertising failed with error code: $errorCode")
            isAdvertising = false
            val errorMsg = when (errorCode) {
                ADVERTISE_FAILED_DATA_TOO_LARGE -> "Data too large"
                ADVERTISE_FAILED_TOO_MANY_ADVERTISERS -> "Too many advertisers"
                ADVERTISE_FAILED_ALREADY_STARTED -> "Already started"
                ADVERTISE_FAILED_INTERNAL_ERROR -> "Internal error"
                ADVERTISE_FAILED_FEATURE_UNSUPPORTED -> "Feature unsupported"
                else -> "Unknown error"
            }
            handler.post { onError?.invoke("BLE advertising failed: $errorMsg") }
        }
    }
    
    // ==================== Scanning ====================
    
    @SuppressLint("MissingPermission")
    private fun startScanning() {
        if (bleScanner == null) {
            Log.e(TAG, "BLE Scanner not available")
            return
        }
        
        val scanFilter = ScanFilter.Builder()
            .setServiceUuid(ParcelUuid(SERVICE_UUID))
            .build()
        
        val scanSettings = ScanSettings.Builder()
            .setScanMode(ScanSettings.SCAN_MODE_LOW_LATENCY)
            .setReportDelay(0)
            .build()
        
        try {
            bleScanner?.startScan(listOf(scanFilter), scanSettings, scanCallback)
            isScanning = true
            Log.i(TAG, "BLE Scanning started")
            
            // Auto-restart scanning periodically
            handler.postDelayed({
                if (isScanning) {
                    stopScanning()
                    startScanning()
                }
            }, SCAN_PERIOD_MS)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start scanning: ${e.message}")
            onError?.invoke("Failed to start BLE scanning")
        }
    }
    
    @SuppressLint("MissingPermission")
    private fun stopScanning() {
        if (isScanning) {
            bleScanner?.stopScan(scanCallback)
            isScanning = false
            Log.i(TAG, "BLE Scanning stopped")
        }
    }
    
    private val scanCallback = object : ScanCallback() {
        @SuppressLint("MissingPermission")
        override fun onScanResult(callbackType: Int, result: ScanResult) {
            val device = result.device
            val address = device.address
            
            if (!discoveredDevices.containsKey(address)) {
                val deviceName = device.name ?: "Unknown Device"
                val peer = BLEPeer(
                    id = address,
                    name = deviceName,
                    address = address,
                    platform = "unknown", // Will be updated after connection
                    rssi = result.rssi
                )
                discoveredDevices[address] = peer
                Log.i(TAG, "Discovered BLE peer: $deviceName ($address)")
                handler.post { onPeerDiscovered?.invoke(peer) }
                
                // Auto-connect to discovered peers
                connectToDevice(device)
            }
        }
        
        override fun onScanFailed(errorCode: Int) {
            Log.e(TAG, "Scan failed with error code: $errorCode")
            isScanning = false
            val errorMsg = when (errorCode) {
                SCAN_FAILED_ALREADY_STARTED -> "Already started"
                SCAN_FAILED_APPLICATION_REGISTRATION_FAILED -> "App registration failed"
                SCAN_FAILED_INTERNAL_ERROR -> "Internal error"
                SCAN_FAILED_FEATURE_UNSUPPORTED -> "Feature unsupported"
                else -> "Unknown error"
            }
            handler.post { onError?.invoke("BLE scan failed: $errorMsg") }
        }
    }
    
    // ==================== Connection Management ====================
    
    @SuppressLint("MissingPermission")
    fun connectToDevice(device: BluetoothDevice) {
        if (connectedGatts.containsKey(device.address)) {
            Log.d(TAG, "Already connected to ${device.address}")
            return
        }
        
        Log.i(TAG, "Connecting to ${device.address}")
        val gatt = device.connectGatt(context, false, gattCallback, BluetoothDevice.TRANSPORT_LE)
        connectedGatts[device.address] = gatt
    }
    
    @SuppressLint("MissingPermission")
    fun connectToPeer(peer: BLEPeer) {
        val device = bluetoothAdapter?.getRemoteDevice(peer.address)
        if (device != null) {
            connectToDevice(device)
        }
    }
    
    @SuppressLint("MissingPermission")
    private fun disconnectAll() {
        connectedGatts.values.forEach { gatt ->
            gatt.disconnect()
            gatt.close()
        }
        connectedGatts.clear()
        connectedDevices.clear()
        discoveredDevices.clear()
    }
    
    private val gattCallback = object : BluetoothGattCallback() {
        @SuppressLint("MissingPermission")
        override fun onConnectionStateChange(gatt: BluetoothGatt, status: Int, newState: Int) {
            val address = gatt.device.address
            
            when (newState) {
                BluetoothProfile.STATE_CONNECTED -> {
                    Log.i(TAG, "Connected to GATT server: $address")
                    // Request larger MTU for better throughput
                    gatt.requestMtu(MAX_MTU)
                }
                BluetoothProfile.STATE_DISCONNECTED -> {
                    Log.i(TAG, "Disconnected from GATT server: $address")
                    connectedGatts.remove(address)
                    discoveredDevices.remove(address)
                    handler.post { onPeerDisconnected?.invoke(address) }
                    gatt.close()
                }
            }
        }
        
        @SuppressLint("MissingPermission")
        override fun onMtuChanged(gatt: BluetoothGatt, mtu: Int, status: Int) {
            Log.d(TAG, "MTU changed to $mtu for ${gatt.device.address}")
            // Discover services after MTU negotiation
            gatt.discoverServices()
        }
        
        @SuppressLint("MissingPermission")
        override fun onServicesDiscovered(gatt: BluetoothGatt, status: Int) {
            if (status == BluetoothGatt.GATT_SUCCESS) {
                Log.i(TAG, "Services discovered for ${gatt.device.address}")
                
                val service = gatt.getService(SERVICE_UUID)
                if (service != null) {
                    // Read device info to get peer details
                    val deviceInfoChar = service.getCharacteristic(DEVICE_INFO_CHAR_UUID)
                    if (deviceInfoChar != null) {
                        gatt.readCharacteristic(deviceInfoChar)
                    }
                    
                    // Enable notifications for messages
                    val messageChar = service.getCharacteristic(MESSAGE_CHAR_UUID)
                    if (messageChar != null) {
                        enableNotifications(gatt, messageChar)
                    }
                } else {
                    Log.w(TAG, "InterMesh service not found on ${gatt.device.address}")
                }
            } else {
                Log.e(TAG, "Service discovery failed for ${gatt.device.address}")
            }
        }
        
        @SuppressLint("MissingPermission")
        override fun onCharacteristicRead(
            gatt: BluetoothGatt,
            characteristic: BluetoothGattCharacteristic,
            status: Int
        ) {
            if (status == BluetoothGatt.GATT_SUCCESS && characteristic.uuid == DEVICE_INFO_CHAR_UUID) {
                val data = characteristic.value
                if (data != null) {
                    val info = String(data, StandardCharsets.UTF_8)
                    val parts = info.split("|")
                    if (parts.size >= 3) {
                        val peer = BLEPeer(
                            id = parts[0],
                            name = parts[1],
                            address = gatt.device.address,
                            platform = parts[2]
                        )
                        discoveredDevices[gatt.device.address] = peer
                        Log.i(TAG, "Peer info: ${peer.name} (${peer.platform})")
                        handler.post { onPeerConnected?.invoke(peer) }
                    }
                }
            }
        }
        
        override fun onCharacteristicChanged(
            gatt: BluetoothGatt,
            characteristic: BluetoothGattCharacteristic
        ) {
            val data = characteristic.value
            if (data != null) {
                Log.d(TAG, "Received data from ${gatt.device.address}")
                handler.post { onMessageReceived?.invoke(gatt.device.address, data) }
            }
        }
    }
    
    @SuppressLint("MissingPermission")
    private fun enableNotifications(gatt: BluetoothGatt, characteristic: BluetoothGattCharacteristic) {
        gatt.setCharacteristicNotification(characteristic, true)
        val descriptor = characteristic.getDescriptor(CCCD_UUID)
        if (descriptor != null) {
            descriptor.value = BluetoothGattDescriptor.ENABLE_NOTIFICATION_VALUE
            gatt.writeDescriptor(descriptor)
        }
    }
    
    // ==================== Send Messages ====================
    
    /**
     * Send a message to a specific peer
     */
    @SuppressLint("MissingPermission")
    fun sendMessage(peerAddress: String, message: ByteArray): Boolean {
        val gatt = connectedGatts[peerAddress] ?: return false
        
        val service = gatt.getService(SERVICE_UUID) ?: return false
        val characteristic = service.getCharacteristic(MESSAGE_CHAR_UUID) ?: return false
        
        characteristic.value = message
        characteristic.writeType = BluetoothGattCharacteristic.WRITE_TYPE_DEFAULT
        return gatt.writeCharacteristic(characteristic)
    }
    
    /**
     * Send a message to all connected peers
     */
    fun broadcastMessage(message: ByteArray) {
        connectedGatts.keys.forEach { address ->
            sendMessage(address, message)
        }
    }
    
    /**
     * Send a text message to all connected peers
     */
    fun sendTextMessage(text: String) {
        broadcastMessage(text.toByteArray(StandardCharsets.UTF_8))
    }
    
    // ==================== Getters ====================
    
    fun isRunning(): Boolean = isScanning || isAdvertising
    
    fun getDiscoveredPeers(): List<BLEPeer> = discoveredDevices.values.toList()
    
    fun getConnectedPeers(): List<BLEPeer> = discoveredDevices.values.filter { 
        connectedGatts.containsKey(it.address) 
    }
    
    fun getConnectedPeerCount(): Int = connectedGatts.size
}
