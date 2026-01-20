# AgentBox Universal Agent Image
# Contains Claude Code, Codex, and OpenCode CLI

FROM node:18-alpine

# 代理配置
ARG HTTP_PROXY
ARG HTTPS_PROXY
ENV HTTP_PROXY=${HTTP_PROXY} HTTPS_PROXY=${HTTPS_PROXY}
ENV http_proxy=${HTTP_PROXY} https_proxy=${HTTPS_PROXY}

LABEL maintainer="AgentBox"
LABEL org.opencontainers.image.title="AgentBox Agent"
LABEL org.opencontainers.image.description="Universal agent image with Claude Code, Codex, and OpenCode"

# Install system dependencies
RUN apk add --no-cache \
    git \
    curl \
    ca-certificates \
    bash \
    shadow

# Create non-root user (compatible with Codex which expects /home/agent/)
RUN useradd -m -s /bin/bash agent

# Install Claude Code CLI (if available) - global, accessible to all users
RUN npm install -g @anthropic-ai/claude-code 2>/dev/null || echo "Claude Code CLI not available"

# Install Codex CLI (if available) - global, accessible to all users
RUN npm install -g @openai/codex 2>/dev/null || echo "Codex CLI not available"

# Install OpenCode CLI (package name is opencode-ai on npm)
# Note: Alpine uses musl libc, so we need to link the musl version
RUN npm install -g opencode-ai 2>/dev/null && \
    cd /usr/local/lib/node_modules/opencode-ai/node_modules && \
    rm -rf opencode-linux-x64 && \
    ln -sf opencode-linux-x64-musl opencode-linux-x64 || \
    echo "OpenCode CLI not available via npm"

# Install Bun for agent user (required for OpenCode)
USER agent
RUN curl -fsSL https://bun.sh/install | bash
ENV PATH="/home/agent/.bun/bin:${PATH}"

# Set working directory
WORKDIR /workspace

# Switch back to root for container startup, then drop to agent
USER root
RUN chown -R agent:agent /workspace

# Run as agent user
USER agent

# Default command (keep container running)
CMD ["sleep", "infinity"]
