#!/bin/bash
set -e

echo "=================================="
echo "Welcome to AgentBox Agent Container"
echo "=================================="

# Display available tools
echo ""
echo "Available Agents:"
command -v claude >/dev/null 2>&1 && echo "  - Claude Code: $(claude --version 2>/dev/null || echo 'installed')"
command -v codex >/dev/null 2>&1 && echo "  - Codex CLI: $(codex --version 2>/dev/null || echo 'installed')"

echo ""
echo "Development Tools:"
echo "  - Node.js: $(node --version)"
echo "  - npm: $(npm --version)"
echo "  - git: $(git --version | cut -d' ' -f3)"
command -v gh >/dev/null 2>&1 && echo "  - GitHub CLI: $(gh --version | head -1 | cut -d' ' -f3)"
command -v python3 >/dev/null 2>&1 && echo "  - Python: $(python3 --version | cut -d' ' -f2)"

echo ""
echo "Shell: $SHELL"
echo "Workspace: $(pwd)"
echo "=================================="
echo ""

# Execute the command passed to the container
exec "$@"
