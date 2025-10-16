# Release Process

This document describes how to create a new release of MCP Proto Server.

## Pre-Release Checklist

Before creating a release:

- [ ] All tests pass locally (`python test_server.py`)
- [ ] Version number is updated (if versioning)
- [ ] CHANGELOG or release notes are prepared
- [ ] Code is committed and pushed to main branch
- [ ] Build works locally (`./build_local.sh` or `build_local.bat`)

## Creating a Release

### Step 1: Create a Git Tag

Create and push a version tag:

```bash
# Create an annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

Version naming convention:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.0.1` - Patch release (bug fixes)

### Step 2: GitHub Actions Builds

Once the tag is pushed, GitHub Actions will automatically:

1. **Build executables** for all platforms:
   - Linux (AMD64)
   - macOS Intel (AMD64)
   - macOS Apple Silicon (ARM64)
   - Windows (AMD64)

2. **Test each executable** to ensure it works

3. **Create a GitHub Release** with:
   - Auto-generated release notes
   - All platform binaries attached
   - Compressed archives (tar.gz) for Unix systems
   - Raw .exe for Windows

### Step 3: Verify the Release

After GitHub Actions completes (usually 5-10 minutes):

1. Visit: https://github.com/umuterturk/mcp-proto/releases
2. Verify the new release is listed
3. Check that all binaries are attached:
   - `mcp-proto-server-linux-amd64.tar.gz`
   - `mcp-proto-server-macos-amd64.tar.gz`
   - `mcp-proto-server-macos-arm64.tar.gz`
   - `mcp-proto-server-windows-amd64.exe`
4. Download and test at least one binary
5. Edit release notes if needed

## Manual Release (if needed)

If GitHub Actions fails or you need to build manually:

### Build All Platforms Locally

You'll need access to each platform (or use VMs/containers):

**On Linux:**
```bash
./build_local.sh
mv dist/mcp-proto-server mcp-proto-server-linux-amd64
tar -czf mcp-proto-server-linux-amd64.tar.gz mcp-proto-server-linux-amd64
```

**On macOS:**
```bash
./build_local.sh
mv dist/mcp-proto-server mcp-proto-server-macos-$(uname -m)
tar -czf mcp-proto-server-macos-$(uname -m).tar.gz mcp-proto-server-macos-$(uname -m)
```

**On Windows:**
```cmd
build_local.bat
move dist\mcp-proto-server.exe mcp-proto-server-windows-amd64.exe
```

### Create Release Manually

1. Go to: https://github.com/umuterturk/mcp-proto/releases/new
2. Choose the tag you created
3. Fill in release title and notes
4. Drag and drop all binaries
5. Publish release

## Release Notes Template

```markdown
## MCP Proto Server v1.0.0

### What's New
- Feature 1
- Feature 2
- Improvement 1

### Bug Fixes
- Fix 1
- Fix 2

### Breaking Changes
- Change 1 (if any)

### Installation

Download the appropriate binary for your platform:
- **Linux**: `mcp-proto-server-linux-amd64.tar.gz`
- **macOS Intel**: `mcp-proto-server-macos-amd64.tar.gz`
- **macOS Apple Silicon**: `mcp-proto-server-macos-arm64.tar.gz`
- **Windows**: `mcp-proto-server-windows-amd64.exe`

See [README.md](README.md) for setup instructions.

### Full Changelog
https://github.com/umuterturk/mcp-proto/compare/v0.9.0...v1.0.0
```

## Troubleshooting

### Build Fails on GitHub Actions

Check the Actions tab for error logs:
1. Go to: https://github.com/umuterturk/mcp-proto/actions
2. Click on the failed workflow run
3. Check the logs for each platform
4. Fix the issue and push a new commit
5. Delete the old tag and recreate it

### Missing Dependencies

If builds fail due to missing dependencies:
1. Update `requirements.txt`
2. Update `hiddenimports` in `mcp_proto_server.spec`
3. Test locally first
4. Push and retry

### Binary Doesn't Work

If a binary fails to run:
1. Check if all data files are included (see `datas` in spec file)
2. Verify hidden imports are complete
3. Test with `--verbose` flag for detailed logs
4. Check platform-specific requirements (e.g., glibc version on Linux)

## Post-Release

After releasing:

1. **Announce** the release:
   - Update README if needed
   - Post in relevant channels/forums
   - Tweet/social media (if applicable)

2. **Monitor** for issues:
   - Watch GitHub issues
   - Check download counts
   - Monitor for bug reports

3. **Update documentation**:
   - Ensure README reflects latest version
   - Update any version-specific docs

## Development Releases

For testing releases before official launch:

1. Use pre-release tags: `v1.0.0-rc1`, `v1.0.0-beta1`
2. Mark as "Pre-release" in GitHub
3. These won't show as "Latest" release
4. Perfect for testing with early users

```bash
git tag -a v1.0.0-rc1 -m "Release Candidate 1"
git push origin v1.0.0-rc1
```

## Rollback

If a release has critical issues:

1. Create a new patch release with the fix
2. OR delete the release and tag:
   ```bash
   # Delete remote tag
   git push --delete origin v1.0.0
   
   # Delete local tag
   git tag -d v1.0.0
   ```
3. Fix the issue and re-release

## Versioning Strategy

Semantic Versioning (SemVer):
- **MAJOR** (v2.0.0): Breaking changes
- **MINOR** (v1.1.0): New features, backwards compatible
- **PATCH** (v1.0.1): Bug fixes, backwards compatible

Pre-release identifiers:
- `alpha`: Very early, experimental
- `beta`: Feature complete, testing phase
- `rc`: Release candidate, production-ready testing

