# Mobile Implementation Guide

## Overview

InterMesh provides a complete mobile implementation for iOS and Android devices. The mobile layer includes:

1. **MeshApp** - Core application instance managing mesh networking
2. **MobileUIController** - UI state management and event handling
3. **Platform-specific implementations** - iOS and Android UI components

## Architecture

### Core Components

```
┌─────────────────────────────────────────┐
│       iOS/Android Application           │
├─────────────────────────────────────────┤
│    MobileUIController (State Manager)    │
├─────────────────────────────────────────┤
│      MobileApp (Wrapper Interface)       │
├─────────────────────────────────────────┤
│        MeshApp (Core Application)        │
├─────────────────────────────────────────┤
│  Mesh Networking (Node, Router, Proxy)  │
└─────────────────────────────────────────┘
```

## MeshApp - Core Application

The `MeshApp` is the main application instance that manages all mesh networking operations.

### Creating a MeshApp Instance

```go
app := mesh.NewMeshApp(
    "device-id",
    "Device Name",
    "192.168.1.100",
    "aa:bb:cc:dd:ee:ff",
)
```

### Basic Operations

```go
// Start the application
err := app.Start()

// Connect to mesh network
err := app.ConnectToNetwork()

// Enable internet sharing
err := app.EnableInternetSharing()

// Request internet access through proxies
proxyID, err := app.RequestInternetAccess()

// Get network statistics
stats := app.GetNetworkStats()

// Stop the application
app.Stop()
```

## MobileApp - Mobile Wrapper

The `MobileApp` provides a simplified interface for mobile platforms.

### Creating a MobileApp Instance

```go
// Go code
app := intermesh.NewMobileApp("device-id", "Device Name", "192.168.1.100", "aa:bb:cc:dd:ee:ff")
```

### Available Methods

- `Start()` - Initialize the app
- `Stop()` - Stop the app
- `ConnectToNetwork()` - Connect to mesh
- `DisconnectFromNetwork()` - Disconnect from mesh
- `IsConnected()` - Check connection status
- `EnableInternetSharing()` - Share device internet
- `DisableInternetSharing()` - Stop sharing
- `IsInternetSharingEnabled()` - Check sharing status
- `SetInternetStatus(bool)` - Update device internet status
- `HasInternet()` - Check device internet
- `RequestInternetAccess()` - Get internet through proxy
- `ReleaseInternetAccess(proxyID)` - Release proxy connection
- `GetNetworkStats()` - Get network statistics
- `GetAvailableProxyCount()` - Count available proxies
- `GetConnectedPeerCount()` - Count connected peers
- `GetNodeID()` - Get device ID
- `GetNodeName()` - Get device name
- `SetNodeName(name)` - Set device name

## MobileUIController - State Management

The `MobileUIController` manages UI state and handles all user interactions.

### Creating a Controller

```go
mobileApp := intermesh.NewMobileApp(...)
controller := intermesh.NewMobileUIController(mobileApp)
```

### Setting Callbacks

```go
// UI update callback
controller.SetUIUpdateCallback(func() {
    // Refresh UI elements
})

// Error callback
controller.SetErrorCallback(func(errMsg string) {
    // Display error to user
})

// Status change callback
controller.SetStatusChangeCallback(func(status string) {
    // Update status display
})
```

### User Actions

```go
// Connect/Disconnect button
err := controller.ToggleConnectButton()

// Internet sharing toggle
err := controller.ToggleInternetSharingSwitch()

// Request internet access
err := controller.ToggleInternetAccessButton()
```

### Getting UI State

```go
// Get button states
connectBtn := controller.GetConnectionButtonState()
internetBtn := controller.GetInternetAccessButtonState()

// Get toggle states
sharingToggle := controller.GetInternetSharingToggleState()

// Get status display
status := controller.GetStatusDisplay()

// Get detailed stats
stats := controller.GetDetailedStats()
```

## iOS Implementation

### Creating iOS ViewController

```go
app := intermesh.NewMobileApp(...)
viewController := intermesh.NewiOSViewController(app)
```

### Accessing iOS UI Elements

```go
// Get button references
connectButton := viewController.GetConnectButton()
internetButton := viewController.GetInternetButton()

// Get toggle reference
sharingToggle := viewController.GetSharingToggle()

// Get status label
statusLabel := viewController.GetStatusLabel()

// Button properties
connectButton.Title  // String
connectButton.Enabled // Boolean

// Toggle properties
sharingToggle.Title   // String
sharingToggle.Enabled // Boolean
sharingToggle.Checked // Boolean
```

### iOS UI Layout (SwiftUI example)

```swift
VStack(spacing: 20) {
    // Connection Button
    Button(action: { viewController.connectButton.Target?() }) {
        Text(viewController.connectButton.Title)
    }
    .disabled(!viewController.connectButton.Enabled)
    
    // Internet Sharing Toggle
    Toggle(isOn: Binding(
        get: { viewController.sharingToggle.Checked },
        set: { viewController.sharingToggle.OnToggle?($0) }
    )) {
        Text(viewController.sharingToggle.Title)
    }
    .disabled(!viewController.sharingToggle.Enabled)
    
    // Internet Access Button
    Button(action: { viewController.internetButton.Target?() }) {
        Text(viewController.internetButton.Title)
    }
    .disabled(!viewController.internetButton.Enabled)
    
    // Status Display
    Text(viewController.statusLabel.Text)
}
```

## Android Implementation

### Creating Android Activity

```go
app := intermesh.NewMobileApp(...)
activity := intermesh.NewAndroidActivity(app)
```

### Accessing Android UI Elements

```go
// Get button references
connectButton := activity.GetConnectButton()
internetButton := activity.GetInternetButton()

// Get toggle reference
sharingToggle := activity.GetSharingToggle()

// Get text view references
statusTextView := activity.GetStatusTextView()
proxyCountTextView := activity.GetProxyCountTextView()
peerCountTextView := activity.GetPeerCountTextView()

// Button properties
connectButton.Title     // String
connectButton.Enabled   // Boolean
connectButton.OnClick   // Callback function

// Toggle properties
sharingToggle.Title     // String
sharingToggle.Enabled   // Boolean
sharingToggle.Checked   // Boolean
sharingToggle.OnChecked // Callback function
```

### Android UI Layout (Jetpack Compose example)

```kotlin
Column(modifier = Modifier.padding(16.dp)) {
    // Connection Button
    Button(
        onClick = { activity.connectButton.OnClick?.invoke() },
        enabled = activity.connectButton.Enabled
    ) {
        Text(activity.connectButton.Title)
    }
    
    // Internet Sharing Toggle
    Row(verticalAlignment = Alignment.CenterVertically) {
        Text(activity.sharingToggle.Title)
        Switch(
            checked = activity.sharingToggle.Checked,
            onCheckedChange = { activity.sharingToggle.OnChecked?.invoke(it) },
            enabled = activity.sharingToggle.Enabled
        )
    }
    
    // Internet Access Button
    Button(
        onClick = { activity.internetButton.OnClick?.invoke() },
        enabled = activity.internetButton.Enabled
    ) {
        Text(activity.internetButton.Title)
    }
    
    // Status Displays
    Text(activity.statusTextView.Text)
    Text(activity.proxyCountTextView.Text)
    Text(activity.peerCountTextView.Text)
}
```

## Event Flow

### Connection Flow

```
User taps "Connect" button
         ↓
ToggleConnectButton() called
         ↓
ConnectToNetwork() executed
         ↓
OnConnectionStateChanged() fired
         ↓
UI updates via callback
```

### Internet Sharing Flow

```
User enables sharing toggle
         ↓
ToggleInternetSharingSwitch() called
         ↓
EnableInternetSharing() executed
         ↓
ProxyManager registers device as proxy
         ↓
OnStatusChange callback fires
         ↓
UI updates
```

### Internet Access Flow

```
User taps "Request Internet"
         ↓
ToggleInternetAccessButton() called
         ↓
RequestInternetAccess() finds best proxy
         ↓
CreateProxyConnection() establishes connection
         ↓
OnStatusChange callback fires with proxy info
         ↓
UI updates to show "Connected via [ProxyID]"
```

## Network Statistics

Get current network information:

```go
stats := mobileApp.GetNetworkStats()

// Available fields:
stats.NodeID                  // Device identifier
stats.PeerCount               // Connected peers
stats.AvailableProxies        // Available proxy devices
stats.InternetStatus          // Device has internet
stats.InternetSharingEnabled  // Sharing is active
stats.ConnectedNetworks       // Sub-mesh networks
stats.DataTransferred         // Bytes transferred
```

## Error Handling

All operations return errors that should be displayed to the user:

```go
if err := controller.ToggleConnectButton(); err != nil {
    // Display error message to user
    showError(err.Error())
}
```

Common errors:
- "not connected to mesh" - Device not in mesh network
- "device does not have internet connectivity" - Can't share internet
- "no available proxy" - No proxies for internet access
- Network-related errors

## Global Mesh vs Invitation System

The implementation uses a **global mesh approach**:

- Devices automatically discover nearby peers
- No invitation system required
- Any device can become part of the mesh
- Internet-connected devices automatically become proxies
- Non-internet devices discover and use available proxies
- Peer discovery runs continuously in the background

## Best Practices

1. **Always call Stop()** - Cleanup resources when app closes
2. **Handle errors gracefully** - Show user-friendly error messages
3. **Update UI on callbacks** - Don't block callback threads
4. **Check connection status** - Before performing network operations
5. **Set internet status** - Update when device connectivity changes
6. **Register listeners early** - Before performing operations
7. **Refresh UI periodically** - Network stats may change frequently

## Testing

Unit tests are provided in `pkg/mesh/app_test.go`:

```bash
go test -v ./pkg/mesh
```

All 17 tests verify:
- App creation and initialization
- Connection/disconnection
- Internet status management
- Network statistics
- Internet access requests
- Internet sharing
- Listener callbacks
- Proxy management
- Peer management
