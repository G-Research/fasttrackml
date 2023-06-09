# [Choice] Go version (use -bullseye variants on local arm64/Apple Silicon): 1, 1.18, 1.17, 1-bullseye, 1.18-bullseye, 1.17-bullseye, 1-buster, 1.18-buster, 1.17-buster
ARG VARIANT=1-bullseye
FROM mcr.microsoft.com/devcontainers/go:0-${VARIANT}

# [Choice] Node.js version: none, lts/*, 16, 14, 12, 10
ARG NODE_VERSION="none"
RUN if [ "${NODE_VERSION}" != "none" ]; then su vscode -c "umask 0002 && . /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends libssl-dev postgresql-client python-is-python3 python3-pip python3-dev python3-venv sqlite3 emacs-nox ripgrep fzf \
    && pip install pipenv

# [Optional] Uncomment the next lines to use go get to install anything else you need
USER vscode
RUN go install github.com/vektra/mockery/v2@v2.30.16 \
 && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3 \
 && go install golang.org/x/tools/cmd/goimports@v0.11.0 \
 && go install mvdan.cc/gofumpt@v0.5.0
USER root

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1

# Install xh
RUN curl -sfL https://raw.githubusercontent.com/ducaale/xh/master/install.sh | XH_BINDIR=/usr/local/bin sh