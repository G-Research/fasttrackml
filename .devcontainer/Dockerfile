# [Choice] Go version (use -bookworm, or -bullseye variants on local arm64/Apple Silicon): 1, 1.21, 1.20, 1-bookworm, 1.21-bookworm, 1.20-bookworm, 1-bullseye, 1.21-bullseye, 1.20-bullseye
ARG VARIANT=1-bookworm
FROM mcr.microsoft.com/devcontainers/go:1-${VARIANT}

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get -y install --no-install-recommends \
    emacs-nox \
    fzf \
    libssl-dev \
    pipenv \
    postgresql-client \
    python-is-python3 \
    python3-dev \
    python3-pip \
    python3-venv \
    ripgrep \
    sqlite3

# [Optional] Uncomment the next lines to use go get to install anything else you need
USER vscode
RUN go install github.com/vektra/mockery/v2@v2.32.4 \
 && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.1 \
 && go install golang.org/x/tools/cmd/goimports@v0.11.0 \
 && go install mvdan.cc/gofumpt@v0.5.0
USER root

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1

# Install xh
RUN curl -sfL https://raw.githubusercontent.com/ducaale/xh/master/install.sh | XH_BINDIR=/usr/local/bin sh

# Add docker-compose symlink to Docker Compose v2
# This is a workaround until https://github.com/devcontainers/features/pull/621 lands
RUN ln -s /usr/libexec/docker/cli-plugins/docker-compose /usr/local/bin/