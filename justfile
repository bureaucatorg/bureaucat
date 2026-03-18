image := "docker.io/codingcoffee/bureaucat"

# Show current version from latest git tag
version:
    @git describe --tags --abbrev=0 2>/dev/null || echo "no tags yet"

# Release a new version: just release patch|minor|major
release kind="patch":
    #!/usr/bin/env bash
    set -euo pipefail

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

    # Tag and push
    git tag -a "${version}" -m "Release ${version}"
    git push origin "${version}"
    echo "Pushed tag ${version} to origin"

    # Build docker image
    echo "Building docker image..."
    docker build \
        --build-arg VERSION="${version}" \
        -t {{image}}:${version} \
        -t {{image}}:latest \
        .

    # Push to docker hub
    echo "Pushing to Docker Hub..."
    docker push {{image}}:${version}
    docker push {{image}}:latest

    echo ""
    echo "Released ${version}"

# Build docker image without releasing (uses current git describe)
build:
    #!/usr/bin/env bash
    set -euo pipefail
    version=$(git describe --tags --always --dirty)
    echo "Building ${version}..."
    docker build \
        --build-arg VERSION="${version}" \
        -t {{image}}:${version} \
        -t {{image}}:latest \
        .
    echo "Built {{image}}:${version}"

# Push the latest built image to Docker Hub
push:
    #!/usr/bin/env bash
    set -euo pipefail
    version=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
    docker push {{image}}:${version}
    docker push {{image}}:latest
    echo "Pushed {{image}}:${version} and latest"
