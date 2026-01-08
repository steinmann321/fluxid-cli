# Setup Instructions for v0.1.0 Release with Package Managers

## Prerequisites Checklist

Before running `goreleaser release`, complete these setup steps:

### 1. Create Homebrew Tap Repository

```bash
# Option A: Using GitHub web interface
# Go to: https://github.com/new
# Repository name: homebrew-tap
# Owner: steinmann321
# Public repository
# Click "Create repository"

# Option B: Using gh CLI (if installed)
gh repo create steinmann321/homebrew-tap --public \
  --description "Homebrew tap for fluxid CLI"
```

### 2. Configure GitHub Token

```bash
# Create token at: https://github.com/settings/tokens/new
# Required scopes: repo (full control)
# Token name: "GoReleaser for fluxid"

# Add to your shell profile (~/.zshrc or ~/.bashrc):
export GITHUB_TOKEN="ghp_your_token_here"

# Reload shell or run:
source ~/.zshrc  # or ~/.bashrc
```

### 3. Setup Chocolatey (Optional for Windows users)

```bash
# Sign up at: https://community.chocolatey.org/account/Register
# Get API key from: https://community.chocolatey.org/account

# Add to your shell profile:
export CHOCOLATEY_API_KEY="your_api_key_here"

# To skip Chocolatey publishing for now, edit .goreleaser.yml:
# Change: skip_publish: false
# To: skip_publish: true
```

### 4. Add Package Icon (Optional but Recommended)

```bash
# Create assets directory
mkdir -p assets

# Add your icon (256x256 PNG recommended)
# Save as: assets/icon.png

# Commit and push
git add assets/icon.png
git commit -m "Add package icon"
git push
```

## Release Process

Once prerequisites are complete:

```bash
# 1. Verify environment variables
echo "GitHub Token: ${GITHUB_TOKEN:0:10}..."  # Should show ghp_xxxxx
echo "Chocolatey Key: ${CHOCOLATEY_API_KEY:0:10}..."  # Optional

# 2. Clean any previous builds
rm -rf dist/

# 3. Run GoReleaser
goreleaser release --clean

# 4. Verify the release
# - GitHub Release: https://github.com/steinmann321/fluxid-cli/releases/tag/v0.1.0
# - Homebrew Tap: https://github.com/steinmann321/homebrew-tap/blob/main/Formula/fluxid.rb
# - Chocolatey: https://community.chocolatey.org/packages/fluxid (may take 1-3 days for approval)
```

## Testing Installation

### Test Homebrew

```bash
# On macOS/Linux
brew tap steinmann321/tap
brew install fluxid
fluxid version

# Should show: fluxid 0.1.0 (commit: d2bb5a4...)
```

### Test Chocolatey

```powershell
# On Windows (PowerShell as Administrator)
choco install fluxid
fluxid version
```

## Manual Release (If GoReleaser Fails)

If you prefer to do the release manually without package managers for now:

```bash
# 1. Build binaries only
goreleaser build --clean --snapshot

# 2. Create GitHub release manually
# - Go to: https://github.com/steinmann321/fluxid-cli/releases/new
# - Select tag: v0.1.0
# - Upload files from dist/
# - Use RELEASE_NOTES_v0.1.0.md for description

# 3. Setup package managers later
# Follow PACKAGE_MANAGERS.md when ready
```

## Troubleshooting

### "missing GITHUB_TOKEN" error
- Ensure `export GITHUB_TOKEN="..."` is in your shell profile
- Reload shell: `source ~/.zshrc`
- Verify: `echo $GITHUB_TOKEN`

### "missing CHOCOLATEY_API_KEY" error
- Either set the API key: `export CHOCOLATEY_API_KEY="..."`
- Or edit `.goreleaser.yml` and set `skip_publish: true` under `chocolateys:`

### "repository not found: homebrew-tap"
- Create the repository: https://github.com/new
- Repository must be: steinmann321/homebrew-tap
- Must be public

### Homebrew formula not updating
- Check token has `repo` scope
- Verify repository permissions
- Check GoReleaser output for errors

## Next Steps After Release

1. **Test installation methods**:
   ```bash
   # Homebrew
   brew tap steinmann321/tap && brew install fluxid

   # Direct download
   curl -LO https://github.com/steinmann321/fluxid-cli/releases/download/v0.1.0/fluxid_darwin_arm64_v8.0.tar.gz
   ```

2. **Update README.md** with installation instructions

3. **Announce the release**:
   - Twitter/X
   - Reddit (r/golang, r/commandline)
   - Hacker News
   - Dev.to

4. **Monitor**:
   - GitHub issues
   - Homebrew installation feedback
   - Chocolatey moderation status

## Quick Reference

- **Project**: https://github.com/steinmann321/fluxid-cli
- **Releases**: https://github.com/steinmann321/fluxid-cli/releases
- **Homebrew Tap**: https://github.com/steinmann321/homebrew-tap
- **Chocolatey**: https://community.chocolatey.org/packages/fluxid
- **Issues**: https://github.com/steinmann321/fluxid-cli/issues

For detailed package manager setup, see: [PACKAGE_MANAGERS.md](./PACKAGE_MANAGERS.md)
