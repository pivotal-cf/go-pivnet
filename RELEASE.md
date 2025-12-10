# Release Process

This document describes how to create a new release for go-pivnet.

## Prerequisites

1. Install [goreleaser](https://goreleaser.com/install/):
   ```bash
   brew install goreleaser
   # or
   go install github.com/goreleaser/goreleaser@latest
   ```

2. Ensure you have a GitHub token with appropriate permissions set:
   ```bash
   export GITHUB_TOKEN=your_github_token
   ```

## Creating a Release

### 1. Create and Push a Git Tag

Create a new tag following semantic versioning (e.g., `v7.1.0`):

```bash
# Create an annotated tag
git tag -a v7.1.0 -m "Release v7.1.0"

# Push the tag to GitHub
git push origin v7.1.0
```

Or create a tag from the GitHub UI when creating a release.

### 2. Run GoReleaser

Run goreleaser to create the release:

```bash
goreleaser release
```

This will:
- Generate a changelog from git commits
- Create a draft GitHub release
- Upload any build artifacts (if builds are configured)

### 3. Review and Publish

1. Go to the [GitHub releases page](https://github.com/pivotal-cf/go-pivnet/releases)
2. Review the draft release
3. Edit the release notes if needed
4. Click "Publish release" when ready

## Dry Run (Testing)

To test the release process without actually creating a release:

```bash
goreleaser release --snapshot
```

This creates a snapshot release that won't be published to GitHub.

**What happens with `--snapshot`:**
- Creates local artifacts in the `dist/` folder
- Does **NOT** create a GitHub release
- Useful for testing the release process locally

**To check snapshot results:**
- Look in the `dist/` folder for generated files
- Check `dist/metadata.json` for release metadata
- Since builds are skipped for this library, the folder will mainly contain metadata files

**To create an actual release on GitHub:**
- Run `goreleaser release` (without `--snapshot`)
- Then check: https://github.com/pivotal-cf/go-pivnet/releases
- The release will appear as a **draft** that you can review and publish

## Notes

- Releases are created as **drafts** by default, so you can review before publishing
- The changelog automatically excludes commits starting with `docs:`, `test:`, and `chore:`
- Make sure your git tags follow semantic versioning (e.g., `v7.1.0`, `v7.2.0`)

