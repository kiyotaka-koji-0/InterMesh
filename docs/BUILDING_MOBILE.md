# Building Mobile Apps for InterMesh

## Prerequisites

```bash
# Install gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# For Android (on Linux)
sudo apt-get install openjdk-11-jdk
export ANDROID_HOME=$HOME/Android/Sdk
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools

# Install Android SDK
# Download from: https://developer.android.com/studio#command-tools
```

## Building for Android

### Step 1: Generate Android Library (AAR)

```bash
cd /home/kiyotaka/Documents/Coding_Repos/InterMesh

# Generate AAR file
gomobile bind -target=android -o mobile/output/intermesh.aar ./mobile

# This creates:
# - intermesh.aar (Android library)
# - intermesh-sources.jar (source files)
```

### Step 2: Create Android App

See `mobile/android-app/` directory for complete Android Studio project.

### Step 3: Build APK

**Option A: Using Android Studio**
1. Open `mobile/android-app/` in Android Studio
2. Build → Build Bundle(s) / APK(s) → Build APK(s)
3. APK will be in `app/build/outputs/apk/`

**Option B: Using Command Line**
```bash
cd mobile/android-app
./gradlew assembleDebug

# Output: app/build/outputs/apk/debug/app-debug.apk
```

### Step 4: Install APK on Device

```bash
# Via USB
adb install app/build/outputs/apk/debug/app-debug.apk

# Via wireless (if device on same network)
adb connect <device-ip>:5555
adb install app/build/outputs/apk/debug/app-debug.apk
```

## Building for iOS (on Linux)

⚠️ **Important**: Building iOS apps normally requires macOS and Xcode. However, here are options:

### Option 1: Cross-compilation (Limited)

```bash
# Generate iOS framework
gomobile bind -target=ios -o mobile/output/Intermesh.xcframework ./mobile

# This creates the framework but you still need macOS to:
# - Sign the app
# - Package as IPA
# - Install on device
```

**Workaround**: Transfer the `.xcframework` to a Mac or use a CI/CD service.

### Option 2: Use Darling (macOS on Linux) - Experimental

```bash
# Install Darling (macOS compatibility layer)
# https://www.darlinghq.org/

# This is experimental and may not work perfectly
```

### Option 3: Use Remote macOS Build Service

**GitHub Actions** (Free for public repos):
```yaml
# .github/workflows/ios-build.yml
name: Build iOS
on: [push]
jobs:
  build:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build
        run: |
          gomobile init
          gomobile bind -target=ios ./mobile
```

**GitLab CI** with macOS runners:
```yaml
# .gitlab-ci.yml
ios-build:
  stage: build
  tags:
    - macos
  script:
    - gomobile bind -target=ios ./mobile
```

**Codemagic** (Free tier available):
- Sign up at codemagic.io
- Connect your repository
- Configure iOS build

### Option 4: Virtual macOS

Use a cloud macOS instance:
- **MacStadium** - Rent macOS VMs
- **MacinCloud** - Remote macOS access
- **AWS EC2 Mac instances** - macOS on AWS

## Testing Connectivity Without Full App

### Quick Test: Use `gomobile build`

This creates a standalone APK without needing Android Studio:

```bash
# Build standalone Android APK
gomobile build -target=android -o intermesh.apk ./cmd/mobile-demo

# Install directly
adb install intermesh.apk
```

However, this creates a basic app without custom UI.

## Recommended Approach for Linux

### For Android (Recommended ✅)

1. **Generate AAR library**:
```bash
gomobile bind -target=android -o intermesh.aar ./mobile
```

2. **Use provided Android Studio project** (see `mobile/android-app/`)

3. **Build APK**:
```bash
cd mobile/android-app
./gradlew assembleDebug
```

4. **Install on device**:
```bash
adb install app/build/outputs/apk/debug/app-debug.apk
```

### For iOS Testing (Workarounds)

**Option A**: Use GitHub Actions (Easiest)
- Push code to GitHub
- GitHub Actions builds iOS app automatically
- Download IPA from artifacts
- Install via Xcode or TestFlight

**Option B**: Remote Mac Access
- Rent temporary Mac access
- Build and sign remotely
- Install via TestFlight or direct install

**Option C**: Ask someone with Mac
- Share the `.xcframework`
- They build and share the IPA

## Simple Testing Setup

### 1. Build Android APK Now

```bash
# Quick build without Android Studio
cd /home/kiyotaka/Documents/Coding_Repos/InterMesh

# Build basic APK
gomobile build -target=android -o intermesh-test.apk ./cmd/mobile-demo

# Install
adb install intermesh-test.apk
```

### 2. Test Connectivity

The APK will include the mobile demo which:
- Connects to mesh network
- Detects internet connectivity
- Shows peer discovery
- Enables internet sharing
- Displays real-time statistics

### 3. Test Between Devices

Install on multiple Android devices and test:
- Peer discovery between devices
- Internet sharing from one device to another
- Proxy connection establishment
- Network statistics

## Troubleshooting

### Android SDK Not Found

```bash
# Download Android command-line tools
wget https://dl.google.com/android/repository/commandlinetools-linux-9477386_latest.zip
unzip commandlinetools-linux-9477386_latest.zip -d $HOME/android-sdk

# Set environment
export ANDROID_HOME=$HOME/android-sdk
export PATH=$PATH:$ANDROID_HOME/cmdline-tools/bin:$ANDROID_HOME/platform-tools

# Install required packages
sdkmanager "platform-tools" "platforms;android-30" "build-tools;30.0.3"
```

### gomobile: command not found

```bash
# Ensure Go bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Reinstall gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
```

### ADB not detecting device

```bash
# Enable USB debugging on Android device
# Settings → About Phone → Tap "Build number" 7 times
# Settings → Developer Options → Enable USB Debugging

# Check device
adb devices

# If not showing, try:
sudo adb kill-server
sudo adb start-server
adb devices
```

## Next Steps

1. Try the quick Android build above
2. Install on device and test
3. For iOS, set up GitHub Actions or use cloud Mac service
4. See `mobile/android-app/` for full Android Studio project

Need help with any specific step? Let me know!
