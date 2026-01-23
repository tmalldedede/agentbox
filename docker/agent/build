#!/bin/bash
# Build AgentBox Agent Docker Images

set -e

REGISTRY="${REGISTRY:-agentbox}"
VERSION="${VERSION:-latest}"

echo "Building AgentBox Agent Images..."
echo "Registry: $REGISTRY"
echo "Version: $VERSION"

# Build Claude Code image
echo ""
echo "=== Building Claude Code Image ==="
docker build \
    --target claude-code \
    -t "$REGISTRY/claude-code:$VERSION" \
    -f docker/agent/Dockerfile \
    .

# Build Codex image
echo ""
echo "=== Building Codex Image ==="
docker build \
    --target codex \
    -t "$REGISTRY/codex:$VERSION" \
    -f docker/agent/Dockerfile \
    .

# Build combined image
echo ""
echo "=== Building Combined Image ==="
docker build \
    --target combined \
    -t "$REGISTRY/agent:$VERSION" \
    -f docker/agent/Dockerfile \
    .

echo ""
echo "=== Build Complete ==="
echo "Images built:"
echo "  - $REGISTRY/claude-code:$VERSION"
echo "  - $REGISTRY/codex:$VERSION"
echo "  - $REGISTRY/agent:$VERSION"
echo ""
echo "To push to registry:"
echo "  docker push $REGISTRY/claude-code:$VERSION"
echo "  docker push $REGISTRY/codex:$VERSION"
echo "  docker push $REGISTRY/agent:$VERSION"
