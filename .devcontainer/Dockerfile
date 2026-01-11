FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

# Install ripgrep
RUN curl -LO https://github.com/BurntSushi/ripgrep/releases/download/15.1.0/ripgrep_15.1.0-1_amd64.deb \
    && dpkg -i ripgrep_15.1.0-1_amd64.deb \
    && rm ripgrep_15.1.0-1_amd64.deb
