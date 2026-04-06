#!/bin/sh
set -e

GARAGE_ADMIN_TOKEN=$(cat ./garage/garage.toml| grep admin_token | awk -F" " '{print $NF}' | tr -d '"')

API="http://localhost:3903"
AUTH="Authorization: Bearer ${GARAGE_ADMIN_TOKEN}"
CT="Content-Type: application/json"

# Bucket/key definitions
BUREAUCAT_BUCKET=bureaucat
BUREAUCAT_KEY_NAME=bureaucat-bucket-key

# Wait for API to be ready
echo "Waiting for Garage admin API..."
until curl -sf -H "$AUTH" "$API/v2/GetClusterHealth" > /dev/null 2>&1; do
  sleep 2
done
echo "Garage API is up."

# 1. Get node ID from cluster status
NODE_ID=$(curl -sf -H "$AUTH" "$API/v2/GetClusterStatus" | jq -r '.nodes[0].id')
echo "Node ID: $NODE_ID"


# 2. Check if layout already has this node
if [ -z "$HAS_ROLE" ]; then
  echo "Assigning layout..."
  curl -sf -X POST -H "$AUTH" -H "$CT" \
    -d "{\"roles\":[{\"id\":\"$NODE_ID\",\"zone\":\"dc1\",\"capacity\":1000000000,\"tags\":[]}]}" \
    "$API/v2/UpdateClusterLayout" > /dev/null

  CURRENT_VERSION=$(echo "$LAYOUT" | jq -r '.version')
  NEXT_VERSION=$((CURRENT_VERSION + 1))

  curl -sf -X POST -H "$AUTH" -H "$CT" \
    -d "{\"version\":$NEXT_VERSION}" \
    "$API/v2/ApplyClusterLayout" > /dev/null

  echo "Layout applied (version $NEXT_VERSION). Waiting for cluster to be ready..."

  # Wait until layout is fully active
  until curl -sf -H "$AUTH" "$API/v2/GetClusterHealth" | jq -e '.status == "healthy"' > /dev/null 2>&1; do
    sleep 2
    echo "  ...still waiting for layout to propagate"
  done
  echo "Cluster is healthy."
else
  echo "Layout already configured, skipping."
fi

# Helper: create a key if it doesn't exist, sets ACCESS_KEY_ID and SECRET_KEY
create_key() {
  local KEY_NAME="$1"
  EXISTING_KEY=$(curl -sf -H "$AUTH" "$API/v2/ListKeys" | jq -r --arg name "$KEY_NAME" '.[] | select(.name == $name) | .id // empty')

  if [ -z "$EXISTING_KEY" ]; then
    echo "Creating key '$KEY_NAME'..."
    KEY_RESULT=$(curl -sf -X POST -H "$AUTH" -H "$CT" \
      -d "{\"name\":\"$KEY_NAME\"}" \
      "$API/v2/CreateKey")

    ACCESS_KEY_ID=$(echo "$KEY_RESULT" | jq -r '.accessKeyId')
    SECRET_KEY=$(echo "$KEY_RESULT" | jq -r '.secretAccessKey')

    echo "============================================"
    echo "[$KEY_NAME] ACCESS KEY ID:     $ACCESS_KEY_ID"
    echo "[$KEY_NAME] SECRET ACCESS KEY: $SECRET_KEY"
    echo "============================================"
  else
    ACCESS_KEY_ID="$EXISTING_KEY"
    echo "Key '$KEY_NAME' already exists ($ACCESS_KEY_ID), skipping."
  fi
}

# Helper: create a bucket if it doesn't exist and grant key access
create_bucket() {
  local BUCKET_NAME="$1"
  local KEY_ID="$2"
  EXISTING_BUCKET=$(curl -sf -H "$AUTH" "$API/v2/ListBuckets" | jq -r --arg name "$BUCKET_NAME" '.[] | select(.globalAliases[]? == $name) | .id // empty')

  if [ -z "$EXISTING_BUCKET" ]; then
    echo "Creating bucket '$BUCKET_NAME'..."
    BUCKET_RESULT=$(curl -sf -X POST -H "$AUTH" -H "$CT" \
      -d "{\"globalAlias\":\"$BUCKET_NAME\"}" \
      "$API/v2/CreateBucket")

    BUCKET_ID=$(echo "$BUCKET_RESULT" | jq -r '.id')

    # Grant read/write access
    curl -sf -X POST -H "$AUTH" -H "$CT" \
      -d "{\"bucketId\":\"$BUCKET_ID\",\"accessKeyId\":\"$KEY_ID\",\"permissions\":{\"read\":true,\"write\":true,\"owner\":false}}" \
      "$API/v2/AllowBucketKey" > /dev/null

    echo "Bucket '$BUCKET_NAME' created and key granted access."
  else
    echo "Bucket '$BUCKET_NAME' already exists, skipping."
  fi
}

# 3. Create BUREAUCAT key + bucket
create_key "$BUREAUCAT_KEY_NAME"
BUREAUCAT_ACCESS_KEY_ID="$ACCESS_KEY_ID"
create_bucket "$BUREAUCAT_BUCKET" "$BUREAUCAT_ACCESS_KEY_ID"

echo "Garage init complete."
