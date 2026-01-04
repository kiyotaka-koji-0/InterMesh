// +build android

package intermesh

// AndroidActivity is the main Android activity
type AndroidActivity struct {
	controller           *MobileUIController
	connectButton        *AndroidButton
	sharingToggle        *AndroidToggle
	internetButton       *AndroidButton
	statusTextView       *AndroidTextView
	proxyCountTextView   *AndroidTextView
	peerCountTextView    *AndroidTextView
}

// AndroidButton represents an Android button
type AndroidButton struct {
	Title       string
	Enabled     bool
	OnClick     func()
}

// AndroidToggle represents an Android toggle switch
type AndroidToggle struct {
	Title       string
	Enabled     bool
	Checked     bool
	OnChecked   func(bool)
}

// AndroidTextView represents an Android text view
type AndroidTextView struct {
	Text string
}

// NewAndroidActivity creates a new Android activity
func NewAndroidActivity(app *MobileApp) *AndroidActivity {
	controller := NewMobileUIController(app)

	activity := &AndroidActivity{
		controller:          controller,
		connectButton:       &AndroidButton{Title: "Connect"},
		sharingToggle:       &AndroidToggle{Title: "Share Internet"},
		internetButton:      &AndroidButton{Title: "Request Internet"},
		statusTextView:      &AndroidTextView{},
		proxyCountTextView:  &AndroidTextView{},
		peerCountTextView:   &AndroidTextView{},
	}

	// Setup button click listeners
	activity.connectButton.OnClick = func() {
		if err := controller.ToggleConnectButton(); err == nil {
			activity.updateUI()
		}
	}

	activity.sharingToggle.OnChecked = func(checked bool) {
		if err := controller.ToggleInternetSharingSwitch(); err == nil {
			activity.updateUI()
		}
	}

	activity.internetButton.OnClick = func() {
		if err := controller.ToggleInternetAccessButton(); err == nil {
			activity.updateUI()
		}
	}

	// Setup callbacks
	controller.SetUIUpdateCallback(func() {
		activity.updateUI()
	})

	controller.SetStatusChangeCallback(func(status string) {
		activity.statusTextView.Text = status
		activity.updateUI()
	})

	controller.SetErrorCallback(func(errMsg string) {
		activity.statusTextView.Text = "Error: " + errMsg
		activity.updateUI()
	})

	activity.updateUI()
	return activity
}

// updateUI refreshes the Android UI
func (a *AndroidActivity) updateUI() {
	// Update button states
	connectButtonState := a.controller.GetConnectionButtonState()
	a.connectButton.Title = connectButtonState.Title
	a.connectButton.Enabled = connectButtonState.Enabled

	// Update toggle states
	sharingToggleState := a.controller.GetInternetSharingToggleState()
	a.sharingToggle.Title = sharingToggleState.Title
	a.sharingToggle.Enabled = sharingToggleState.Enabled
	a.sharingToggle.Checked = sharingToggleState.Checked

	// Update internet button states
	internetButtonState := a.controller.GetInternetAccessButtonState()
	a.internetButton.Title = internetButtonState.Title
	a.internetButton.Enabled = internetButtonState.Enabled

	// Update text views
	stats := a.controller.app.GetNetworkStats()
	a.statusTextView.Text = a.controller.GetStatusDisplay()
	a.proxyCountTextView.Text = "Available Proxies: " + string(rune(stats.AvailableProxies))
	a.peerCountTextView.Text = "Connected Peers: " + string(rune(stats.PeerCount))
}

// GetConnectButton returns the connect button
func (a *AndroidActivity) GetConnectButton() *AndroidButton {
	return a.connectButton
}

// GetSharingToggle returns the sharing toggle
func (a *AndroidActivity) GetSharingToggle() *AndroidToggle {
	return a.sharingToggle
}

// GetInternetButton returns the internet access button
func (a *AndroidActivity) GetInternetButton() *AndroidButton {
	return a.internetButton
}

// GetStatusTextView returns the status text view
func (a *AndroidActivity) GetStatusTextView() *AndroidTextView {
	return a.statusTextView
}

// GetProxyCountTextView returns the proxy count text view
func (a *AndroidActivity) GetProxyCountTextView() *AndroidTextView {
	return a.proxyCountTextView
}

// GetPeerCountTextView returns the peer count text view
func (a *AndroidActivity) GetPeerCountTextView() *AndroidTextView {
	return a.peerCountTextView
}
