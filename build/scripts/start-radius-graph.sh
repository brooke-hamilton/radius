#!/bin/bash

# start-radius-graph.sh — Start Radius components using Git graph storage
# instead of PostgreSQL. Creates the state repo if it does not exist.

set -e

# Get the script directory and repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DEBUG_ROOT="$REPO_ROOT/debug_files"

# Graph store configuration
GRAPH_REPO_PATH="${GRAPH_REPO_PATH:-/tmp/radius-graphstore-repo}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Output helpers
print_info() { echo -e "\033[0;34mℹ${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }

check_prerequisites() {
  echo "🔍 Checking prerequisites..."
  local missing_tools=()

  command -v dlv >/dev/null 2>&1 || missing_tools+=("dlv -> go install github.com/go-delve/delve/cmd/dlv@latest")
  command -v go >/dev/null 2>&1 || missing_tools+=("go -> https://golang.org/doc/install")
  command -v k3d >/dev/null 2>&1 || missing_tools+=("k3d -> https://k3d.io/")
  command -v kubectl >/dev/null 2>&1 || missing_tools+=("kubectl -> https://kubernetes.io/docs/tasks/tools/")
  command -v terraform >/dev/null 2>&1 || missing_tools+=("terraform -> https://developer.hashicorp.com/terraform/install")
  command -v git >/dev/null 2>&1 || missing_tools+=("git -> https://git-scm.com/downloads")

  if [ ${#missing_tools[@]} -ne 0 ]; then
    print_error "Missing required tools:"
    for tool in "${missing_tools[@]}"; do
      echo "  - $tool"
    done
    exit 1
  fi

  # PostgreSQL is NOT required for graph store mode
  print_info "PostgreSQL is not required — all components use Git graph storage"
  print_success "Prerequisite check complete"
}

# Check debug binaries exist
if [ ! -f "$DEBUG_ROOT/bin/ucpd" ]; then
  print_error "Debug environment not found. Please run 'make debug-setup' first."
  exit 1
fi

mkdir -p "$DEBUG_ROOT/logs"

check_prerequisites

# ── Initialize the Git state repository ──────────────────────────────
echo "📦 Initializing Git state repository at $GRAPH_REPO_PATH..."
if [ -d "$GRAPH_REPO_PATH/.git" ]; then
  print_info "Git state repository already exists at $GRAPH_REPO_PATH"
else
  mkdir -p "$GRAPH_REPO_PATH"
  git -C "$GRAPH_REPO_PATH" init
  git -C "$GRAPH_REPO_PATH" commit --allow-empty -m "Initial commit (Radius graph store)"
  print_success "Created Git state repository at $GRAPH_REPO_PATH"
fi

# ── Stop existing components ─────────────────────────────────────────
echo "🧹 Stopping any existing components..."
for component in dynamic-rp applications-rp controller ucp; do
  if [ -f "$DEBUG_ROOT/logs/${component}.pid" ]; then
    pid=$(cat "$DEBUG_ROOT/logs/${component}.pid")
    if kill -0 "$pid" 2>/dev/null; then
      echo "Stopping existing $component (PID: $pid)"
      kill "$pid" 2>/dev/null || true
      sleep 2
      kill -0 "$pid" 2>/dev/null && kill -9 "$pid" 2>/dev/null || true
    fi
    rm -f "$DEBUG_ROOT/logs/${component}.pid"
  fi
done

if command -v pgrep >/dev/null 2>&1; then
  pkill -f "ucpd" 2>/dev/null || true
  pkill -f "applications-rp" 2>/dev/null || true
  pkill -f "dynamic-rp" 2>/dev/null || true
  pkill -f "controller.*--config-file.*controller.yaml" 2>/dev/null || true
  pkill -f "dlv.*exec.*ucpd" 2>/dev/null || true
  pkill -f "dlv.*exec.*applications-rp" 2>/dev/null || true
  pkill -f "dlv.*exec.*dynamic-rp" 2>/dev/null || true
  pkill -f "dlv.*exec.*controller" 2>/dev/null || true
else
  ps aux | grep -E "(ucpd|applications-rp|dynamic-rp|controller.*--config-file.*controller.yaml|dlv.*exec)" | grep -v grep | awk '{print $2}' | xargs -r kill 2>/dev/null || true
fi
print_success "Cleanup complete"

mkdir -p "$DEBUG_ROOT/logs"

# ── Create ucp-host Service + Endpoints in k3d ──────────────────────
# The deployment engine runs inside k3d and needs to reach UCP on the host.
# host.k3d.internal does not reliably connect in all environments (WSL2,
# Codespaces, Docker Desktop). Instead we detect the host IP from the k3d
# node's default gateway, create a headless Service "ucp-host" with manual
# Endpoints, and the deployment engine manifest references that Service.
echo "🌐 Setting up UCP host networking for k3d..."

# Detect the host IP reachable from inside the k3d node.
# We need the IP that allows a pod in k3d to reach host processes (UCP on :9000).
# Note: UCP is not running yet, so we cannot probe port 9000. Instead we find
# the IP via network topology and verify TCP reachability after UCP starts.
HOST_IP=""

# Method 1: WSL2 / Codespaces — the host eth0 IP is routable from k3d containers
# because both the host and the k3d Docker network share the same Linux kernel.
ETH0_IP=$(ip -4 addr show eth0 2>/dev/null | grep -oP 'inet \K[\d.]+' || true)
if [ -n "$ETH0_IP" ]; then
  HOST_IP="$ETH0_IP"
fi

# Method 2: native Linux Docker — the Docker bridge gateway routes to the host.
if [ -z "$HOST_IP" ] && command -v docker >/dev/null 2>&1; then
  K3D_NETWORK=$(docker inspect k3d-radius-debug-server-0 --format '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' 2>/dev/null || true)
  if [ -n "$K3D_NETWORK" ]; then
    GW_IP=$(docker network inspect "$K3D_NETWORK" -f '{{(index .IPAM.Config 0).Gateway}}' 2>/dev/null || true)
    if [ -n "$GW_IP" ]; then
      HOST_IP="$GW_IP"
    fi
  fi
fi

# Method 3: fall back to host.k3d.internal as resolved from the node.
if [ -z "$HOST_IP" ]; then
  HOST_IP=$(docker exec k3d-radius-debug-server-0 sh -c \
    "getent hosts host.k3d.internal | awk '{print \$1}'" 2>/dev/null || true)
fi

if [ -z "$HOST_IP" ]; then
  print_error "Could not detect host IP for k3d networking"
  exit 1
fi

print_info "Host IP for k3d → UCP connectivity: $HOST_IP"

# Apply the Service + Endpoints that route ucp-host:9000 → $HOST_IP:9000
kubectl --context k3d-radius-debug apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: ucp-host
  namespace: default
spec:
  ports:
  - port: 9000
    targetPort: 9000
    name: ucp
  clusterIP: None
---
apiVersion: v1
kind: Endpoints
metadata:
  name: ucp-host
  namespace: default
subsets:
- addresses:
  - ip: "$HOST_IP"
  ports:
  - port: 9000
    name: ucp
EOF
print_success "Created ucp-host Service → $HOST_IP:9000"

# ── Update repoPath in graphstore configs (idem if already correct) ──
echo "📝 Configuring graph store repo path: $GRAPH_REPO_PATH"
for cfg in ucp-graphstore.yaml applications-rp-graphstore.yaml dynamic-rp-graphstore.yaml; do
  cfg_path="$SCRIPT_DIR/../configs/$cfg"
  if [ -f "$cfg_path" ]; then
    # Use a portable sed: replace only the repoPath value line
    if sed --version >/dev/null 2>&1; then
      # GNU sed
      sed -i "s|repoPath:.*|repoPath: \"$GRAPH_REPO_PATH\"|" "$cfg_path"
    else
      # BSD/macOS sed
      sed -i '' "s|repoPath:.*|repoPath: \"$GRAPH_REPO_PATH\"|" "$cfg_path"
    fi
  fi
done

# ── Start UCP ────────────────────────────────────────────────────────
echo "🚀 Starting UCP (graph store) with dlv on port 40001..."
dlv exec "$DEBUG_ROOT/bin/ucpd" \
  --listen=127.0.0.1:40001 --headless=true --api-version=2 --accept-multiclient --continue \
  -- --config-file="$SCRIPT_DIR/../configs/ucp-graphstore.yaml" \
  > "$DEBUG_ROOT/logs/ucp.log" 2>&1 &
echo $! > "$DEBUG_ROOT/logs/ucp.pid"

echo "Waiting for UCP to initialize (may take up to 2 minutes)..."
max_attempts=60
attempt=0
while [ $attempt -lt $max_attempts ]; do
  if curl -s "http://localhost:9000/apis/api.ucp.dev/v1alpha3" > /dev/null 2>&1; then
    if grep -q "Successfully registered manifests" "$DEBUG_ROOT/logs/ucp.log" 2>/dev/null; then
      break
    fi
  fi
  [ $((attempt % 10)) -eq 0 ] && [ $attempt -gt 0 ] && echo "  Still waiting for UCP... (${attempt}s elapsed)"
  sleep 2
  attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
  print_error "UCP failed to start within 2 minutes"
  echo "Check: $DEBUG_ROOT/logs/ucp.log"
  exit 1
fi
print_success "UCP started (graph store)"

# ── Start Controller (no database — unchanged) ──────────────────────
echo "Starting Controller with dlv on port 40002..."
dlv exec "$DEBUG_ROOT/bin/controller" \
  --listen=127.0.0.1:40002 --headless=true --api-version=2 --accept-multiclient --continue \
  -- --config-file="$SCRIPT_DIR/../configs/controller.yaml" --cert-dir="" \
  > "$DEBUG_ROOT/logs/controller.log" 2>&1 &
echo $! > "$DEBUG_ROOT/logs/controller.pid"

attempt=0; max_attempts=15
while [ $attempt -lt $max_attempts ]; do
  curl -s "http://localhost:7073/healthz" > /dev/null 2>&1 && break
  sleep 2; attempt=$((attempt + 1))
done
[ $attempt -eq $max_attempts ] \
  && print_warning "Controller health check timed out (check: $DEBUG_ROOT/logs/controller.log)" \
  || print_success "Controller started"

# ── Start Applications RP (graph store) ──────────────────────────────
echo "Starting Applications RP (graph store) with dlv on port 40003..."
dlv exec "$DEBUG_ROOT/bin/applications-rp" \
  --listen=127.0.0.1:40003 --headless=true --api-version=2 --accept-multiclient --continue \
  -- --config-file="$SCRIPT_DIR/../configs/applications-rp-graphstore.yaml" \
  > "$DEBUG_ROOT/logs/applications-rp.log" 2>&1 &
echo $! > "$DEBUG_ROOT/logs/applications-rp.pid"

attempt=0; max_attempts=15
while [ $attempt -lt $max_attempts ]; do
  curl -s "http://localhost:8080/healthz" > /dev/null 2>&1 && break
  sleep 2; attempt=$((attempt + 1))
done
[ $attempt -eq $max_attempts ] \
  && print_warning "Applications RP health check timed out (check: $DEBUG_ROOT/logs/applications-rp.log)" \
  || print_success "Applications RP started (graph store)"

# ── Start Dynamic RP (graph store) ──────────────────────────────────
echo "Starting Dynamic RP (graph store) with dlv on port 40004..."
dlv exec "$DEBUG_ROOT/bin/dynamic-rp" \
  --listen=127.0.0.1:40004 --headless=true --api-version=2 --accept-multiclient --continue \
  -- --config-file="$SCRIPT_DIR/../configs/dynamic-rp-graphstore.yaml" \
  > "$DEBUG_ROOT/logs/dynamic-rp.log" 2>&1 &
echo $! > "$DEBUG_ROOT/logs/dynamic-rp.pid"

attempt=0; max_attempts=15
while [ $attempt -lt $max_attempts ]; do
  curl -s "http://localhost:8082/healthz" > /dev/null 2>&1 && break
  sleep 2; attempt=$((attempt + 1))
done
[ $attempt -eq $max_attempts ] \
  && print_warning "Dynamic RP health check timed out (check: $DEBUG_ROOT/logs/dynamic-rp.log)" \
  || print_success "Dynamic RP started (graph store)"

# ── Summary ──────────────────────────────────────────────────────────
echo ""
echo "🎉 All components started with Git graph storage!"
echo ""
echo "🔗 UCP API:          http://localhost:9000  (dlv 40001)"
echo "🔗 Applications RP:  http://localhost:8080  (dlv 40003)"
echo "🔗 Dynamic RP:       http://localhost:8082  (dlv 40004)"
echo "🔗 Controller:       http://localhost:7073  (dlv 40002)"
echo ""
echo "📦 Git state repo:   $GRAPH_REPO_PATH"
echo "   Inspect state:    git -C $GRAPH_REPO_PATH for-each-ref refs/infra/"
echo "   View commits:     git -C $GRAPH_REPO_PATH log refs/infra/radius-ucp --oneline"
echo "   Browse tree:      git -C $GRAPH_REPO_PATH ls-tree -r refs/infra/radius-ucp --name-only"
echo ""
echo "🐛 Attach VS Code debugger to dlv ports 40001-40004"
