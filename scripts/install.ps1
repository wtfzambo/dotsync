#!/usr/bin/env powershell
#
# dotsync installation script for Windows
# Usage: powershell -Command "iwr -useb https://raw.githubusercontent.com/wtfzambo/dotsync/main/scripts/install.ps1 | iex"
#

$ErrorActionPreference = "Stop"

$REPO = "wtfzambo/dotsync"
$BINARY_NAME = "dotsync.exe"

function Write-Info {
    param($Message)
    Write-Host "==> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param($Message)
    Write-Host "==> $Message" -ForegroundColor Green
}

function Write-Warning {
    param($Message)
    Write-Host "==> $Message" -ForegroundColor Yellow
}

function Write-Error {
    param($Message)
    Write-Host "Error: $Message" -ForegroundColor Red
}

# Detect platform
function Get-Platform {
    $arch = switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { "x86_64" }
        "ARM64" { "arm64" }
        default { 
            Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
            exit 1
        }
    }
    return "Windows_$arch"
}

# Download and install from GitHub releases
function Install-FromRelease {
    param($Platform)
    
    Write-Info "Installing $BINARY_NAME from GitHub releases..."
    
    $tempDir = Join-Path $env:TEMP "dotsync-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir | Out-Null
    
    try {
        # Get latest release version
        Write-Info "Fetching latest release..."
        $latestUrl = "https://api.github.com/repos/$REPO/releases/latest"
        
        try {
            $release = Invoke-RestMethod -Uri $latestUrl -Headers @{"Accept" = "application/vnd.github.v3+json"}
        } catch {
            Write-Warning "Failed to fetch latest version"
            return $false
        }
        
        $version = $release.tag_name -replace '^v', ''
        Write-Info "Latest version: $version"
        
        # Construct download URL
        $archiveName = "dotsync_${version}_${Platform}.zip"
        $downloadUrl = "https://github.com/$REPO/releases/download/v$version/$archiveName"
        
        Write-Info "Downloading $archiveName..."
        $archivePath = Join-Path $tempDir $archiveName
        
        try {
            Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing
        } catch {
            Write-Warning "Download failed, trying fallback installation method..."
            return $false
        }
        
        # Extract archive
        Write-Info "Extracting archive..."
        try {
            # Try using tar (available in Windows 10+)
            & tar -xzf $archivePath -C $tempDir 2>&1 | Out-Null
            if ($LASTEXITCODE -ne 0) {
                throw "tar extraction failed"
            }
        } catch {
            Write-Warning "Failed to extract archive"
            return $false
        }
        
        # Determine install location
        $installDir = Join-Path $env:USERPROFILE ".local\bin"
        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }
        
        # Install binary
        Write-Info "Installing to $installDir..."
        $binaryPath = Join-Path $tempDir "dotsync.exe"
        if (-not (Test-Path $binaryPath)) {
            # Try without .exe extension (in case archive contains binary without extension)
            $binaryPath = Join-Path $tempDir "dotsync"
            if (-not (Test-Path $binaryPath)) {
                # Look for any executable in temp dir
                $binaryPath = Get-ChildItem -Path $tempDir -Filter "*.exe" | Select-Object -First 1 -ExpandProperty FullName
                if (-not $binaryPath) {
                    $binaryPath = Get-ChildItem -Path $tempDir -File | Where-Object { $_.Name -eq "dotsync" -or $_.Name -eq "dotsync.exe" } | Select-Object -First 1 -ExpandProperty FullName
                }
            }
        }
        
        if (-not $binaryPath -or -not (Test-Path $binaryPath)) {
            Write-Warning "Could not find binary in archive"
            return $false
        }
        
        $destPath = Join-Path $installDir $BINARY_NAME
        Copy-Item -Path $binaryPath -Destination $destPath -Force
        
        Write-Success "$BINARY_NAME installed to $destPath"
        
        # Check if install_dir is in PATH
        $userPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
        if (-not ($userPath -like "*$installDir*")) {
            Write-Warning "$installDir is not in your PATH"
            Write-Host ""
            Write-Host "Adding to PATH..."
            [System.Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", [System.EnvironmentVariableTarget]::User)
            Write-Success "Added $installDir to PATH (restart your terminal)"
            Write-Host ""
        }
        
        return $true
    } finally {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Check if Go is installed and meets minimum version
function Test-Go {
    try {
        $goVersion = & go version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Info "Go detected: $goVersion"
            
            # Extract version number
            if ($goVersion -match "go(\d+)\.(\d+)") {
                $major = [int]$matches[1]
                $minor = [int]$matches[2]
                
                if ($major -eq 1 -and $minor -lt 25) {
                    Write-Error "Go 1.25 or later is required (found: $major.$minor)"
                    Write-Host ""
                    Write-Host "Please upgrade Go:"
                    Write-Host "  - Download from https://go.dev/dl/"
                    Write-Host ""
                    return $false
                }
            }
            return $true
        }
    } catch {
        return $false
    }
    return $false
}

# Install using go install (fallback)
function Install-WithGo {
    Write-Info "Installing $BINARY_NAME using 'go install'..."
    
    try {
        & go install "github.com/$REPO/cmd/dotsync@latest" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "$BINARY_NAME installed successfully via go install"
            
            # Determine where Go installed the binary
            $goBin = & go env GOBIN 2>$null
            if (-not $goBin) {
                $goPath = & go env GOPATH
                $binDir = Join-Path $goPath "bin"
            } else {
                $binDir = $goBin
            }
            
            # Check if GOPATH/bin (or GOBIN) is in PATH
            $userPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
            if (-not ($userPath -like "*$binDir*")) {
                Write-Warning "$binDir is not in your PATH"
                Write-Host ""
                Write-Host "Adding to PATH..."
                [System.Environment]::SetEnvironmentVariable("Path", "$userPath;$binDir", [System.EnvironmentVariableTarget]::User)
                Write-Success "Added $binDir to PATH (restart your terminal)"
                Write-Host ""
            }
            
            return $true
        }
    } catch {
        Write-Error "go install failed: $_"
    }
    return $false
}

# Verify installation
function Test-Installation {
    # Refresh PATH in current session
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::Machine) + ";" + [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
    
    $binary = Get-Command "dotsync" -ErrorAction SilentlyContinue
    if ($binary) {
        Write-Success "$BINARY_NAME is installed and ready!"
        Write-Host ""
        & dotsync --version
        Write-Host ""
        Write-Host "Get started:"
        Write-Host "  dotsync --help"
        Write-Host ""
        return $true
    } else {
        Write-Error "$BINARY_NAME was installed but is not in PATH"
        Write-Host ""
        Write-Host "Please restart your terminal and run 'dotsync --version' to verify"
        Write-Host ""
        return $false
    }
}

# Main installation flow
function Main {
    Write-Host ""
    Write-Host " dotsync Installer (Windows)" -ForegroundColor Blue
    Write-Host ""
    
    # Check if running on Windows
    if (-not $IsWindows -and $env:OS -ne "Windows_NT") {
        Write-Error "This installer is for Windows only. Use install.sh for Linux/macOS."
        exit 1
    }
    
    Write-Info "Detecting platform..."
    $platform = Get-Platform
    Write-Info "Platform: $platform"
    
    # Try downloading from GitHub releases first
    if (Install-FromRelease -Platform $platform) {
        $null = Test-Installation
        exit 0
    }
    
    Write-Warning "Failed to install from releases, trying fallback method..."
    
    # Try go install as fallback
    if (Test-Go) {
        if (Install-WithGo) {
            $null = Test-Installation
            exit 0
        }
    }
    
    # All methods failed
    Write-Error "Installation failed"
    Write-Host ""
    Write-Host "Manual installation:"
    Write-Host "  1. Download from https://github.com/$REPO/releases/latest"
    Write-Host "  2. Extract and move 'dotsync.exe' to a directory in your PATH"
    Write-Host ""
    Write-Host "Or install from source:"
    Write-Host "  1. Install Go 1.25+ from https://go.dev/dl/"
    Write-Host "  2. Run: go install github.com/$REPO/cmd/dotsync@latest"
    Write-Host ""
    exit 1
}

Main
