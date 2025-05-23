# PowerShell script to set up and run the Scheduling API application

# Output formatting
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    else {
        $input | Write-Output
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

Write-ColorOutput Green "Scheduling API Setup and Run Script"
Write-Output "========================================"

# Check if Go is installed
$goCommand = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCommand) {
    Write-ColorOutput Red "Error: Go is not installed or not in PATH"
    Write-Output "Please install Go from https://golang.org/dl/"
    exit 1
}

# Check Go version
$goVersion = (go version) -replace "go version ", ""
Write-ColorOutput Yellow "Go version: $goVersion"

# Check if .env file exists
if (-not (Test-Path .env)) {
    Write-ColorOutput Yellow "Creating default .env file..."
    @"
# Server settings
SERVER_ADDRESS=:8080
GIN_MODE=debug

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=scheduling_db
DB_SSLMODE=disable

# JWT settings
JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRE_HOURS=24

# CORS settings
CORS_ALLOWED_ORIGINS=*
"@ | Out-File -FilePath .env -Encoding utf8

    Write-ColorOutput Green ".env file created successfully"
    Write-ColorOutput Yellow "Please edit .env file with your database credentials"
}

# Download dependencies
Write-ColorOutput Yellow "Downloading dependencies..."
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-ColorOutput Red "Failed to download dependencies"
    exit 1
}

# Build the application
Write-ColorOutput Yellow "Building application..."
go build -o scheduling-api.exe .\cmd\api
if ($LASTEXITCODE -ne 0) {
    Write-ColorOutput Red "Build failed"
    exit 1
}

Write-ColorOutput Green "Build successful!"
Write-ColorOutput Yellow "Starting application..."

# Run the application
.\scheduling-api.exe

# Exit with the exit code of the application
exit $LASTEXITCODE

