# AgentBox Binary Reverse Engineering Image
# Based on agentbox/agent:v2, adds RE tools
# Tools: radare2, ghidra, frida, pwntools, yara, lief

FROM agentbox/agent:v2

LABEL maintainer="agentbox"
LABEL description="Binary reverse engineering and malware analysis"
LABEL version="1.0.0"

USER root

# ============ System Dependencies ============
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Java (for Ghidra)
    default-jdk \
    # Build tools
    cmake \
    pkg-config \
    # Radare2 dependencies
    libssl-dev \
    libmagic-dev \
    # YARA dependencies
    automake \
    libtool \
    flex \
    bison \
    libjansson-dev \
    && rm -rf /var/lib/apt/lists/*

# ============ Radare2 ============
ARG R2_VERSION=5.9.6
RUN git clone --depth 1 --branch ${R2_VERSION} https://github.com/radareorg/radare2.git /tmp/radare2 \
    && cd /tmp/radare2 \
    && sys/install.sh \
    && rm -rf /tmp/radare2

# ============ YARA ============
ARG YARA_VERSION=4.5.2
RUN git clone --depth 1 --branch v${YARA_VERSION} https://github.com/VirusTotal/yara.git /tmp/yara \
    && cd /tmp/yara \
    && ./bootstrap.sh \
    && ./configure --enable-cuckoo --enable-magic --enable-dotnet \
    && make -j$(nproc) \
    && make install \
    && ldconfig \
    && rm -rf /tmp/yara

# ============ Ghidra ============
ARG GHIDRA_VERSION=12.0.1
RUN wget -q "https://github.com/NationalSecurityAgency/ghidra/releases/download/Ghidra_${GHIDRA_VERSION}_build/ghidra_${GHIDRA_VERSION}_PUBLIC_20260114.zip" -O /tmp/ghidra.zip \
    && unzip -q /tmp/ghidra.zip -d /opt \
    && mv /opt/ghidra_${GHIDRA_VERSION}_PUBLIC /opt/ghidra \
    && rm /tmp/ghidra.zip \
    && chmod +x /opt/ghidra/ghidraRun /opt/ghidra/support/analyzeHeadless

ENV GHIDRA_HOME=/opt/ghidra
ENV PATH="${GHIDRA_HOME}/support:${PATH}"

# ============ Python Packages ============
USER node

# Layer 1: Core RE libraries
RUN python3 -m pip install --user --break-system-packages --quiet \
    "lief>=0.14.0" \
    "capstone>=5.0.0" \
    "keystone-engine>=0.9.2" \
    "unicorn>=2.0.0"

# Layer 2: pwntools (large)
RUN python3 -m pip install --user --break-system-packages --quiet \
    "pwntools>=4.11.0"

# Layer 3: Frida
RUN python3 -m pip install --user --break-system-packages --quiet \
    "frida-tools>=12.0.0"

# Layer 4: YARA Python bindings
RUN python3 -m pip install --user --break-system-packages --quiet \
    "yara-python>=4.5.0"

# Layer 5: Additional analysis tools
RUN python3 -m pip install --user --break-system-packages --quiet \
    "angr>=9.2.0" \
    "r2pipe>=1.8.0"

# ============ Verify Installation ============
RUN echo "=== Verifying RE tools ===" \
    && r2 -v | head -1 \
    && yara --version \
    && java -version 2>&1 | head -1 \
    && python3 -c "import lief, capstone, frida, yara; print('Python RE libs: OK')" \
    && echo "=== All RE tools installed ==="

WORKDIR /workspace
