# InterMesh iOS Build Script (requires macOS)

Write-Host "🚀 InterMesh iOS Build Script" -ForegroundColor Cyan
Write-Host "==============================`n" -ForegroundColor Cyan

# Check OS
if ($IsMacOS -eq $false) {
    Write-Host "❌ This script requires macOS to build iOS apps" -ForegroundColor Red
    Write-Host "`nAlternatives for Linux/Windows:" -ForegroundColor Yellow
    Write-Host "1. Use GitHub Actions (see .github/workflows/ios-build.yml)"
    Write-Host "2. Use a cloud macOS service (MacStadium, MacinCloud)"
    Write-Host "3. Transfer the .xcframework to a Mac for building"
    exit 1
}

# Check for Go
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Go is not installed" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Go installed: $(go version)" -ForegroundColor Green

# Check for gomobile
if (-not (Get-Command gomobile -ErrorAction SilentlyContinue)) {
    Write-Host "⚠ gomobile not found, installing..." -ForegroundColor Yellow
    go install golang.org/x/mobile/cmd/gomobile@latest
    $env:PATH += ":$(go env GOPATH)/bin"
    gomobile init
}
Write-Host "✓ gomobile installed`n" -ForegroundColor Green

# Create output directory
Write-Host "Creating output directory..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path mobile/output | Out-Null
Write-Host "✓ Output directory ready`n" -ForegroundColor Green

# Generate iOS framework
Write-Host "Generating iOS framework (XCFramework)..." -ForegroundColor Yellow
Write-Host "This may take a few minutes...`n"

gomobile bind -target=ios -o mobile/output/Intermesh.xcframework ./mobile

if (Test-Path "mobile/output/Intermesh.xcframework") {
    Write-Host "✓ XCFramework generated successfully`n" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to generate XCFramework" -ForegroundColor Red
    exit 1
}

Write-Host "✅ iOS Framework Ready!" -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "1. Create a new iOS project in Xcode"
Write-Host "2. Drag Intermesh.xcframework into your project"
Write-Host "3. Add WiFi permissions to Info.plist"
Write-Host "4. Import the framework: import Intermesh"
Write-Host "5. Use MobileApp to connect to mesh network"
Write-Host "`nSee docs/MOBILE.md for complete integration guide"
