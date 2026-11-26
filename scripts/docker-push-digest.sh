#!/bin/sh
# Some docker-compatible engines omit the digest from `docker push` output, which goreleaser
# parses; after a push, fetch it from the registry and print it in docker's format.
set -e
[ "$1" = "push" ] || exec docker-real "$@"
docker-real "$@"
image="$2"
digest=$(docker-real manifest inspect -v "$image" | grep -o 'sha256:[a-f0-9]\{64\}' | head -n1)
[ -n "$digest" ] || { echo "could not resolve digest for $image" >&2; exit 1; }
echo "${image##*:}: digest: $digest"
