# AgentBox Universal Agent Image
# Contains both Claude Code and Codex CLI

FROM node:20-slim

LABEL maintainer="AgentBox"
LABEL org.opencontainers.image.title="AgentBox Agent"
LABEL org.opencontainers.image.description="Universal agent image with Claude Code and Codex"

# Install system dependencies (minimal)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code || echo "Claude Code not available yet"

# Install Codex CLI
RUN npm install -g @openai/codex || echo "Codex not available yet"

# Create non-root user
RUN useradd -m -s /bin/bash agent

# Set working directory
WORKDIR /workspace

# Switch to non-root user
USER agent

# Default command (keep container running)
CMD ["sleep", "infinity"]
