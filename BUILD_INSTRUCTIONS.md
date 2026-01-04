# Building InterMesh Mobile Apps - Complete Guide

## Summary

You have a complete Android app ready to build! Here's what you need to do:

## ✅ What's Ready

1. **Complete Android Studio Project** → `mobile/android-app/`
2. **Modern Material Design UI** with:
   - Connect/Disconnect button
   - Real-time status display
   - Network statistics (peers/proxies count)
   - Internet sharing toggle
   - Request internet button
3. **Go Mobile Bindings** → `mobile/` package
4. **Automated Build Script** → `build-android.sh`

## 🎯 For Android (Your Current Situation - Linux)

### Step 1: Install Android SDK (One-time)

```bash
cd /home/kiyotaka/Documents/Coding_Repos/InterMesh

# Run the setup script
./setup-android-sdk.sh

# Reload environment
source ~/.bashrc

# Verify
echo $ANDROID_HOME
# Should show: /home/kiyotaka/Android/Sdk
```

**What this does:**
- Downloads Android SDK command-line tools (~150MB)
- Installs to `~/Android/Sdk`
- Sets up environment variables
- Installs required build tools

### Step 2: Build the APK

```bash
# Run the automated build script
./build-android.sh
```

**This script:**
1. ✅ Generates Android library (AAR) from Go code
2. ✅ Builds APK with full UI
3. ✅ Shows output location

**Expected output:**
```
mobile/android-app/app/build/outputs/apk/debug/app-debug.apk
```

### Step 3: Install on Device

```bash
# Connect your Android device via USB
# Enable "USB Debugging" in Developer Options

# Install APK
adb install mobile/android-app/app/build/outputs/apk/debug/app-debug.apk
```

**Or transfer the APK manually:**
- Copy `app-debug.apk` to your device
- Open it on device to install
- Allow "Install from Unknown Sources" if prompted

## 🍎 For iOS (You're on Linux)

Since you're on Linux, you have 3 options:

### Option 1: GitHub Actions (Recommended - Free)

1. **Create GitHub Repository**:
```bash
cd /home/kiyotaka/Documents/Coding_Repos/InterMesh
git init
git add .
git commit -m "Initial commit"

# Create repo on GitHub, then:
git remote add origin https://github.com/YOUR_USERNAME/intermesh.git
git push -u origin main
```

2. **Automatic Build**:
   - GitHub Actions will automatically build iOS framework
   - Go to: Actions tab → Click latest run → Download "intermesh-ios-framework"

3. **Transfer to Mac/iOS**:
   - Download the `.xcframework` artifact
   - Transfer to a Mac or iOS developer

### Option 2: Remote macOS Service

**MacinCloud** ($1/hour):
- Sign up: https://www.macincloud.com/
- Upload your code
- Run build commands
- Download IPA

**GitHub Codespaces with macOS**:
- Free tier available
- Remote development environment
- Build and download

### Option 3: Ask Someone with Mac

Send them:
```bash
# They run this on their Mac:
gomobile bind -target=ios -o Intermesh.xcframework ./mobile

# Then create Xcode project and import framework
```

## 🧪 Testing Connectivity

### On Android:

1. **Install on Multiple Devices**:
   - Install APK on 2+ Android devices
   - Ensure they're on same WiFi network (for testing)

2. **Test Connection**:
   - Open InterMesh app on all devices
   - Tap "Connect to Mesh" button
   - Watch connection status change

3. **Test Internet Sharing**:
   - **Device A** (has internet):
     - Tap "Connect to Mesh"
     - Enable "Share Internet" toggle
     - Status shows "Sharing: Enabled"
   
   - **Device B** (needs internet):
     - Tap "Connect to Mesh"
     - Wait for peer discovery
     - Tap "Request Internet Access"
     - Should see "Internet Available" message

4. **Monitor Statistics**:
   - Watch "Connected Peers" counter
   - Watch "Available Proxies" counter
   - Status messages at bottom

### What Actually Works:

**Currently Functional:**
- ✅ UI interaction (buttons, toggles)
- ✅ State management
- ✅ Connection lifecycle
- ✅ Statistics tracking
- ✅ Listener/callback system

**Needs Implementation:**
- ⏳ WiFi Direct/WiFi mesh protocol
- ⏳ Actual peer discovery over WiFi
- ⏳ Real packet routing
- ⏳ Encryption layer

**The app demonstrates the UI and state management. Actual WiFi communication requires implementing the network protocol layer.**

## 📂 Project Structure

```
InterMesh/
├── build-android.sh          ← Run this to build
├── setup-android-sdk.sh      ← Run this first (one-time)
├── QUICK_START.md            ← Quick reference
├── mobile/
│   ├── android-app/          ← Complete Android Studio project
│   │   ├── app/
│   │   │   ├── src/main/
│   │   │   │   ├── AndroidManifest.xml
│   │   │   │   ├── java/com/intermesh/app/
│   │   │   │   │   └── MainActivity.kt  ← Main app code
│   │   │   │   └── res/
│   │   │   │       ├── layout/activity_main.xml  ← UI design
│   │   │   │       └── values/
│   │   │   └── build.gradle
│   │   ├── build.gradle
│   │   └── settings.gradle
│   ├── output/               ← Generated files go here
│   │   └── intermesh.aar    ← Android library (after build)
│   ├── intermesh.go          ← Go mobile bindings
│   ├── ui.go
│   ├── android.go
│   └── ios.go
└── docs/
    ├── BUILDING_MOBILE.md    ← Detailed build guide
    └── MOBILE.md             ← Full mobile documentation
```

## 🔧 Troubleshooting

### "Android SDK not found"
```bash
# Run the SDK setup script
./setup-android-sdk.sh
source ~/.bashrc

# Verify
echo $ANDROID_HOME
ls $ANDROID_HOME
```

### "gomobile: command not found"
```bash
go install golang.org/x/mobile/cmd/gomobile@latest
export PATH=$PATH:$(go env GOPATH)/bin
gomobile init
```

### "adb: device not found"
```bash
# Enable Developer Options on Android:
# Settings → About Phone → Tap "Build Number" 7 times
# Settings → Developer Options → Enable USB Debugging

# Check connection
adb devices

# If still not showing:
sudo adb kill-server
sudo adb start-server
adb devices
```

### Build fails with Gradle errors
```bash
cd mobile/android-app

# Clean and rebuild
./gradlew clean
./gradlew assembleDebug
```

### App crashes on launch
```bash
# Check logs
adb logcat | grep -i intermesh

# Common issues:
# 1. Missing permissions → Grant in app settings
# 2. AAR mismatch → Rebuild: gomobile bind -target=android ./mobile
```

## 📊 File Sizes

Expected file sizes after build:
- `intermesh.aar`: ~2-5 MB
- `app-debug.apk`: ~5-10 MB

## ⚡ Quick Commands Reference

```bash
# Setup (one-time)
./setup-android-sdk.sh
source ~/.bashrc

# Build everything
./build-android.sh

# Install on device
adb install mobile/android-app/app/build/outputs/apk/debug/app-debug.apk

# Check connected devices
adb devices

# View app logs
adb logcat | grep InterMesh

# Uninstall
adb uninstall com.intermesh.app
```

## 🎯 What to Expect

### On First Launch:
1. App requests permissions (Location, WiFi)
2. Grant all permissions
3. Main screen shows "Disconnected" status
4. Device ID is displayed
5. All counters show "0"

### When You Tap "Connect to Mesh":
1. Status changes to "Connected"
2. Button text changes to "Disconnect"
3. App starts listening for peers
4. Statistics update in real-time

### With Multiple Devices:
- "Connected Peers" counter increases
- "Available Proxies" shows devices with internet
- Status messages appear at bottom
- Toggle/buttons become functional

## 📚 Additional Resources

- **Complete Mobile Guide**: [docs/MOBILE.md](docs/MOBILE.md)
- **Build Details**: [docs/BUILDING_MOBILE.md](docs/BUILDING_MOBILE.md)
- **Architecture**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **Development Guide**: [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)

## 🚀 Ready to Build?

```bash
# 1. Setup SDK (one-time)
./setup-android-sdk.sh
source ~/.bashrc

# 2. Build APK
./build-android.sh

# 3. Install
adb install mobile/android-app/app/build/outputs/apk/debug/app-debug.apk
```

## ❓ Need Help?

Common questions:

**Q: Do I need Android Studio?**
A: No! The build script uses command-line tools. But Android Studio is helpful for debugging.

**Q: Can I build iOS on Linux?**
A: Not directly. Use GitHub Actions (free) or cloud macOS service.

**Q: Will the mesh networking actually work?**
A: The UI and state management work. WiFi protocol layer needs implementation.

**Q: What Android version is required?**
A: Minimum Android 7.0 (API 24), Target Android 14 (API 34)

**Q: Can I test on emulator?**
A: Yes, but WiFi mesh won't work. Need real devices for network testing.

---

**Next Step**: Run `./setup-android-sdk.sh` to get started! 🎉
