# iOS Build and Installation Guide

This guide will help you build the InterMesh iOS app and install it on your iPad.

## Prerequisites

- **macOS Computer** (required for iOS development)
- **Xcode 15 or later** (free from the Mac App Store)
- **Apple ID** (free account works for development)
- **iPad** running iOS 15.0 or later
- **Lightning/USB-C cable** to connect iPad to Mac

## Option 1: Build Using GitHub Actions (Recommended)

GitHub Actions will automatically build the iOS framework and app whenever you push code.

### Steps:

1. **Wait for GitHub Actions to complete**:
   - Go to your repository on GitHub
   - Click the "Actions" tab
   - Wait for the "Build iOS Framework and App" workflow to complete
   - This runs automatically on every push to `main` branch

2. **Download the artifacts**:
   - Click on the completed workflow run
   - Scroll down to "Artifacts" section
   - Download:
     - `intermesh-ios-framework` (the Go mobile framework)
     - `InterMesh-iOS-App` (the iOS app IPA)

3. **Install on iPad using Xcode**:
   ```bash
   # On your Mac, after downloading the IPA:
   # Connect your iPad via USB
   # Open Xcode and go to: Window > Devices and Simulators
   # Drag the IPA file onto your iPad device
   ```

## Option 2: Build Locally on Your Mac

### Step 1: Install Prerequisites

```bash
# Install Xcode Command Line Tools
xcode-select --install

# Install Go (if not already installed)
brew install go

# Install gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
export PATH=$PATH:$(go env GOPATH)/bin
gomobile init
```

### Step 2: Clone and Build

```bash
# Clone the repository
git clone https://github.com/kiyotaka-koji-0/InterMesh.git
cd InterMesh

# Build the iOS framework
mkdir -p mobile/output
gomobile bind -target=ios -o mobile/output/Intermesh.xcframework ./mobile

# This creates: mobile/output/Intermesh.xcframework
```

### Step 3: Open in Xcode

```bash
# Open the iOS project
open mobile/ios-app/InterMesh.xcodeproj
```

### Step 4: Configure Signing

1. **In Xcode**:
   - Select the "InterMesh" project in the left sidebar
   - Select the "InterMesh" target
   - Go to "Signing & Capabilities" tab
   - Under "Team", click "Add Account..." if you haven't added your Apple ID
   - Sign in with your Apple ID (free account works!)
   - Select your team from the dropdown
   - Xcode will automatically create a provisioning profile

2. **Change Bundle Identifier** (optional but recommended):
   - Under "Signing & Capabilities", change the Bundle Identifier to something unique
   - For example: `com.yourname.intermesh`

### Step 5: Build and Install on iPad

#### Option A: Direct Installation (Easiest)

1. **Connect your iPad** to your Mac via USB cable

2. **Trust the computer**:
   - On your iPad, you'll see a prompt "Trust This Computer?"
   - Tap "Trust" and enter your iPad passcode

3. **Select your iPad as the build destination**:
   - In Xcode, at the top near the Run button
   - Click the device dropdown (might say "Any iOS Device")
   - Select your iPad from the list

4. **Build and Run**:
   - Click the ▶️ (Play) button in Xcode, or press `Cmd + R`
   - Xcode will build the app and install it on your iPad
   - **First time**: You may see "Untrusted Developer" error on iPad

5. **Trust the Developer Certificate** (first time only):
   - On your iPad: Settings > General > VPN & Device Management
   - Under "Developer App", tap on your Apple ID
   - Tap "Trust [Your Apple ID]"
   - Tap "Trust" again in the confirmation dialog

6. **Launch the app**:
   - The InterMesh app should now be on your iPad home screen
   - Tap to open it!

#### Option B: Build IPA for Distribution

If you want to build an IPA file to install on multiple devices:

```bash
# Build for iOS device (not simulator)
cd mobile/ios-app
xcodebuild clean archive \
  -project InterMesh.xcodeproj \
  -scheme InterMesh \
  -sdk iphoneos \
  -configuration Release \
  -archivePath ./build/InterMesh.xcarchive

# Export the IPA
xcodebuild -exportArchive \
  -archivePath ./build/InterMesh.xcarchive \
  -exportOptionsPlist ExportOptions.plist \
  -exportPath ./build \
  -allowProvisioningUpdates

# The IPA will be at: ./build/InterMesh.ipa
```

**Install the IPA**:
- Use Xcode: Window > Devices and Simulators > drag IPA to your iPad
- Or use Apple Configurator 2 (free from Mac App Store)
- Or use iOS App Signer + AltStore/Sideloadly

## Option 3: TestFlight Distribution (For Multiple Devices)

If you want to share the app with others or install it on multiple iPads:

### Prerequisites:
- **Apple Developer Program** membership ($99/year) - **Required for TestFlight**
- Or use a free alternative like **AltStore** or **Sideloadly**

### Using TestFlight (Paid Apple Developer Account):

1. **Join Apple Developer Program**:
   - Go to https://developer.apple.com/programs/
   - Enroll with your Apple ID ($99/year)

2. **Create App in App Store Connect**:
   - Go to https://appstoreconnect.apple.com/
   - Click "+" > "New App"
   - Fill in app details:
     - Platform: iOS
     - Name: InterMesh
     - Bundle ID: com.intermesh.app (or your custom ID)

3. **Archive and Upload**:
   ```bash
   # In Xcode, select "Any iOS Device" as destination
   # Go to: Product > Archive
   # When archive completes, click "Distribute App"
   # Choose "TestFlight & App Store"
   # Follow the wizard to upload
   ```

4. **Invite Testers**:
   - In App Store Connect, go to TestFlight tab
   - Add internal testers (Apple ID email addresses)
   - Once uploaded build is processed, tap "Notify Testers"

5. **Install via TestFlight**:
   - Testers receive email invitation
   - Install "TestFlight" app from App Store on iPad
   - Open invitation link and install InterMesh

### Using AltStore (Free Alternative):

1. **Install AltStore on Mac**:
   - Download from https://altstore.io/
   - Follow installation instructions

2. **Install AltServer on iPad**:
   - Connect iPad to Mac
   - In AltStore menu, select "Install AltStore" > your iPad

3. **Install InterMesh IPA**:
   - On iPad, open AltStore app
   - Tap "My Apps" > "+" button
   - Select the InterMesh.ipa file
   - Note: Free accounts can only have 3 apps, refresh every 7 days

## Troubleshooting

### "Failed to build framework"
```bash
# Make sure go.mod has the mobile dependency
cd InterMesh
go get golang.org/x/mobile
go mod tidy

# Reinitialize gomobile
gomobile init

# Try building again
gomobile bind -target=ios -o mobile/output/Intermesh.xcframework ./mobile
```

### "Code signing error"
- Make sure you're signed in with your Apple ID in Xcode
- Go to Xcode > Preferences > Accounts > Add your Apple ID
- Select your team in project settings

### "Untrusted Developer"
- Settings > General > VPN & Device Management
- Tap your developer certificate
- Tap "Trust"

### "Failed to verify code signature"
- Clean build folder: Xcode > Product > Clean Build Folder
- Delete DerivedData: `rm -rf ~/Library/Developer/Xcode/DerivedData/InterMesh-*`
- Restart Xcode and rebuild

### "App crashes on launch"
```bash
# Check the console in Xcode
# Connect iPad and select Window > Devices and Simulators
# Select your iPad > View Device Logs
# Look for crash logs related to InterMesh
```

### "Cannot find 'IntermeshNewMobileApp' in scope"
- Make sure the Intermesh.xcframework is properly linked
- Check Framework Search Paths in Build Settings includes `$(PROJECT_DIR)/../output`
- Clean and rebuild the project

## Features

The InterMesh iOS app includes:

- ✅ **Mesh Network Connection**: Connect to the InterMesh network
- ✅ **Internet Sharing**: Share your iPad's internet connection
- ✅ **Request Internet**: Request internet from other devices
- ✅ **Network Statistics**: View connected peers and available proxies
- ✅ **Device Information**: View your device ID and connection status
- ✅ **Modern UI**: Clean SwiftUI interface optimized for iPad

## Permissions

The app requires the following permissions:

- **Local Network Access**: To discover nearby devices
- **Location (When In Use)**: Required for WiFi functionality on iOS
  - Note: No location data is collected or stored

When you first launch the app, you'll be prompted to grant these permissions. Make sure to allow them for the app to function properly.

## Using the App

1. **Launch InterMesh** on your iPad
2. **Grant Permissions** when prompted
3. **Tap "Connect to Mesh"** to join the network
4. **Enable "Share Internet"** to share your connection (optional)
5. **Tap "Request Internet Access"** to find available proxies
6. **View Statistics** to see connected peers and proxies

## Network Requirements

For the mesh network to work:
- Multiple devices must be running InterMesh
- Devices should be on the same WiFi network (or in range for WiFi Direct/Mesh)
- Local network permissions must be granted

## Building for App Store (Future)

To publish to the App Store:

1. Complete all Apple Developer Program requirements
2. Add app icons (1024x1024 required)
3. Create app screenshots for all required device sizes
4. Prepare app description and metadata
5. Submit for App Review (usually takes 1-2 days)

For now, ad-hoc distribution or TestFlight is recommended for testing.

## Need Help?

- Check the GitHub Issues: https://github.com/kiyotaka-koji-0/InterMesh/issues
- Review the main README: https://github.com/kiyotaka-koji-0/InterMesh
- Apple Developer Documentation: https://developer.apple.com/documentation/

---

**Quick Start Summary:**

```bash
# 1. Install prerequisites
xcode-select --install
brew install go
go install golang.org/x/mobile/cmd/gomobile@latest

# 2. Build framework
cd InterMesh
gomobile bind -target=ios -o mobile/output/Intermesh.xcframework ./mobile

# 3. Open in Xcode
open mobile/ios-app/InterMesh.xcodeproj

# 4. Connect iPad, select it in Xcode, press Run ▶️

# 5. On iPad: Trust developer in Settings

# 6. Launch InterMesh app!
```

Enjoy using InterMesh on your iPad! 🎉
