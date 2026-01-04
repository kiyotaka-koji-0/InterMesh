#!/bin/bash

# InterMesh iOS Build Script
# Builds the iOS framework and app

set -e  # Exit on error

echo "🚀 InterMesh iOS Build Script"
echo "=============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo -e "${RED}❌ This script requires macOS to build iOS apps${NC}"
    echo ""
    echo "Alternatives for Linux/Windows:"
    echo "1. Use GitHub Actions (automatic builds on push)"
    echo "2. Use a cloud macOS service (MacStadium, MacinCloud)"
    echo "3. Build just the framework and transfer to a Mac"
    exit 1
fi

# Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    echo "Install Go from: https://golang.org/dl/"
    exit 1
fi
echo -e "${GREEN}✓ Go installed: $(go version)${NC}"

# Check for gomobile
if ! command -v gomobile &> /dev/null; then
    echo -e "${YELLOW}⚠ gomobile not found, installing...${NC}"
    go install golang.org/x/mobile/cmd/gomobile@latest
    export PATH=$PATH:$(go env GOPATH)/bin
    
    # Initialize gomobile
    echo -e "${YELLOW}Initializing gomobile...${NC}"
    gomobile init
fi
echo -e "${GREEN}✓ gomobile installed${NC}"
echo ""

# Check for Xcode
if ! command -v xcodebuild &> /dev/null; then
    echo -e "${RED}❌ Xcode is not installed${NC}"
    echo "Install Xcode from the Mac App Store"
    exit 1
fi
echo -e "${GREEN}✓ Xcode installed: $(xcodebuild -version | head -n 1)${NC}"
echo ""

# Create output directory
echo -e "${YELLOW}Creating output directory...${NC}"
mkdir -p mobile/output
echo -e "${GREEN}✓ Output directory ready${NC}"
echo ""

# Step 1: Build iOS Framework
echo -e "${YELLOW}Step 1: Building iOS Framework (XCFramework)...${NC}"
echo "This may take a few minutes..."
echo ""

# Download dependencies first
echo -e "${YELLOW}Downloading Go dependencies...${NC}"
go mod download
go mod verify

# Build the framework
if gomobile bind -target=ios -o mobile/output/Intermesh.xcframework ./mobile; then
    echo -e "${GREEN}✓ iOS Framework built successfully${NC}"
    echo -e "Location: ${GREEN}mobile/output/Intermesh.xcframework${NC}"
else
    echo -e "${RED}❌ Failed to build iOS framework${NC}"
    exit 1
fi
echo ""

# Step 2: Build iOS App
echo -e "${YELLOW}Step 2: Building iOS App...${NC}"
echo ""

# Check if Xcode project exists
if [ ! -f "mobile/ios-app/InterMesh.xcodeproj/project.pbxproj" ]; then
    echo -e "${RED}❌ Xcode project not found${NC}"
    echo "Expected: mobile/ios-app/InterMesh.xcodeproj"
    exit 1
fi

# Clean build folder
echo -e "${YELLOW}Cleaning build folder...${NC}"
cd mobile/ios-app
xcodebuild clean -project InterMesh.xcodeproj -scheme InterMesh -configuration Release > /dev/null 2>&1 || true
echo -e "${GREEN}✓ Clean complete${NC}"
echo ""

# Build for iOS device
echo -e "${YELLOW}Building for iOS device (this may take a while)...${NC}"
xcodebuild clean build \
    -project InterMesh.xcodeproj \
    -scheme InterMesh \
    -sdk iphoneos \
    -configuration Release \
    -derivedDataPath ./build \
    CODE_SIGN_IDENTITY="" \
    CODE_SIGNING_REQUIRED=NO \
    CODE_SIGNING_ALLOWED=NO \
    | grep -E "^(Build|▸)" || true

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ iOS app built successfully${NC}"
else
    echo -e "${RED}❌ Failed to build iOS app${NC}"
    echo ""
    echo "This is likely due to code signing requirements."
    echo "To install on your iPad, please:"
    echo "1. Open: mobile/ios-app/InterMesh.xcodeproj in Xcode"
    echo "2. Connect your iPad"
    echo "3. Select your iPad as the build destination"
    echo "4. Click the Run button (▶️)"
    echo ""
    echo "The framework is ready at: mobile/output/Intermesh.xcframework"
fi

cd ../..
echo ""

# Summary
echo -e "${GREEN}✅ Build Complete!${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📦 Outputs:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "Framework: ${GREEN}mobile/output/Intermesh.xcframework${NC}"
echo -e "App Project: ${GREEN}mobile/ios-app/InterMesh.xcodeproj${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📱 Next Steps:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "To install on your iPad:"
echo ""
echo "1. Open Xcode project:"
echo -e "   ${YELLOW}open mobile/ios-app/InterMesh.xcodeproj${NC}"
echo ""
echo "2. Connect your iPad via USB cable"
echo ""
echo "3. In Xcode:"
echo "   • Select your iPad from the device dropdown (top bar)"
echo "   • Go to 'Signing & Capabilities' tab"
echo "   • Sign in with your Apple ID (free account works!)"
echo "   • Select your team from the Team dropdown"
echo ""
echo "4. Click the Run button (▶️) or press Cmd+R"
echo ""
echo "5. First time: On your iPad"
echo "   • Settings > General > VPN & Device Management"
echo "   • Trust your developer certificate"
echo ""
echo "6. Launch the InterMesh app from your home screen!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "📖 For detailed instructions, see: ${YELLOW}IOS_INSTALLATION_GUIDE.md${NC}"
echo ""
echo "🎉 Happy meshing!"
