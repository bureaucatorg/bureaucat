#!/bin/sh
set -e

# Provision Garage (cluster layout + bucket + access key) via the garage CLI
# inside the container, and write the bucket credentials into .env. Uses
# `docker exec`, so it works for BOTH the dev and prod stacks without needing the
# admin HTTP port exposed.

CONTAINER="${GARAGE_CONTAINER:-bureaucat-garage}"
BUCKET="${FILES_BUCKET_NAME:-bureaucat}"
KEY_NAME=bureaucat-bucket-key
ENV_FILE="./.env"

g() { docker exec "$CONTAINER" /garage "$@" 2>/dev/null; }

echo "Waiting for Garage ($CONTAINER)..."
until g status >/dev/null 2>&1; do sleep 2; done

# 1. Assign a cluster layout if this node has no role yet.
NODE_ID=$(g status | awk '/NO ROLE ASSIGNED/{print $1}' | head -n1)
if [ -n "$NODE_ID" ]; then
  echo "Assigning layout to node $NODE_ID..."
  g layout assign "$NODE_ID" -z dc1 -c 1G >/dev/null
  g layout apply --version 1 >/dev/null
  # bucket/key ops fail with "Layout not ready" until the layout is live
  until g bucket list >/dev/null 2>&1; do sleep 1; done
  echo "Layout applied."
fi

# 2. Create the bucket if it doesn't exist.
if ! g bucket list | grep -qw "$BUCKET"; then
  echo "Creating bucket '$BUCKET'..."
  g bucket create "$BUCKET" >/dev/null
fi

# 3. Create the access key if missing (Garage only reveals the secret at creation).
SECRET_KEY=""
if g key list | grep -qw "$KEY_NAME"; then
  ACCESS_KEY_ID=$(g key list | awk -v n="$KEY_NAME" '$0 ~ n {print $1}' | head -n1)
  echo "Key '$KEY_NAME' already exists ($ACCESS_KEY_ID)."
else
  echo "Creating key '$KEY_NAME'..."
  KEY_OUT=$(g key create "$KEY_NAME")
  ACCESS_KEY_ID=$(echo "$KEY_OUT" | awk -F': *' '/Key ID/{print $2}' | tr -d ' ')
  SECRET_KEY=$(echo "$KEY_OUT"   | awk -F': *' '/Secret key/{print $2}' | tr -d ' ')
fi

# 4. Grant the key read+write on the bucket (idempotent).
g bucket allow --read --write "$BUCKET" --key "$KEY_NAME" >/dev/null

# 5. Wire the credentials into .env (only possible right after key creation).
if [ -n "$SECRET_KEY" ] && [ -f "$ENV_FILE" ]; then
  sed -i "s|^FILES_BUCKET_ACCESS_KEY_ID=.*|FILES_BUCKET_ACCESS_KEY_ID=${ACCESS_KEY_ID}|" "$ENV_FILE"
  sed -i "s|^FILES_BUCKET_SECRET_ACCESS_KEY=.*|FILES_BUCKET_SECRET_ACCESS_KEY=${SECRET_KEY}|" "$ENV_FILE"
  echo "Wrote bucket credentials into $ENV_FILE"
elif [ -z "$SECRET_KEY" ]; then
  echo "Key already existed; .env left unchanged (Garage reveals the secret only at creation)."
fi

echo "Garage init complete."
