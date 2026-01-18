# OpenAI Codex Agent Image
# Base image with Codex CLI installed

FROM node:20-slim

LABEL maintainer="AgentBox"
LABEL org.opencontainers.image.title="AgentBox Codex"
LABEL org.opencontainers.image.description="OpenAI Codex agent for AgentBox"

# Install dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    vim \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install Codex CLI
RUN npm install -g @openai/codex

# Create non-root user
RUN useradd -m -s /bin/bash agent
USER agent

# Set working directory
WORKDIR /workspace

# Default command (keep container running)
CMD ["sleep", "infinity"]
