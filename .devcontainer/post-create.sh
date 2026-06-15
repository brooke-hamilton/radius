#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly GOLANGCI_LINT_VERSION_FILE="${SCRIPT_DIR}/../.golangci-lint-version"

if [[ ! -f "${GOLANGCI_LINT_VERSION_FILE}" ]]; then
    echo "Error: missing golangci-lint version file: ${GOLANGCI_LINT_VERSION_FILE}" >&2
    exit 1
fi

# Strip line endings and surrounding whitespace from the version value.
GOLANGCI_LINT_VERSION="$(tr -d '\r\n' < "${GOLANGCI_LINT_VERSION_FILE}")"
GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION#"${GOLANGCI_LINT_VERSION%%[![:space:]]*}"}"
GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION%"${GOLANGCI_LINT_VERSION##*[![:space:]]}"}"

if [[ -z "${GOLANGCI_LINT_VERSION}" ]]; then
    echo "Error: golangci-lint version file is empty: ${GOLANGCI_LINT_VERSION_FILE}" >&2
    exit 1
fi
readonly GOLANGCI_LINT_VERSION

echo "============================================================================"
echo "Starting post-create setup..."
echo "============================================================================"

# Set SHELL for pnpm setup (not always set in devcontainer post-create context)
echo "Setting SHELL environment variable..."
export SHELL="${SHELL:-/bin/bash}"

# Resolve the repository root from this script's location so the script works
# both inside the dev container (/workspaces/radius) and on a CI runner such as
# GitHub Actions, where the checkout lives in ${GITHUB_WORKSPACE}.
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly REPO_ROOT

# Add the workspace as a git safe directory to avoid "dubious ownership" errors.
# Guard with a lookup so repeated runs do not append duplicate entries.
echo "Adding workspace as git safe directory..."
if ! git config --global --get-all safe.directory 2>/dev/null | grep -Fxq "${REPO_ROOT}"; then
    git config --global --add safe.directory "${REPO_ROOT}"
fi

# Install pnpm via corepack
echo "Installing pnpm via corepack..."
make generate-pnpm-installed

# Configure pnpm store directory inside the container to avoid hard-link issues
# with mounted workspace filesystem (hard links cannot cross filesystem boundaries)
echo "Configuring pnpm store directory..."
pnpm config set store-dir /tmp/.pnpm-store

# Install the binary form of golangci-lint, as recommended
# https://golangci-lint.run/welcome/install/#local-installation
echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION}..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b "$(go env GOPATH)/bin" "${GOLANGCI_LINT_VERSION}"

# Other go tools
echo "Installing controller-gen..."
go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0

echo "Installing mockgen..."
go install go.uber.org/mock/mockgen@v0.4.0

echo "Installing cspell..."
# Ensure pnpm global bin directory exists and is on PATH before installing
# global packages. `pnpm setup` updates shell rc files for future sessions,
# but we also need PATH updated for the current script execution.
# Derive from ${HOME} so the path is valid both in the dev container (vscode
# user) and on a CI runner such as GitHub Actions (runner user).
export PNPM_HOME="${PNPM_HOME:-${HOME}/.local/share/pnpm}"
PNPM_BIN_DIR="$(pnpm config get global-bin-dir 2>/dev/null || true)"
if [[ -z "${PNPM_BIN_DIR}" || "${PNPM_BIN_DIR}" == "undefined" ]]; then
    PNPM_BIN_DIR="${PNPM_HOME}/bin"
    pnpm config set global-bin-dir "${PNPM_BIN_DIR}"
fi
mkdir -p "${PNPM_BIN_DIR}"
export PATH="${PNPM_BIN_DIR}:${PATH}"
pnpm add -g cspell

echo "============================================================================"
echo "Post-create setup completed successfully!"
echo "============================================================================"
