package intermesh

// This file defines the UI binding interfaces that iOS and Android can implement

// UIButton represents a button action
type UIButton struct {
	Title   string
	Enabled bool
}

// UIToggle represents a toggle switch state
type UIToggle struct {
	Title   string
	Enabled bool
	Checked bool
}

// UIText represents text display
type UIText struct {
	Label   string
	Content string
}

// MobileUIController manages the mobile UI state and interactions
type MobileUIController struct {
	app            *MobileApp
	onUIUpdate     func()
	onError        func(string)
	onStatusChange func(string)
}

// NewMobileUIController creates a new UI controller
func NewMobileUIController(app *MobileApp) *MobileUIController {
	controller := &MobileUIController{
		app: app,
	}

	// Register listeners for updates
	app.app.AddConnectionListener(&MobileConnectionListener{
		onConnected: func() {
			controller.onStatusUpdate("Connected to mesh network")
		},
		onError: func(errMsg string) {
			controller.onError(errMsg)
		},
	})

	return controller
}

// SetUIUpdateCallback sets the callback for UI updates
func (controller *MobileUIController) SetUIUpdateCallback(callback func()) {
	controller.onUIUpdate = callback
}

// SetErrorCallback sets the callback for errors
func (controller *MobileUIController) SetErrorCallback(callback func(string)) {
	controller.onError = callback
}

// SetStatusChangeCallback sets the callback for status changes
func (controller *MobileUIController) SetStatusChangeCallback(callback func(string)) {
	controller.onStatusChange = callback
}

// onStatusUpdate calls the status change callback
func (controller *MobileUIController) onStatusUpdate(status string) {
	if controller.onStatusChange != nil {
		controller.onStatusChange(status)
	}
	if controller.onUIUpdate != nil {
		controller.onUIUpdate()
	}
}

// ToggleConnectButton handles connect/disconnect button action
func (controller *MobileUIController) ToggleConnectButton() error {
	if controller.app.IsConnected() {
		controller.app.DisconnectFromNetwork()
		controller.onStatusUpdate("Disconnected from mesh network")
	} else {
		if err := controller.app.Start(); err != nil {
			controller.onError(err.Error())
			return err
		}
		if err := controller.app.ConnectToNetwork(); err != nil {
			controller.onError(err.Error())
			return err
		}
		controller.onStatusUpdate("Connected to mesh network")
	}
	return nil
}

// ToggleInternetSharingSwitch handles internet sharing toggle
func (controller *MobileUIController) ToggleInternetSharingSwitch() error {
	if !controller.app.IsConnected() {
		controller.onError("Device is not connected to mesh network")
		return NewMeshError("not connected to mesh")
	}

	if controller.app.IsInternetSharingEnabled() {
		controller.app.DisableInternetSharing()
		controller.onStatusUpdate("Internet sharing disabled")
	} else {
		if err := controller.app.EnableInternetSharing(); err != nil {
			controller.onError(err.Error())
			return err
		}
		controller.onStatusUpdate("Internet sharing enabled - device is now a proxy")
	}
	return nil
}

// ToggleInternetAccessButton handles internet access request
func (controller *MobileUIController) ToggleInternetAccessButton() error {
	if !controller.app.IsConnected() {
		controller.onError("Device is not connected to mesh network")
		return NewMeshError("not connected to mesh")
	}

	if controller.app.HasInternet() {
		controller.onStatusUpdate("Device already has internet connectivity")
		return nil
	}

	proxyID, err := controller.app.RequestInternetAccess()
	if err != nil {
		controller.onError(err.Error())
		return err
	}

	message := "Connected to internet through proxy: " + proxyID
	controller.onStatusUpdate(message)
	return nil
}

// GetConnectionButtonState returns the state of the connect button
func (controller *MobileUIController) GetConnectionButtonState() *UIButton {
	isConnected := controller.app.IsConnected()
	title := "Disconnect"
	if !isConnected {
		title = "Connect"
	}
	return &UIButton{
		Title:   title,
		Enabled: true,
	}
}

// GetInternetSharingToggleState returns the state of the internet sharing toggle
func (controller *MobileUIController) GetInternetSharingToggleState() *UIToggle {
	isConnected := controller.app.IsConnected()
	hasInternet := controller.app.HasInternet()
	isSharing := controller.app.IsInternetSharingEnabled()

	return &UIToggle{
		Title:   "Share Internet",
		Enabled: isConnected && hasInternet,
		Checked: isSharing,
	}
}

// GetInternetAccessButtonState returns the state of the internet access button
func (controller *MobileUIController) GetInternetAccessButtonState() *UIButton {
	isConnected := controller.app.IsConnected()
	hasInternet := controller.app.HasInternet()

	title := "Request Internet Access"
	if hasInternet {
		title = "Using Local Internet"
	}

	return &UIButton{
		Title:   title,
		Enabled: isConnected && !hasInternet,
	}
}

// GetStatusDisplay returns the current status for display
func (controller *MobileUIController) GetStatusDisplay() string {
	stats := controller.app.GetNetworkStats()

	status := "Status:\n"
	status += "Connected: "
	if controller.app.IsConnected() {
		status += "Yes\n"
	} else {
		status += "No\n"
	}

	status += "Peers: " + string(rune(stats.PeerCount)) + "\n"
	status += "Available Proxies: " + string(rune(stats.AvailableProxies)) + "\n"

	internetStr := "No"
	if stats.InternetStatus {
		internetStr = "Yes"
	}
	status += "Internet Access: " + internetStr + "\n"

	sharingStr := "Off"
	if stats.InternetSharingEnabled {
		sharingStr = "On"
	}
	status += "Sharing: " + sharingStr + "\n"

	return status
}

// GetDetailedStats returns detailed network statistics
func (controller *MobileUIController) GetDetailedStats() map[string]string {
	stats := controller.app.GetNetworkStats()

	m := make(map[string]string)
	m["node_id"] = stats.NodeID
	m["connected_peers"] = string(rune(stats.PeerCount))
	m["available_proxies"] = string(rune(stats.AvailableProxies))
	m["has_internet"] = boolToString(stats.InternetStatus)
	m["sharing_enabled"] = boolToString(stats.InternetSharingEnabled)
	m["connected_networks"] = string(rune(stats.ConnectedNetworks))

	return m
}

// Helper function to convert bool to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// RefreshUI triggers a UI refresh
func (controller *MobileUIController) RefreshUI() {
	if controller.onUIUpdate != nil {
		controller.onUIUpdate()
	}
}

// NewMeshError creates a new mesh error for mobile use
func NewMeshError(message string) error {
	return &meshError{message: message}
}

type meshError struct {
	message string
}

func (e *meshError) Error() string {
	return e.message
}
