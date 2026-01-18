# Claude Code Agent Image
# Base image with Claude Code CLI installed

FROM node:20-slim

LABEL maintainer="AgentBox"
LABEL org.opencontainers.image.title="AgentBox Claude Code"
LABEL org.opencontainers.image.description="Claude Code agent for AgentBox"

# Install dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    vim \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Create non-root user
RUN useradd -m -s /bin/bash agent
USER agent

# Set working directory
WORKDIR /workspace

# Default command (keep container running)
CMD ["sleep", "infinity"]
