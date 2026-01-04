# Quick Start: Building Mobile Apps

## � Prerequisites

### For Android (Linux):
```bash
# 1. Install Android SDK (one-time setup)
./setup-android-sdk.sh
source ~/.bashrc

# 2. Verify installation
echo $ANDROID_HOME
```

## �🚀 Build Android APK (Linux - Easiest Option)

### 1. One-Command Build
```bash
./build-android.sh
```

This script will:
- ✅ Check all requirements
- ✅ Generate Android AAR library
- ✅ Build APK with UI
- ✅ Show installation instructions

### 2. Manual Build Steps (if needed)

```bash
# Step 1: Generate AAR
gomobile bind -target=android -o mobile/output/intermesh.aar ./mobile

# Step 2: Build APK
cd mobile/android-app
./gradlew assembleDebug

# APK location: app/build/outputs/apk/debug/app-debug.apk
```

### 3. Install on Device

```bash
# Enable USB debugging on Android device first!
adb install mobile/android-app/app/build/outputs/apk/debug/app-debug.apk
```

## 📱 What You'll Get

The Android app includes:
- ✅ Connect/Disconnect button
- ✅ Real-time connection status
- ✅ Network statistics (peers, proxies)
- ✅ Internet sharing toggle
- ✅ Request internet access button
- ✅ Live status messages

## 🍎 Build iOS App (Requires macOS OR GitHub Actions)

### Option 1: Using GitHub Actions (Recommended for Linux)

1. **Push to GitHub**:
```bash
git init
git add .
git commit -m "Initial commit"
git remote add origin <your-repo-url>
git push -u origin main
```

2. **Enable GitHub Actions**:
   - Go to your repository on GitHub
   - Click "Actions" tab
   - The iOS build will run automatically

3. **Download Framework**:
   - After build completes, go to Actions → Latest run
   - Download "intermesh-ios-framework" artifact
   - Transfer to your Mac/iOS device

### Option 2: On macOS (Local Build)

```bash
# Generate iOS framework
gomobile bind -target=ios -o mobile/output/Intermesh.xcframework ./mobile

# Then create iOS app in Xcode and import the framework
```

### Option 3: Cloud macOS Services

**MacinCloud** (Pay-per-hour):
1. Rent macOS VM: https://www.macincloud.com/
2. Upload your code
3. Run: `gomobile bind -target=ios ./mobile`
4. Build with Xcode

**GitHub Codespaces** (Free tier):
1. Create codespace with macOS
2. Build iOS framework
3. Download and use locally

## ⚡ Quick Test (No UI, but works immediately)

If you just want to test mesh functionality right now:

```bash
# Build simple test app
gomobile build -target=android -o test.apk ./cmd/mobile-demo

# Install
adb install test.apk
```

This creates a basic app that demonstrates mesh connectivity without full UI.

## 🧪 Testing Connectivity

### What to Test:

1. **Peer Discovery**:
   - Install APK on 2+ Android devices
   - Open app on all devices
   - Tap "Connect to Mesh"
   - Watch peers count increase

2. **Internet Sharing**:
   - Device A: Enable mobile data/WiFi
   - Device A: Enable "Share Internet" toggle
   - Device B: Tap "Request Internet Access"
   - Device B should show "Internet Available"

3. **Multi-hop Routing**:
   - Device A ↔ Device B ↔ Device C
   - A and C are out of range
   - B routes packets between them

### Required Permissions:

The app will request:
- ✅ Location (for WiFi scanning)
- ✅ WiFi state access
- ✅ Network access

## 📊 What Works Now vs Later

### ✅ Currently Working:
- App UI and interaction
- Connection state management
- Statistics tracking
- Internet status detection
- Listener/callback system

### 🔧 Needs Implementation:
- Actual WiFi Direct communication
- Real peer discovery over WiFi
- Packet routing over WiFi
- Encryption layer

## 🐛 Troubleshooting

### "gomobile: command not found"
```bash
go install golang.org/x/mobile/cmd/gomobile@latest
export PATH=$PATH:$(go env GOPATH)/bin
gomobile init
```

### "ANDROID_HOME not set"
```bash
# Download Android SDK from:
# https://developer.android.com/studio#command-tools

export ANDROID_HOME=$HOME/Android/Sdk
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools
```

### "adb: device not found"
```bash
# Enable USB debugging on Android device
# Settings → About Phone → Tap Build Number 7 times
# Settings → Developer Options → USB Debugging

# Then:
adb kill-server
adb start-server
adb devices
```

### APK installs but crashes
```bash
# Check logs:
adb logcat | grep InterMesh

# Common issues:
# - Missing permissions
# - AAR not regenerated after code changes
```

## 🎯 Next Steps

After building and installing:

1. **Test on single device**: Verify UI works, buttons respond
2. **Test connectivity**: Install on 2 devices, test peer discovery
3. **Implement WiFi protocol**: Add actual WiFi Direct/mesh networking
4. **Add encryption**: Secure the mesh traffic
5. **Optimize**: Performance tuning and battery optimization

## 📚 More Information

- Full mobile guide: [docs/MOBILE.md](docs/MOBILE.md)
- Architecture: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- Development: [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
- Building details: [docs/BUILDING_MOBILE.md](docs/BUILDING_MOBILE.md)

---

**Ready to build?**
```bash
./build-android.sh
```
