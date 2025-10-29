# Release Tool

A simple Git tag generator that uses CalVer format so you don't have to think about version numbers.

## Tag Format

Tags follow this pattern: `Year.Month.Week.Release.Fix`

Examples:
- `2025.10.4.1.0` - First release in week 4 of October 2025
- `2025.10.4.1.1` - Hotfix for that release
- `2025.10.4.2.0` - Second release in the same week
- `2025.11.1.1.0` - New month, resets to week 1, release 1

## Installation

```bash
go build -o release ./cmd/release
```

Move the binary somewhere in your PATH if you want to use it globally.

## Usage

```bash
# Create a release tag (default)
release

# Create a fix/patch tag
release --fix
release -f
```

### What's the difference?

**Release** (default):
- Increments the release number
- Resets to `.1.0` when moving to a new week
- Use this for regular releases

**Fix**:
- Only increments the fix number (last digit)
- Use this for hotfixes/patches

The tool will automatically:
- Check you're on the `main` branch
- Make sure current commit isn't already tagged
- Calculate the next version number
- Create and push the tag

## Why this exists

I got lazy and annoyed every time I needed to figure out what the latest tag was and decide what the new tag should be. So I made this tool that just uses the date instead. No more thinking about whether it should be 1.2.3 or 2.0.0 or whatever. It's just the date. Simple.
