# AgentBox Universal Agent Image v2
# Based on Claude Code devcontainer configuration
# Includes common skill dependencies (covers 73% skills)

FROM node:20

ARG TZ=UTC
ENV TZ="$TZ"

ARG CLAUDE_CODE_VERSION=latest
ARG CODEX_VERSION=0.77.0

# Install development tools (aligned with claude-code devcontainer)
# + skill dependencies (whois, dnsutils, poppler, etc.)
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
    # === Skill Dependencies (v3) ===
    # DNS/WHOIS tools
    whois \
    dnsutils \
    # PDF tools
    poppler-utils \
    qpdf \
    # Image processing (for pyzbar)
    libzbar0 \
    # Build deps for pip packages
    libffi-dev \
    libxml2-dev \
    libxslt1-dev \
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

# === Python Skill Dependencies (v3) ===
# Install pip for node user
USER node
RUN curl -sSL https://bootstrap.pypa.io/get-pip.py -o /tmp/get-pip.py \
    && python3 /tmp/get-pip.py --user --break-system-packages --quiet \
    && rm /tmp/get-pip.py

ENV PATH="/home/node/.local/bin:${PATH}"

# Install common Python packages for skills (covers 73% of skills)
# Layer 1: Network libraries
RUN python3 -m pip install --user --break-system-packages --quiet \
    "requests>=2.28.0" \
    "aiohttp>=3.8.0" \
    "httpx>=0.24.0"

# Layer 2: DNS/Domain analysis
RUN python3 -m pip install --user --break-system-packages --quiet \
    "dnspython>=2.3.0" \
    "python-whois>=0.8.0" \
    "tld>=0.12.0" \
    "idna>=3.4"

# Layer 3: Data processing
RUN python3 -m pip install --user --break-system-packages --quiet \
    "beautifulsoup4>=4.12.0" \
    "lxml>=4.9.0" \
    "pandas>=2.0.0" \
    "pyyaml>=6.0"

# Layer 4: Document processing
RUN python3 -m pip install --user --break-system-packages --quiet \
    "pdfplumber>=0.9.0" \
    "python-docx>=0.8.11" \
    "openpyxl>=3.1.0"

# Layer 5: Image/QR code
RUN python3 -m pip install --user --break-system-packages --quiet \
    "pillow>=9.0.0" \
    "pyzbar>=0.1.9"

# Layer 6: CLI beautification
RUN python3 -m pip install --user --break-system-packages --quiet \
    "rich>=12.0.0" \
    "colorama>=0.4.6" \
    "tqdm>=4.65.0"

# Layer 7: Utilities
RUN python3 -m pip install --user --break-system-packages --quiet \
    "python-dotenv>=1.0.0" \
    "pyzipper>=0.3.6" \
    "py7zr>=0.20.0" \
    "defusedxml>=0.7.0"

# Switch to non-root user (already node)
# USER node

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
