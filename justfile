image := "ghcr.io/bureaucatorg/bureaucat"
runner := "bureaucat-release:local"

podman-auth := env_var_or_default("REGISTRY_AUTH_FILE", env_var_or_default("XDG_RUNTIME_DIR", "/run/user/1000") + "/containers/auth.json")
podman-sock := env_var_or_default("PODMAN_SOCK", env_var_or_default("XDG_RUNTIME_DIR", "/run/user/1000") + "/podman/podman.sock")

# The whole release runs inside this image (go + bun + goreleaser + docker cli).
# The image's docker cli talks to the host podman socket and reuses `podman login` creds.
release-run := '''
podman run --rm \
    --security-opt label=disable \
    -v "$PWD":/app \
    -v ''' + podman-sock + ''':/var/run/docker.sock \
    -v ''' + podman-auth + ''':/root/.docker/config.json:ro \
    -e GITHUB_TOKEN \
    -e DOCKER_BUILDKIT=0 \
    -w /app \
    ''' + runner

# Show current version from latest git tag
version:
    @git describe --tags --abbrev=0 2>/dev/null || echo "no tags yet"

# Build the release-runner image (cached; rebuild when Dockerfile.release changes)
release-image:
    podman build -f Dockerfile.release -t {{runner}} .

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
    echo "Running goreleaser in Podman..."
    {{release-run}} goreleaser release --clean

    echo ""
    echo "Released ${version}"

# (Re)publish the current latest tag without bumping — use to resume a release
# that was tagged but failed partway (skips the git clean/tag-on-HEAD checks).
publish: release-image
    {{release-run}} goreleaser release --clean --skip=validate

# Build everything locally without publishing (snapshot, no tag/push required)
build: release-image
    {{release-run}} goreleaser release --clean --snapshot

# Build the Bureaucat CLI binary locally (host go toolchain)
build-cli:
    go build -o ./bureaucat ./cmd/bureaucat

# Run the Bureaucat CLI without building a binary first
cli *args:
    go run ./cmd/bureaucat {{args}}
