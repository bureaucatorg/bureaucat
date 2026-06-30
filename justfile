image := "docker.io/codingcoffee/bureaucat"
runner := "bureaucat-release:local"

# The whole release runs inside this image (go + bun + goreleaser + docker cli).
# Host docker socket + docker config + GITHUB_TOKEN are passed through so
# goreleaser can build/push the image and create the GitHub release.
release-run := '''
docker run --rm \
    -v "$PWD":/app \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$HOME/.docker":/root/.docker \
    -e GITHUB_TOKEN \
    -e DOCKER_CONFIG=/root/.docker \
    -w /app \
    ''' + runner

# Show current version from latest git tag
version:
    @git describe --tags --abbrev=0 2>/dev/null || echo "no tags yet"

# Build the release-runner image (cached; rebuild when Dockerfile.release changes)
release-image:
    docker build -f Dockerfile.release -t {{runner}} .

# Release a new version: just release patch|minor|major
release kind="patch": release-image
    #!/usr/bin/env bash
    set -euo pipefail

    if [[ -z "${GITHUB_TOKEN:-}" ]]; then
        echo "GITHUB_TOKEN is not set (needed to publish the GitHub release)."; exit 1
    fi

    # Get current version from latest tag (default v0.0.0 if no tags)
    current=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    current="${current#v}"
    IFS='.' read -r major minor patch <<< "$current"

    case "{{kind}}" in
        patch) patch=$((patch + 1)) ;;
        minor) minor=$((minor + 1)); patch=0 ;;
        major) major=$((major + 1)); minor=0; patch=0 ;;
        *) echo "usage: just release [patch|minor|major]"; exit 1 ;;
    esac

    version="v${major}.${minor}.${patch}"
    echo "Current: v${current}"
    echo "Next:    ${version}"
    echo ""

    read -p "Proceed? [y/N] " confirm
    [[ "$confirm" =~ ^[yY]$ ]] || { echo "Aborted."; exit 1; }

    # Update VERSION file
    echo "${version}" > VERSION
    git add VERSION
    git commit -m "chore: bump version to ${version}"
    git push origin HEAD

    # Tag and push (goreleaser builds the release from this tag)
    git tag -a "${version}" -m "Release ${version}"
    git push origin "${version}"
    echo "Pushed tag ${version} to origin"

    # Build binaries + frontend, push docker image, create GitHub release —
    # all inside the runner container.
    echo "Running goreleaser in Docker..."
    {{release-run}} goreleaser release --clean

    echo ""
    echo "Released ${version}"

# Build everything locally without publishing (snapshot, no tag/push required)
build: release-image
    {{release-run}} goreleaser release --clean --snapshot

# Build the Bureaucat CLI binary locally (host go toolchain)
build-cli:
    go build -o ./bureaucat ./cmd/bureaucat

# Run the Bureaucat CLI without building a binary first
cli *args:
    go run ./cmd/bureaucat {{args}}
