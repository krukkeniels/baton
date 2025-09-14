# Baton CLI Orchestrator Windows Installation Script
param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\baton"
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Status {
    param([string]$Message)
    Write-Host "==> $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
    exit 1
}

# Main installation function
function Install-Baton {
    Write-Host "🚀 Baton CLI Orchestrator Installation" -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host ""

    # Detect architecture
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $platform = "windows-$arch"
    $binaryName = "baton.exe"

    Write-Status "Detected platform: $platform"

    # Create installation directory
    if (!(Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
        Write-Success "Created installation directory: $InstallDir"
    }

    # Download URL
    $githubRepo = "race-day/baton"
    if ($Version -eq "latest") {
        $downloadUrl = "https://github.com/$githubRepo/releases/latest/download/baton-$platform.exe"
    } else {
        $downloadUrl = "https://github.com/$githubRepo/releases/download/$Version/baton-$platform.exe"
    }

    $binaryPath = Join-Path $InstallDir $binaryName

    Write-Status "Downloading from: $downloadUrl"

    try {
        # Download binary
        Invoke-WebRequest -Uri $downloadUrl -OutFile $binaryPath -UseBasicParsing
        Write-Success "Downloaded baton binary"
    } catch {
        Write-Error-Custom "Failed to download baton binary: $($_.Exception.Message)"
    }

    # Verify binary
    try {
        $output = & $binaryPath --version 2>&1
        Write-Success "Binary verification passed"
    } catch {
        Write-Warning "Binary verification failed, but continuing with installation"
    }

    # Add to PATH
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        Write-Status "Adding $InstallDir to user PATH"
        $newPath = "$userPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Success "Added to PATH"
        Write-Warning "Please restart your terminal to use baton"
    } else {
        Write-Success "Installation directory already in PATH"
    }

    # Create desktop shortcut (optional)
    $createShortcut = Read-Host "Create desktop shortcut? [y/N]"
    if ($createShortcut -eq "y" -or $createShortcut -eq "Y") {
        $desktopPath = [Environment]::GetFolderPath("Desktop")
        $shortcutPath = Join-Path $desktopPath "Baton CLI.lnk"

        $WshShell = New-Object -comObject WScript.Shell
        $shortcut = $WshShell.CreateShortcut($shortcutPath)
        $shortcut.TargetPath = "powershell.exe"
        $shortcut.Arguments = "-NoExit -Command `"& '$binaryPath'`""
        $shortcut.WorkingDirectory = "$env:USERPROFILE"
        $shortcut.IconLocation = "$binaryPath,0"
        $shortcut.Description = "Baton CLI Orchestrator"
        $shortcut.Save()

        Write-Success "Created desktop shortcut"
    }

    Write-Host ""
    Write-Host "🎉 Installation completed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "1. Restart your terminal or refresh PATH"
    Write-Host "2. Verify installation: baton --version"
    Write-Host "3. Get started: baton init"
    Write-Host "4. Read the docs: baton --help"
    Write-Host ""
    Write-Host "Requirements for full functionality:"
    Write-Host "• Claude Code CLI (for LLM integration)"
    Write-Host "• Git for Windows (for version control features)"
    Write-Host ""

    # Try to run baton --version
    try {
        $env:PATH = "$env:PATH;$InstallDir"
        $versionOutput = & $binaryPath --version 2>&1
        Write-Host "✅ Installation verified: $versionOutput" -ForegroundColor Green
    } catch {
        Write-Host "⚠️  Please restart your terminal to use baton" -ForegroundColor Yellow
    }
}

# Check if running as administrator (optional, for system-wide install)
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Handle command line arguments
if ($args.Length -gt 0) {
    switch ($args[0]) {
        "--help" {
            Write-Host "Baton CLI Orchestrator Windows Installation"
            Write-Host ""
            Write-Host "Usage: powershell -ExecutionPolicy Bypass -File install.ps1 [options]"
            Write-Host ""
            Write-Host "Options:"
            Write-Host "  --version VERSION    Install specific version (default: latest)"
            Write-Host "  --dir DIR           Install to specific directory"
            Write-Host "  --help              Show this help message"
            Write-Host ""
            Write-Host "Examples:"
            Write-Host "  # Install latest version"
            Write-Host "  powershell -c `"irm https://raw.githubusercontent.com/race-day/baton/main/install.ps1 | iex`""
            Write-Host ""
            Write-Host "  # Install specific version"
            Write-Host "  powershell -ExecutionPolicy Bypass -File install.ps1 --version v1.0.0"
            exit 0
        }
        "--version" {
            if ($args.Length -gt 1) {
                $Version = $args[1]
            }
        }
        "--dir" {
            if ($args.Length -gt 1) {
                $InstallDir = $args[1]
            }
        }
    }
}

# Run installation
Install-Baton