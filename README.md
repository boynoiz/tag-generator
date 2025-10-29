# Release Tool

A simple Git tag generator that uses CalVer format so you don't have to think about version numbers.

## Quick Start

```bash
# Build the tool
make build

# Initialize config (first time only)
./dist/release init

# Create a tag
./dist/release          # On main: creates v2025.10.4.1.0
./dist/release --fix    # On main: creates v2025.10.4.1.1
./dist/release          # On dev branch: creates dev-1078645
```

## Installation

```bash
# Build
make build

# Or install to GOPATH/bin
make install

# Or build for all platforms
make build-all
```

## Configuration

Run `release init` to create `.release/config.yaml`:

```yaml
use_prefix: true           # Add prefix to release tags
prefix: v                  # Prefix for release branch (e.g., "v" for v2025.10.4.1.0)
dev_prefix: dev-          # Prefix for dev branches (e.g., "dev-" for dev-1078645)
release_branch: main       # Branch for CalVer tags
```

Commit this config file to your repo so the whole team uses the same settings.

**Why separate prefixes?** Different prefixes make Harbor/registry cleanup rules easy:
- Keep: `^v[0-9]+` → Keeps all versioned releases
- Clean: `^dev-` → Removes all dev builds older than X days

## Tag Behavior

### On Release Branch (default: main)
Creates **CalVer tags** with format: `Year.Month.Week.Release.Fix`

Examples:
- `v2025.10.4.1.0` - First release in week 4 of October 2025
- `v2025.10.4.1.1` - Hotfix for that release
- `v2025.10.4.2.0` - Second release in the same week
- `v2025.11.1.1.0` - New month, resets to week 1, release 1

**Release** (default):
- Increments the release number
- Resets to `.1.0` when moving to a new week
- Use: `release` or `release --fix` for hotfixes

### On Other Branches
Creates **hash-based tags** using git short hash (7 chars)

Examples:
- `dev-1078645` - Tag based on commit hash
- Perfect for dev/staging builds in CI/CD

## CI/CD Integration

This works great in CI/CD pipelines with Harbor or any container registry:

```yaml
# Example: Harbor Tag Retention Rules

# Rule 1: Keep all release tags
- Pattern: ^v[0-9]+
  Action: Keep forever

# Rule 2: Clean up dev tags
- Pattern: ^dev-
  Action: Remove if older than 30 days
```

**Result:**
- Release branch → versioned tags (`v2025.10.4.1.0`) - kept forever
- Dev branches → hash tags (`dev-1078645`) - auto-cleaned after 30 days

The separate prefixes (`v` vs `dev-`) make Harbor cleanup rules trivial!

## Development

```bash
# Install dev tools
make tools-install

# Format and lint
make fmt
make lint-fix

# Run all checks
make qa

# Run tests
make test
```

## Why This Exists

I got lazy and annoyed every time I needed to figure out what the latest tag was and decide what the new tag should be. So I made this tool that just uses the date instead. No more thinking about whether it should be 1.2.3 or 2.0.0 or whatever. It's just the date. Simple.

Plus, having hash-based tags for dev branches makes Harbor registry cleanup a breeze.
