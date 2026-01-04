#!/bin/bash
# InterMesh Android Build Script

set -e

echo "🚀 InterMesh Android Build Script"
echo "=================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for required tools
echo -e "\n${YELLOW}Step 1: Checking requirements...${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go installed: $(go version)${NC}"

# Check gomobile
GOPATH_BIN=$(go env GOPATH)/bin
export PATH=$PATH:$GOPATH_BIN
if ! command -v gomobile &> /dev/null; then
    echo -e "${YELLOW}⚠ gomobile not found, installing...${NC}"
    go install golang.org/x/mobile/cmd/gomobile@latest
    $GOPATH_BIN/gomobile init
fi
echo -e "${GREEN}✓ gomobile installed${NC}"

# Create output directory
echo -e "\n${YELLOW}Step 2: Creating output directory...${NC}"
mkdir -p mobile/output
echo -e "${GREEN}✓ Output directory ready${NC}"

# Generate Android AAR
echo -e "\n${YELLOW}Step 3: Generating Android library (AAR)...${NC}"
echo "This may take a few minutes..."
gomobile bind -target=android -o mobile/output/intermesh.aar ./mobile

if [ -f "mobile/output/intermesh.aar" ]; then
    AAR_SIZE=$(du -h mobile/output/intermesh.aar | cut -f1)
    echo -e "${GREEN}✓ AAR generated successfully: ${AAR_SIZE}${NC}"
else
    echo -e "${RED}❌ Failed to generate AAR${NC}"
    exit 1
fi

# Check for Android SDK
echo -e "\n${YELLOW}Step 4: Checking Android SDK...${NC}"

if [ -z "$ANDROID_HOME" ]; then
    echo -e "${RED}❌ ANDROID_HOME not set${NC}"
    echo "Please set ANDROID_HOME environment variable"
    echo "Example: export ANDROID_HOME=$HOME/Android/Sdk"
    exit 1
fi
echo -e "${GREEN}✓ ANDROID_HOME: $ANDROID_HOME${NC}"

# Build APK
echo -e "\n${YELLOW}Step 5: Building Android APK...${NC}"

cd mobile/android-app

# Check for gradlew
if [ ! -f "gradlew" ]; then
    echo -e "${YELLOW}⚠ Gradle wrapper not found, downloading...${NC}"
    gradle wrapper
    chmod +x gradlew
fi

echo "Building debug APK..."
./gradlew assembleDebug

APK_PATH="app/build/outputs/apk/debug/app-debug.apk"

if [ -f "$APK_PATH" ]; then
    APK_SIZE=$(du -h "$APK_PATH" | cut -f1)
    echo -e "${GREEN}✓ APK built successfully: ${APK_SIZE}${NC}"
    echo -e "\n${GREEN}📱 APK Location:${NC}"
    echo "  $(pwd)/$APK_PATH"
else
    echo -e "${RED}❌ Failed to build APK${NC}"
    exit 1
fi

cd ../..

# Installation instructions
echo -e "\n${GREEN}✅ Build Complete!${NC}"
echo -e "\n${YELLOW}To install on your Android device:${NC}"
echo "1. Enable USB debugging on your device"
echo "2. Connect via USB"
echo "3. Run: adb install mobile/android-app/app/build/outputs/apk/debug/app-debug.apk"
echo ""
echo -e "${YELLOW}Or for wireless install:${NC}"
echo "1. Ensure device is on same WiFi network"
echo "2. Run: adb connect <device-ip>:5555"
echo "3. Run: adb install mobile/android-app/app/build/outputs/apk/debug/app-debug.apk"
echo ""
echo -e "${GREEN}🎉 Happy Testing!${NC}"
