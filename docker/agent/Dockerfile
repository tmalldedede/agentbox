# AgentBox Universal Agent Image
# Based on Claude Code devcontainer configuration

FROM node:20

ARG TZ=UTC
ENV TZ="$TZ"

ARG CLAUDE_CODE_VERSION=latest
ARG CODEX_VERSION=latest

# Install development tools (aligned with claude-code devcontainer)
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Basic tools
    less \
    git \
    git-lfs \
    procps \
    sudo \
    curl \
    wget \
    ca-certificates \
    # Development utilities
    fzf \
    zsh \
    man-db \
    unzip \
    gnupg2 \
    jq \
    nano \
    vim \
    # GitHub CLI
    gh \
    # Build essentials (for native modules)
    build-essential \
    python3 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Ensure node user has access to npm global
RUN mkdir -p /usr/local/share/npm-global && \
    chown -R node:node /usr/local/share

ARG USERNAME=node

# Persist bash history
RUN mkdir /commandhistory && \
    touch /commandhistory/.bash_history && \
    chown -R $USERNAME /commandhistory

# Set environment variables
ENV DEVCONTAINER=true
ENV NPM_CONFIG_PREFIX=/usr/local/share/npm-global
ENV PATH=$PATH:/usr/local/share/npm-global/bin

# Create workspace and config directories
RUN mkdir -p /workspace /home/node/.claude /home/node/.codex && \
    chown -R node:node /workspace /home/node/.claude /home/node/.codex

WORKDIR /workspace

# Install git-delta for better diffs (optional but nice)
ARG GIT_DELTA_VERSION=0.18.2
RUN ARCH=$(dpkg --print-architecture) && \
    wget -q "https://github.com/dandavison/delta/releases/download/${GIT_DELTA_VERSION}/git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb" && \
    dpkg -i "git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb" && \
    rm "git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb"

# Switch to non-root user
USER node

# Set default shell and editor
ENV SHELL=/bin/zsh
ENV EDITOR=nano
ENV VISUAL=nano

# Install zsh with powerline theme
ARG ZSH_IN_DOCKER_VERSION=1.2.0
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v${ZSH_IN_DOCKER_VERSION}/zsh-in-docker.sh)" -- \
    -p git \
    -p fzf \
    -a "source /usr/share/doc/fzf/examples/key-bindings.zsh 2>/dev/null || true" \
    -a "source /usr/share/doc/fzf/examples/completion.zsh 2>/dev/null || true" \
    -a "export HISTFILE=/commandhistory/.bash_history" \
    -x

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION} || echo "Claude Code installation skipped"

# Install Codex CLI
RUN npm install -g @openai/codex@${CODEX_VERSION} || echo "Codex installation skipped"

# Copy entrypoint script
USER root
COPY entrypoint.sh /opt/entrypoint.sh
RUN chmod +x /opt/entrypoint.sh
USER node

# Environment for agent configuration
ENV CLAUDE_CONFIG_DIR=/home/node/.claude
ENV CODEX_CONFIG_DIR=/home/node/.codex
ENV NODE_OPTIONS="--max-old-space-size=4096"

ENTRYPOINT ["/opt/entrypoint.sh"]
CMD ["sleep", "infinity"]
