# Package Manager Setup Guide

This guide explains how to set up Homebrew, Linux packages (.deb, .rpm, .apk), AUR, and Chocolatey package distribution for fluxid.

## Homebrew Tap Setup

### 1. Create Homebrew Tap Repository

Create a new GitHub repository named `homebrew-tap` under your account:

```bash
# On GitHub, create: steinmann321/homebrew-tap
# Or via gh CLI:
gh repo create steinmann321/homebrew-tap --public --description "Homebrew tap for fluxid CLI"
```

### 2. Configure GitHub Token

GoReleaser needs a GitHub token with `repo` scope to push formula updates:

```bash
# Create a personal access token at: https://github.com/settings/tokens
# Scopes needed: repo (full control)

# Set as environment variable (add to ~/.zshrc or ~/.bashrc):
export GITHUB_TOKEN="your_github_token_here"
```

### 3. Release Process

When you run `goreleaser release`, it will automatically:
- Create/update the Homebrew formula in `steinmann321/homebrew-tap`
- Commit and push the formula to the tap repository
- Users can then install via: `brew tap steinmann321/tap && brew install fluxid`

### 4. First Time Setup

After the first release with GoReleaser:

```bash
# Initialize the tap repository structure
cd /tmp
git clone https://github.com/steinmann321/homebrew-tap.git
cd homebrew-tap
mkdir -p Formula
# GoReleaser will create Formula/fluxid.rb automatically on first release
```

## Linux Packages Setup (.deb, .rpm, .apk)

### Overview

GoReleaser automatically creates native Linux packages through the `nfpms` configuration:
- **.deb** packages for Debian, Ubuntu, and derivatives
- **.rpm** packages for RHEL, Fedora, CentOS, and derivatives
- **.apk** packages for Alpine Linux

### 1. No Additional Setup Required!

Unlike Homebrew and Chocolatey, Linux packages don't require any special repositories or API keys. GoReleaser will automatically:
- Build .deb, .rpm, and .apk packages for both amd64 and arm64
- Include them in the GitHub release
- Generate proper package metadata

### 2. Package Contents

Each package includes:
- Binary installed to `/usr/bin/fluxid`
- License file at `/usr/share/doc/fluxid/copyright`
- Proper package metadata (version, maintainer, description)

### 3. Installation Methods

**Debian/Ubuntu**:
```bash
# Download from GitHub releases
wget https://github.com/steinmann321/fluxid-cli/releases/download/v0.1.0/fluxid_0.1.0_linux_amd64.deb

# Install
sudo dpkg -i fluxid_0.1.0_linux_amd64.deb

# Fix dependencies if needed
sudo apt-get install -f
```

**Fedora/RHEL/CentOS**:
```bash
# Download from GitHub releases
wget https://github.com/steinmann321/fluxid-cli/releases/download/v0.1.0/fluxid_0.1.0_linux_amd64.rpm

# Install with rpm
sudo rpm -i fluxid_0.1.0_linux_amd64.rpm

# Or with dnf
sudo dnf install fluxid_0.1.0_linux_amd64.rpm
```

**Alpine Linux**:
```bash
# Download from GitHub releases
wget https://github.com/steinmann321/fluxid-cli/releases/download/v0.1.0/fluxid_0.1.0_linux_amd64.apk

# Install
apk add --allow-untrusted fluxid_0.1.0_linux_amd64.apk
```

### 4. Package Repository (Optional Advanced Setup)

For easier installation, you can set up a custom APT/YUM repository:

**APT Repository (Debian/Ubuntu)**:
```bash
# 1. Create a GitHub release with .deb packages
# 2. Users add your repository:
echo "deb [trusted=yes] https://github.com/steinmann321/fluxid-cli/releases/latest/download/ ./" | sudo tee /etc/apt/sources.list.d/fluxid.list
sudo apt update
sudo apt install fluxid
```

**YUM/DNF Repository (RHEL/Fedora)**:
```bash
# Similar setup for RPM-based distributions
# Create a .repo file pointing to GitHub releases
```

## AUR (Arch User Repository) Setup

### 1. Create AUR Account

1. Sign up at https://aur.archlinux.org/register
2. Add your SSH public key to your AUR account

### 2. Generate AUR SSH Key

```bash
# Generate a dedicated SSH key for AUR
ssh-keygen -t ed25519 -C "fluxid-aur" -f ~/.ssh/aur

# Add to ssh-agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/aur

# Copy public key to AUR account settings
cat ~/.ssh/aur.pub
```

### 3. Configure AUR Key in Environment

```bash
# Export the private key as base64
export AUR_KEY=$(cat ~/.ssh/aur | base64)

# Add to your shell profile (~/.zshrc or ~/.bashrc):
export AUR_KEY="$(cat ~/.ssh/aur | base64)"
```

### 4. Create AUR Package Repository

```bash
# Clone the AUR package repository (will be empty initially)
git clone ssh://aur@aur.archlinux.org/fluxid-bin.git
cd fluxid-bin

# GoReleaser will populate this on first release
# Initial commit
git commit --allow-empty -m "Initial commit"
git push -u origin master
```

### 5. Release Process

When you run `goreleaser release`, it will automatically:
- Generate the PKGBUILD file
- Calculate checksums
- Commit and push to AUR repository
- Users can then install via AUR helpers: `yay -S fluxid-bin`

### 6. Testing AUR Package

```bash
# Test local PKGBUILD
cd /tmp
git clone https://aur.archlinux.org/fluxid-bin.git
cd fluxid-bin
makepkg -si

# Or with AUR helpers
yay -S fluxid-bin
paru -S fluxid-bin
```

## Chocolatey Package Setup

### 1. Create Chocolatey Account

1. Sign up at https://community.chocolatey.org/account/Register
2. Get your API key from https://community.chocolatey.org/account

### 2. Configure Chocolatey API Key

```bash
# Set as environment variable:
export CHOCOLATEY_API_KEY="your_chocolatey_api_key_here"
```

### 3. Package Icon (Optional)

Create an icon and add it to your repository:

```bash
# Add icon to: assets/icon.png (256x256 or 128x128)
# Commit and push
git add assets/icon.png
git commit -m "Add Chocolatey package icon"
git push
```

### 4. Release Process

When you run `goreleaser release`, it will automatically:
- Create the Chocolatey package (.nupkg)
- Publish to chocolatey.org (if `skip_publish: false`)
- Users can then install via: `choco install fluxid`

### 5. First Package Submission

The first time a package is published to Chocolatey, it goes through moderation:
- Review typically takes 1-3 days
- Subsequent updates are automatic
- Monitor status at: https://community.chocolatey.org/packages/fluxid

## Testing Locally

### Test Homebrew Formula

```bash
# After creating the tap
brew tap steinmann321/tap
brew install fluxid --verbose --debug

# Test the installation
fluxid version

# Uninstall
brew uninstall fluxid
brew untap steinmann321/tap
```

### Test Chocolatey Package

```powershell
# Install from local .nupkg (Windows)
choco install fluxid -s . -y

# Test the installation
fluxid version

# Uninstall
choco uninstall fluxid
```

## Release Workflow

### Standard Release with Package Managers

```bash
# 1. Ensure environment variables are set
echo $GITHUB_TOKEN
echo $CHOCOLATEY_API_KEY

# 2. Tag and push
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0

# 3. Run GoReleaser (will create binaries + update package managers)
goreleaser release --clean

# 4. Verify
# - Check GitHub release: https://github.com/steinmann321/fluxid-cli/releases
# - Check Homebrew tap: https://github.com/steinmann321/homebrew-tap
# - Check Chocolatey: https://community.chocolatey.org/packages/fluxid
```

### Skip Package Manager Publishing

If you want to create the packages but not publish them:

```bash
# Homebrew: Remove GITHUB_TOKEN from environment
# Chocolatey: Set skip_publish: true in .goreleaser.yml

goreleaser release --clean --skip=publish
```

## Troubleshooting

### Homebrew

**Issue**: Formula not updating in tap
- Check GitHub token has `repo` scope
- Verify repository name matches: `steinmann321/homebrew-tap`
- Check GoReleaser output for errors

**Issue**: Formula installation fails
- Test formula syntax: `brew audit --strict Formula/fluxid.rb`
- Check binary paths and architecture support

### Chocolatey

**Issue**: Package not publishing
- Verify API key is correct
- Check package is approved (first submission needs approval)
- Monitor at: https://community.chocolatey.org/packages/fluxid

**Issue**: Package validation fails
- Check .nuspec metadata is valid
- Verify all URLs are accessible
- Review Chocolatey moderation feedback

## CI/CD Integration

To automate releases in CI:

```yaml
# GitHub Actions example
- name: Release
  env:
    GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
    CHOCOLATEY_API_KEY: ${{ secrets.CHOCO_API_KEY }}
  run: goreleaser release --clean
```

## References

- GoReleaser Homebrew: https://goreleaser.com/customization/homebrew/
- GoReleaser Chocolatey: https://goreleaser.com/customization/chocolatey/
- Homebrew Formula Cookbook: https://docs.brew.sh/Formula-Cookbook
- Chocolatey Package Creation: https://docs.chocolatey.org/en-us/create/create-packages
