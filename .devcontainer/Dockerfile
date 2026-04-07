FROM mcr.microsoft.com/devcontainers/go:1.24-bookworm

# Node.js 22 (for frontend build tooling & Firebase CLI)
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Firebase CLI
RUN npm install -g firebase-tools

# Go tools
RUN go install github.com/air-verse/air@v1.61.5 \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Create cache directories
RUN mkdir -p /home/vscode/.cache/go-build /home/vscode/go \
    && chown -R vscode:vscode /home/vscode/.cache /home/vscode/go

USER vscode
WORKDIR /workspace
