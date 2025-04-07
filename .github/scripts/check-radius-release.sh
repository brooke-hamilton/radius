#!/bin/bash
# ------------------------------------------------------------
# Copyright 2025 The Radius Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#    
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ------------------------------------------------------------

# This script checks for new Radius releases and returns the latest release version
# It compares with the last tested version (if provided) and determines if
# we need to perform a clean install

set -e

CACHE_FILE="${1:-./dist/cache/.last_tested_release}"
RELEASES_URL="https://api.github.com/repos/radius-project/radius/releases"

# Function to check if rad is installed
is_rad_installed() {
    if command -v rad &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to get current rad version
get_current_rad_version() {
    rad version | grep -oP 'Version: \K[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown"
}

# Function to install radius
install_radius() {
    local version=$1
    echo "Installing Radius version: $version"
    
    # Use the radius installer script
    curl -fsSL https://raw.githubusercontent.com/radius-project/radius/main/deploy/install.sh | bash -s -- $version
    
    # Verify installation
    if ! command -v rad &> /dev/null; then
        echo "Failed to install Radius"
        exit 1
    fi
    
    echo "Successfully installed Radius $version"
}

# Get the latest release version
echo "Checking for latest Radius release..."
LATEST_RELEASE=$(curl -s $RELEASES_URL | jq -r '.[] | select(.prerelease==false) | .tag_name' | sort -V -r | head -n 1)

if [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to fetch latest release information"
    exit 1
fi

echo "Latest release: $LATEST_RELEASE"

# Check if rad is installed
if ! is_rad_installed; then
    echo "Radius CLI (rad) is not installed"
    install_radius $LATEST_RELEASE
    echo "NEW_RELEASE=true" >> $GITHUB_ENV
    echo "RADIUS_VERSION=$LATEST_RELEASE" >> $GITHUB_ENV
    mkdir -p $(dirname "$CACHE_FILE")
    echo "$LATEST_RELEASE" > "$CACHE_FILE"
    exit 0
fi

# Get current installed version
CURRENT_VERSION=$(get_current_rad_version)
echo "Current installed Radius version: $CURRENT_VERSION"

# Compare with latest release
if [ "$LATEST_RELEASE" != "v$CURRENT_VERSION" ]; then
    echo "Current version ($CURRENT_VERSION) differs from latest release ($LATEST_RELEASE)"
    install_radius $LATEST_RELEASE
    echo "NEW_RELEASE=true" >> $GITHUB_ENV
    echo "RADIUS_VERSION=$LATEST_RELEASE" >> $GITHUB_ENV
    mkdir -p $(dirname "$CACHE_FILE")
    echo "$LATEST_RELEASE" > "$CACHE_FILE"
else
    echo "Current version is already the latest: $CURRENT_VERSION"
    echo "NEW_RELEASE=false" >> $GITHUB_ENV
    echo "RADIUS_VERSION=$LATEST_RELEASE" >> $GITHUB_ENV
fi

# Keep this for compatibility with existing usage
if [ -f "$CACHE_FILE" ]; then
    LAST_TESTED=$(cat "$CACHE_FILE")
    echo "Last tested release: $LAST_TESTED"
    
    # Update cache file if needed
    if [ "$LATEST_RELEASE" != "$LAST_TESTED" ]; then
        echo "Updating last tested version in cache file"
        echo "$LATEST_RELEASE" > "$CACHE_FILE"
    fi
else
    mkdir -p $(dirname "$CACHE_FILE")
    echo "$LATEST_RELEASE" > "$CACHE_FILE"
fi