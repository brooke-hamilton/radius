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

# Get the latest release version
echo "Checking for latest Radius release..."
LATEST_RELEASE=$(curl -s $RELEASES_URL | jq -r '.[] | select(.prerelease==false) | .tag_name' | sort -V -r | head -n 1)

if [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to fetch latest release information"
    exit 1
fi

echo "Latest release: $LATEST_RELEASE"

# If no cache file exists, this is the first run
if [ ! -f "$CACHE_FILE" ]; then
    echo "No previously tested release found."
    echo "NEW_RELEASE=true" >> $GITHUB_ENV
    echo "RADIUS_VERSION=$LATEST_RELEASE" >> $GITHUB_ENV
    mkdir -p $(dirname "$CACHE_FILE")
    echo "$LATEST_RELEASE" > "$CACHE_FILE"
    exit 0
fi

# Read the previously tested version
LAST_TESTED=$(cat "$CACHE_FILE")
echo "Last tested release: $LAST_TESTED"

# Compare versions
if [ "$LATEST_RELEASE" != "$LAST_TESTED" ]; then
    echo "New release detected: $LATEST_RELEASE (previously tested: $LAST_TESTED)"
    echo "NEW_RELEASE=true" >> $GITHUB_ENV
    echo "RADIUS_VERSION=$LATEST_RELEASE" >> $GITHUB_ENV
    echo "$LATEST_RELEASE" > "$CACHE_FILE"
else
    echo "No new release detected. Continuing with version: $LAST_TESTED"
    echo "NEW_RELEASE=false" >> $GITHUB_ENV
    echo "RADIUS_VERSION=$LATEST_RELEASE" >> $GITHUB_ENV
fi