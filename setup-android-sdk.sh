#!/bin/bash
# Quick Android SDK Setup for InterMesh

set -e

echo "📦 Android SDK Setup for InterMesh"
echo "==================================="

# Detect OS
OS=$(uname -s)
ARCH=$(uname -m)

# Determine SDK URL
if [ "$OS" = "Linux" ]; then
    SDK_URL="https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip"
elif [ "$OS" = "Darwin" ]; then
    SDK_URL="https://dl.google.com/android/repository/commandlinetools-mac-11076708_latest.zip"
else
    echo "❌ Unsupported OS: $OS"
    exit 1
fi

# Set Android SDK location
ANDROID_SDK_ROOT="$HOME/Android/Sdk"
echo "Installing Android SDK to: $ANDROID_SDK_ROOT"

# Create directories
mkdir -p "$ANDROID_SDK_ROOT/cmdline-tools"
cd /tmp

# Download SDK tools
echo "📥 Downloading Android SDK command-line tools..."
wget -q --show-progress "$SDK_URL" -O cmdline-tools.zip

# Extract
echo "📂 Extracting..."
unzip -q cmdline-tools.zip
mv cmdline-tools "$ANDROID_SDK_ROOT/cmdline-tools/latest"

# Set environment
export ANDROID_HOME="$ANDROID_SDK_ROOT"
export PATH="$PATH:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools"

echo "export ANDROID_HOME=\"$ANDROID_SDK_ROOT\"" >> ~/.bashrc
echo "export PATH=\"\$PATH:\$ANDROID_HOME/cmdline-tools/latest/bin:\$ANDROID_HOME/platform-tools\"" >> ~/.bashrc

# Accept licenses
echo "📜 Accepting licenses..."
yes | sdkmanager --licenses > /dev/null 2>&1 || true

# Install required packages
echo "📦 Installing required SDK packages..."
sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0" "ndk;25.2.9519653"

echo ""
echo "✅ Android SDK installed successfully!"
echo ""
echo "Please run: source ~/.bashrc"
echo "Then try building again: ./build-android.sh"
