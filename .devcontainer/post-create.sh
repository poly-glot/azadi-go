#!/usr/bin/env bash
set -euo pipefail

echo "=== Azadi Go Dev Container Setup ==="

# ============================================================
# Install Firestore emulator (gcloud CLI installed by devcontainer feature)
# ============================================================
sudo apt-get update -qq \
    && sudo apt-get install -y -qq --no-install-recommends google-cloud-cli-firestore-emulator \
    && sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*

# ============================================================
# Claude Code config - symlink ~/.claude.json from mounted dir
# ============================================================
if [ -f ~/.claude/.claude.json ] && [ ! -e ~/.claude.json ]; then
    ln -s ~/.claude/.claude.json ~/.claude.json
fi

# ============================================================
# NPM config
# ============================================================
npm config set cache ~/.npm
npm config set update-notifier false
npm config set fund false
npm config set audit false

# ============================================================
# Git config
# ============================================================
git config --global --add safe.directory /workspace
git config --global init.defaultBranch main
git config --global alias.st status
git config --global alias.co checkout
git config --global alias.ci commit

# ============================================================
# Shell aliases
# ============================================================
cat >> ~/.zshrc << 'ALIASES'

# Claude
alias claude="claude --dangerously-skip-permissions"

# Azadi Go dev aliases
alias dev='cd /workspace && set -a && source local.env && set +a && FIRESTORE_EMULATOR_HOST=localhost:8081 air'
alias dev-run='cd /workspace && set -a && source local.env && set +a && FIRESTORE_EMULATOR_HOST=localhost:8081 go run ./cmd/server'
alias dev-frontend="cd /workspace/frontend && npm run dev"
alias fb-emulator="gcloud emulators firestore start --host-port=0.0.0.0:8081 --database-mode=datastore-mode --project=demo-azadi"
alias fb-emulator-reset="kill $(lsof -ti:8081) 2>/dev/null; sleep 1; fb-emulator"
alias test-unit="cd /workspace && go test ./... -short -count=1"
alias test-all="cd /workspace && go test ./... -count=1 -race"
alias test-cover="cd /workspace && go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html"
alias lint="cd /workspace && golangci-lint run ./..."
alias build="cd /workspace && CGO_ENABLED=0 go build -o bin/azadi ./cmd/server"

# Reference implementation
alias ref="cd /workspace/azadi-reference"

# Docker
alias dc='docker compose'
alias dcup='docker compose up -d'
alias dcdown='docker compose down'

ALIASES

[ -f ~/.bashrc ] && ! grep -q 'exec zsh' ~/.bashrc && echo '[ -t 1 ] && exec zsh' >> ~/.bashrc

# ============================================================
# Install Go dependencies
# ============================================================
echo "Installing dependencies..."

if [ -f /workspace/go.mod ]; then
    (cd /workspace && go mod download > /tmp/go-download.log 2>&1 && echo "Go deps: OK" || echo "Go deps: FAILED") &
fi

if [ -f /workspace/frontend/package.json ]; then
    (cd /workspace/frontend && npm ci > /tmp/npm-install.log 2>&1 && echo "Frontend deps: OK" || echo "Frontend deps: FAILED") &
fi

wait

echo "=== Setup complete ==="
echo ""
echo "Quick start:"
echo "  fb-emulator      -> Start Firestore emulator (:8081)"
echo "  dev-frontend     -> Start Vite dev server (:5173)"
echo "  dev              -> Start Go server with hot reload (:8080)"
echo "  dev-run          -> Start Go server without hot reload (:8080)"
echo "  test-unit        -> Run unit tests"
echo "  test-all         -> Run all tests with race detector"
echo "  test-cover       -> Generate HTML coverage report"
echo "  lint             -> Run golangci-lint"
echo "  build            -> Build production binary"
echo ""
echo "Workflow: Open 3 terminals:"
echo "  1. fb-emulator     (Firestore)"
echo "  2. dev-frontend    (Vite HMR on :5173)"
echo "  3. dev             (Go server on :8080 with hot reload)"
echo "  -> Open http://localhost:8080"
echo ""
echo "Reference implementation: ~/Desktop/azadi is mounted at /workspace/azadi-reference"
echo ""
