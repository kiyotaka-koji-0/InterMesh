//go:build ios
// +build ios

package intermesh

// iOSViewController is the main iOS view controller
type iOSViewController struct {
	controller     *MobileUIController
	connectButton  *iOSButton
	sharingToggle  *iOSToggle
	internetButton *iOSButton
	statusLabel    *iOSLabel
}

// iOSButton represents an iOS button
type iOSButton struct {
	Title   string
	Enabled bool
	Target  func()
}

// iOSToggle represents an iOS toggle switch
type iOSToggle struct {
	Title    string
	Enabled  bool
	Checked  bool
	OnToggle func(bool)
}

// iOSLabel represents an iOS label
type iOSLabel struct {
	Text string
}

// NewiOSViewController creates a new iOS view controller
func NewiOSViewController(app *MobileApp) *iOSViewController {
	controller := NewMobileUIController(app)

	viewController := &iOSViewController{
		controller:     controller,
		connectButton:  &iOSButton{Title: "Connect"},
		sharingToggle:  &iOSToggle{Title: "Share Internet"},
		internetButton: &iOSButton{Title: "Request Internet"},
		statusLabel:    &iOSLabel{},
	}

	// Setup button targets
	viewController.connectButton.Target = func() {
		if err := controller.ToggleConnectButton(); err == nil {
			viewController.updateUI()
		}
	}

	viewController.sharingToggle.OnToggle = func(checked bool) {
		if err := controller.ToggleInternetSharingSwitch(); err == nil {
			viewController.updateUI()
		}
	}

	viewController.internetButton.Target = func() {
		if err := controller.ToggleInternetAccessButton(); err == nil {
			viewController.updateUI()
		}
	}

	// Setup callbacks
	controller.SetUIUpdateCallback(func() {
		viewController.updateUI()
	})

	controller.SetStatusChangeCallback(func(status string) {
		viewController.statusLabel.Text = status
		viewController.updateUI()
	})

	viewController.updateUI()
	return viewController
}

// updateUI refreshes the iOS UI
func (vc *iOSViewController) updateUI() {
	// Update button states
	connectButtonState := vc.controller.GetConnectionButtonState()
	vc.connectButton.Title = connectButtonState.Title
	vc.connectButton.Enabled = connectButtonState.Enabled

	// Update toggle states
	sharingToggleState := vc.controller.GetInternetSharingToggleState()
	vc.sharingToggle.Title = sharingToggleState.Title
	vc.sharingToggle.Enabled = sharingToggleState.Enabled
	vc.sharingToggle.Checked = sharingToggleState.Checked

	// Update internet button states
	internetButtonState := vc.controller.GetInternetAccessButtonState()
	vc.internetButton.Title = internetButtonState.Title
	vc.internetButton.Enabled = internetButtonState.Enabled

	// Update status label
	vc.statusLabel.Text = vc.controller.GetStatusDisplay()
}

// GetConnectButton returns the connect button
func (vc *iOSViewController) GetConnectButton() *iOSButton {
	return vc.connectButton
}

// GetSharingToggle returns the sharing toggle
func (vc *iOSViewController) GetSharingToggle() *iOSToggle {
	return vc.sharingToggle
}

// GetInternetButton returns the internet access button
func (vc *iOSViewController) GetInternetButton() *iOSButton {
	return vc.internetButton
}

// GetStatusLabel returns the status label
func (vc *iOSViewController) GetStatusLabel() *iOSLabel {
	return vc.statusLabel
}
