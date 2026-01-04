//
//  BLEManager.swift
//  InterMesh
//
//  Bluetooth Low Energy manager for cross-platform mesh networking.
//  Uses the same UUIDs as Android for interoperability.
//

import Foundation
import CoreBluetooth
import Combine
import UIKit

/// Represents a discovered BLE peer
struct BLEPeer: Identifiable, Equatable {
    let id: String
    let name: String
    let identifier: UUID
    var platform: String // "android", "ios", or "unknown"
    var rssi: Int
    var isConnected: Bool = false
    
    static func == (lhs: BLEPeer, rhs: BLEPeer) -> Bool {
        return lhs.id == rhs.id
    }
}

/// BLEManager handles Bluetooth Low Energy communication for cross-platform
/// mesh networking between iOS and Android devices.
class BLEManager: NSObject, ObservableObject {
    
    // MARK: - Common UUIDs (MUST MATCH Android)
    
    static let serviceUUID = CBUUID(string: "A1B2C3D4-E5F6-7890-ABCD-EF1234567890")
    static let deviceInfoCharUUID = CBUUID(string: "A1B2C3D4-E5F6-7890-ABCD-EF1234567891")
    static let messageCharUUID = CBUUID(string: "A1B2C3D4-E5F6-7890-ABCD-EF1234567892")
    static let meshDataCharUUID = CBUUID(string: "A1B2C3D4-E5F6-7890-ABCD-EF1234567893")
    
    // MARK: - Published Properties
    
    @Published var isScanning = false
    @Published var isAdvertising = false
    @Published var discoveredPeers: [BLEPeer] = []
    @Published var connectedPeers: [BLEPeer] = []
    @Published var lastReceivedMessage: String = ""
    @Published var connectionStatus: String = "Idle"
    @Published var bleState: CBManagerState = .unknown
    
    // MARK: - Private Properties
    
    private var centralManager: CBCentralManager!
    private var peripheralManager: CBPeripheralManager!
    
    private var discoveredPeripherals: [UUID: CBPeripheral] = [:]
    private var connectedPeripherals: [UUID: CBPeripheral] = [:]
    private var peripheralCharacteristics: [UUID: [CBUUID: CBCharacteristic]] = [:]
    
    private var localDeviceId: String = ""
    private var localDeviceName: String = ""
    
    // Service and characteristics for peripheral mode
    private var meshService: CBMutableService?
    private var deviceInfoCharacteristic: CBMutableCharacteristic?
    private var messageCharacteristic: CBMutableCharacteristic?
    private var meshDataCharacteristic: CBMutableCharacteristic?
    
    // Connected centrals (devices that connected to us)
    private var connectedCentrals: [CBCentral] = []
    
    // MARK: - Callbacks
    
    var onPeerDiscovered: ((BLEPeer) -> Void)?
    var onPeerConnected: ((BLEPeer) -> Void)?
    var onPeerDisconnected: ((String) -> Void)?
    var onMessageReceived: ((String, Data) -> Void)?
    var onError: ((String) -> Void)?
    
    // MARK: - Initialization
    
    override init() {
        super.init()
        
        // Initialize with restoration identifier for background operation
        centralManager = CBCentralManager(delegate: self, queue: nil, options: [
            CBCentralManagerOptionRestoreIdentifierKey: "com.intermesh.central"
        ])
        
        peripheralManager = CBPeripheralManager(delegate: self, queue: nil, options: [
            CBPeripheralManagerOptionRestoreIdentifierKey: "com.intermesh.peripheral"
        ])
        
        localDeviceId = UIDevice.current.identifierForVendor?.uuidString ?? UUID().uuidString
        localDeviceName = UIDevice.current.name
    }
    
    // MARK: - Public Methods
    
    /// Start BLE operations (scanning + advertising)
    func start() {
        guard centralManager.state == .poweredOn else {
            connectionStatus = "Bluetooth not ready"
            onError?("Bluetooth is not powered on")
            return
        }
        
        setupService()
        startAdvertising()
        startScanning()
        
        connectionStatus = "BLE Active"
    }
    
    /// Stop all BLE operations
    func stop() {
        stopScanning()
        stopAdvertising()
        disconnectAll()
        
        connectionStatus = "Idle"
    }
    
    /// Connect to a discovered peer
    func connect(to peer: BLEPeer) {
        guard let peripheral = discoveredPeripherals[peer.identifier] else {
            print("BLE: Peripheral not found for peer \(peer.name)")
            return
        }
        
        print("BLE: Connecting to \(peer.name)")
        centralManager.connect(peripheral, options: nil)
    }
    
    /// Send a message to all connected peers
    func sendMessage(_ message: String) {
        let data = message.data(using: .utf8) ?? Data()
        broadcastData(data)
    }
    
    /// Send data to all connected peers
    func broadcastData(_ data: Data) {
        // Send to peripherals we're connected to as central
        for (uuid, peripheral) in connectedPeripherals {
            if let chars = peripheralCharacteristics[uuid],
               let messageChar = chars[BLEManager.messageCharUUID] {
                peripheral.writeValue(data, for: messageChar, type: .withResponse)
            }
        }
        
        // Send to centrals connected to us (as peripheral)
        if let characteristic = messageCharacteristic {
            peripheralManager.updateValue(data, for: characteristic, onSubscribedCentrals: nil)
        }
    }
    
    /// Send data to a specific peer
    func sendData(_ data: Data, to peerIdentifier: UUID) -> Bool {
        if let peripheral = connectedPeripherals[peerIdentifier],
           let chars = peripheralCharacteristics[peerIdentifier],
           let messageChar = chars[BLEManager.messageCharUUID] {
            peripheral.writeValue(data, for: messageChar, type: .withResponse)
            return true
        }
        return false
    }
    
    // MARK: - Private Methods - Service Setup
    
    private func setupService() {
        // Device info characteristic (readable)
        deviceInfoCharacteristic = CBMutableCharacteristic(
            type: BLEManager.deviceInfoCharUUID,
            properties: [.read],
            value: nil,
            permissions: [.readable]
        )
        
        // Message characteristic (writable + notifiable)
        messageCharacteristic = CBMutableCharacteristic(
            type: BLEManager.messageCharUUID,
            properties: [.write, .notify],
            value: nil,
            permissions: [.writeable]
        )
        
        // Mesh data characteristic (writable + notifiable)
        meshDataCharacteristic = CBMutableCharacteristic(
            type: BLEManager.meshDataCharUUID,
            properties: [.write, .notify],
            value: nil,
            permissions: [.writeable]
        )
        
        meshService = CBMutableService(type: BLEManager.serviceUUID, primary: true)
        meshService?.characteristics = [
            deviceInfoCharacteristic!,
            messageCharacteristic!,
            meshDataCharacteristic!
        ]
    }
    
    // MARK: - Private Methods - Advertising
    
    private func startAdvertising() {
        guard peripheralManager.state == .poweredOn else {
            print("BLE: Peripheral manager not ready")
            return
        }
        
        // Add service if not already added
        if let service = meshService {
            peripheralManager.removeAllServices()
            peripheralManager.add(service)
        }
        
        let advertisementData: [String: Any] = [
            CBAdvertisementDataServiceUUIDsKey: [BLEManager.serviceUUID],
            CBAdvertisementDataLocalNameKey: localDeviceName
        ]
        
        peripheralManager.startAdvertising(advertisementData)
        isAdvertising = true
        print("BLE: Started advertising")
    }
    
    private func stopAdvertising() {
        peripheralManager.stopAdvertising()
        isAdvertising = false
        print("BLE: Stopped advertising")
    }
    
    // MARK: - Private Methods - Scanning
    
    private func startScanning() {
        guard centralManager.state == .poweredOn else {
            print("BLE: Central manager not ready")
            return
        }
        
        centralManager.scanForPeripherals(
            withServices: [BLEManager.serviceUUID],
            options: [CBCentralManagerScanOptionAllowDuplicatesKey: false]
        )
        isScanning = true
        print("BLE: Started scanning for InterMesh devices")
    }
    
    private func stopScanning() {
        centralManager.stopScan()
        isScanning = false
        print("BLE: Stopped scanning")
    }
    
    // MARK: - Private Methods - Connection Management
    
    private func disconnectAll() {
        for peripheral in connectedPeripherals.values {
            centralManager.cancelPeripheralConnection(peripheral)
        }
        connectedPeripherals.removeAll()
        connectedPeers.removeAll()
        discoveredPeripherals.removeAll()
        discoveredPeers.removeAll()
    }
    
    private func updatePeerConnection(identifier: UUID, connected: Bool) {
        if let index = discoveredPeers.firstIndex(where: { $0.identifier == identifier }) {
            discoveredPeers[index].isConnected = connected
            
            if connected {
                if !connectedPeers.contains(where: { $0.identifier == identifier }) {
                    connectedPeers.append(discoveredPeers[index])
                }
            } else {
                connectedPeers.removeAll { $0.identifier == identifier }
            }
        }
    }
}

// MARK: - CBCentralManagerDelegate

extension BLEManager: CBCentralManagerDelegate {
    
    func centralManagerDidUpdateState(_ central: CBCentralManager) {
        bleState = central.state
        
        switch central.state {
        case .poweredOn:
            print("BLE: Central manager powered on")
            connectionStatus = "BLE Ready"
        case .poweredOff:
            print("BLE: Bluetooth is powered off")
            connectionStatus = "Bluetooth Off"
            onError?("Bluetooth is powered off")
        case .unauthorized:
            print("BLE: Bluetooth unauthorized")
            connectionStatus = "Not Authorized"
            onError?("Bluetooth permission denied")
        case .unsupported:
            print("BLE: BLE not supported")
            connectionStatus = "Not Supported"
            onError?("BLE not supported on this device")
        default:
            connectionStatus = "BLE Unknown"
        }
    }
    
    func centralManager(_ central: CBCentralManager, didDiscover peripheral: CBPeripheral,
                        advertisementData: [String: Any], rssi RSSI: NSNumber) {
        let identifier = peripheral.identifier
        
        // Skip if already discovered
        if discoveredPeripherals[identifier] != nil {
            return
        }
        
        discoveredPeripherals[identifier] = peripheral
        
        let name = peripheral.name ?? advertisementData[CBAdvertisementDataLocalNameKey] as? String ?? "Unknown Device"
        
        let peer = BLEPeer(
            id: identifier.uuidString,
            name: name,
            identifier: identifier,
            platform: "unknown",
            rssi: RSSI.intValue
        )
        
        DispatchQueue.main.async {
            self.discoveredPeers.append(peer)
            self.onPeerDiscovered?(peer)
        }
        
        print("BLE: Discovered peer: \(name)")
        
        // Auto-connect to discovered peers
        central.connect(peripheral, options: nil)
    }
    
    func centralManager(_ central: CBCentralManager, didConnect peripheral: CBPeripheral) {
        print("BLE: Connected to \(peripheral.name ?? "Unknown")")
        
        peripheral.delegate = self
        connectedPeripherals[peripheral.identifier] = peripheral
        
        // Discover services
        peripheral.discoverServices([BLEManager.serviceUUID])
        
        updatePeerConnection(identifier: peripheral.identifier, connected: true)
    }
    
    func centralManager(_ central: CBCentralManager, didDisconnectPeripheral peripheral: CBPeripheral, error: Error?) {
        print("BLE: Disconnected from \(peripheral.name ?? "Unknown")")
        
        connectedPeripherals.removeValue(forKey: peripheral.identifier)
        peripheralCharacteristics.removeValue(forKey: peripheral.identifier)
        
        updatePeerConnection(identifier: peripheral.identifier, connected: false)
        
        DispatchQueue.main.async {
            self.onPeerDisconnected?(peripheral.identifier.uuidString)
        }
    }
    
    func centralManager(_ central: CBCentralManager, didFailToConnect peripheral: CBPeripheral, error: Error?) {
        print("BLE: Failed to connect to \(peripheral.name ?? "Unknown"): \(error?.localizedDescription ?? "Unknown error")")
    }
    
    func centralManager(_ central: CBCentralManager, willRestoreState dict: [String: Any]) {
        print("BLE: Restoring central manager state")
    }
}

// MARK: - CBPeripheralDelegate

extension BLEManager: CBPeripheralDelegate {
    
    func peripheral(_ peripheral: CBPeripheral, didDiscoverServices error: Error?) {
        guard error == nil else {
            print("BLE: Error discovering services: \(error!.localizedDescription)")
            return
        }
        
        guard let services = peripheral.services else { return }
        
        for service in services {
            if service.uuid == BLEManager.serviceUUID {
                peripheral.discoverCharacteristics([
                    BLEManager.deviceInfoCharUUID,
                    BLEManager.messageCharUUID,
                    BLEManager.meshDataCharUUID
                ], for: service)
            }
        }
    }
    
    func peripheral(_ peripheral: CBPeripheral, didDiscoverCharacteristicsFor service: CBService, error: Error?) {
        guard error == nil else {
            print("BLE: Error discovering characteristics: \(error!.localizedDescription)")
            return
        }
        
        guard let characteristics = service.characteristics else { return }
        
        var charMap: [CBUUID: CBCharacteristic] = [:]
        
        for characteristic in characteristics {
            charMap[characteristic.uuid] = characteristic
            
            // Read device info
            if characteristic.uuid == BLEManager.deviceInfoCharUUID {
                peripheral.readValue(for: characteristic)
            }
            
            // Subscribe to notifications
            if characteristic.uuid == BLEManager.messageCharUUID ||
               characteristic.uuid == BLEManager.meshDataCharUUID {
                peripheral.setNotifyValue(true, for: characteristic)
            }
        }
        
        peripheralCharacteristics[peripheral.identifier] = charMap
    }
    
    func peripheral(_ peripheral: CBPeripheral, didUpdateValueFor characteristic: CBCharacteristic, error: Error?) {
        guard error == nil, let data = characteristic.value else { return }
        
        if characteristic.uuid == BLEManager.deviceInfoCharUUID {
            // Parse device info: "deviceId|deviceName|platform"
            if let info = String(data: data, encoding: .utf8) {
                let parts = info.split(separator: "|")
                if parts.count >= 3 {
                    let platform = String(parts[2])
                    
                    // Update peer with platform info
                    if let index = discoveredPeers.firstIndex(where: { $0.identifier == peripheral.identifier }) {
                        DispatchQueue.main.async {
                            self.discoveredPeers[index].platform = platform
                            
                            // Update connected peer too
                            if let connectedIndex = self.connectedPeers.firstIndex(where: { $0.identifier == peripheral.identifier }) {
                                self.connectedPeers[connectedIndex].platform = platform
                            }
                            
                            self.onPeerConnected?(self.discoveredPeers[index])
                        }
                    }
                    
                    print("BLE: Peer info - \(parts[1]) (\(platform))")
                }
            }
        } else if characteristic.uuid == BLEManager.messageCharUUID ||
                  characteristic.uuid == BLEManager.meshDataCharUUID {
            // Handle received message
            let peerId = peripheral.identifier.uuidString
            
            DispatchQueue.main.async {
                if let message = String(data: data, encoding: .utf8) {
                    self.lastReceivedMessage = message
                }
                self.onMessageReceived?(peerId, data)
            }
        }
    }
}

// MARK: - CBPeripheralManagerDelegate

extension BLEManager: CBPeripheralManagerDelegate {
    
    func peripheralManagerDidUpdateState(_ peripheral: CBPeripheralManager) {
        switch peripheral.state {
        case .poweredOn:
            print("BLE: Peripheral manager powered on")
        case .poweredOff:
            print("BLE: Peripheral manager powered off")
        default:
            break
        }
    }
    
    func peripheralManager(_ peripheral: CBPeripheralManager, didAdd service: CBService, error: Error?) {
        if let error = error {
            print("BLE: Error adding service: \(error.localizedDescription)")
        } else {
            print("BLE: Service added successfully")
        }
    }
    
    func peripheralManagerDidStartAdvertising(_ peripheral: CBPeripheralManager, error: Error?) {
        if let error = error {
            print("BLE: Error starting advertising: \(error.localizedDescription)")
            DispatchQueue.main.async {
                self.isAdvertising = false
                self.onError?("Failed to start BLE advertising")
            }
        } else {
            print("BLE: Advertising started successfully")
        }
    }
    
    func peripheralManager(_ peripheral: CBPeripheralManager, central: CBCentral, didSubscribeTo characteristic: CBCharacteristic) {
        print("BLE: Central subscribed to \(characteristic.uuid)")
        connectedCentrals.append(central)
    }
    
    func peripheralManager(_ peripheral: CBPeripheralManager, central: CBCentral, didUnsubscribeFrom characteristic: CBCharacteristic) {
        print("BLE: Central unsubscribed from \(characteristic.uuid)")
        connectedCentrals.removeAll { $0.identifier == central.identifier }
    }
    
    func peripheralManager(_ peripheral: CBPeripheralManager, didReceiveRead request: CBATTRequest) {
        if request.characteristic.uuid == BLEManager.deviceInfoCharUUID {
            let deviceInfo = "\(localDeviceId)|\(localDeviceName)|ios"
            if let data = deviceInfo.data(using: .utf8) {
                request.value = data
                peripheral.respond(to: request, withResult: .success)
            } else {
                peripheral.respond(to: request, withResult: .unlikelyError)
            }
        } else {
            peripheral.respond(to: request, withResult: .attributeNotFound)
        }
    }
    
    func peripheralManager(_ peripheral: CBPeripheralManager, didReceiveWrite requests: [CBATTRequest]) {
        for request in requests {
            if let data = request.value {
                let charUUID = request.characteristic.uuid
                
                if charUUID == BLEManager.messageCharUUID || charUUID == BLEManager.meshDataCharUUID {
                    DispatchQueue.main.async {
                        if let message = String(data: data, encoding: .utf8) {
                            self.lastReceivedMessage = message
                        }
                        self.onMessageReceived?(request.central.identifier.uuidString, data)
                    }
                }
            }
            peripheral.respond(to: request, withResult: .success)
        }
    }
    
    func peripheralManager(_ peripheral: CBPeripheralManager, willRestoreState dict: [String: Any]) {
        print("BLE: Restoring peripheral manager state")
    }
}
