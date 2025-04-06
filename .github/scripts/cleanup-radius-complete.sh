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

# This script performs a complete uninstall and cleanup of Radius
# It should be used when we need to perform a clean install of a new version

set -e

echo "Performing complete Radius uninstall and cleanup..."

# Check if rad CLI exists and uninstall Radius if it does
if command -v rad &> /dev/null; then
    echo "Uninstalling Radius using rad CLI..."
    rad uninstall kubernetes --force || true
    echo "Waiting for resources to be deleted..."
    sleep 30
else
    echo "rad CLI not found, skipping rad uninstall"
fi

# Delete Radius system namespace
echo "Deleting Radius system namespace if it exists..."
kubectl delete namespace radius-system --ignore-not-found=true

# Find and delete all application namespaces created by Radius
echo "Finding and deleting application namespaces..."
NAMESPACES=$(kubectl get namespace -l radapp.io/environment-type -o jsonpath='{.items[*].metadata.name}')
if [ -n "$NAMESPACES" ]; then
    echo "Deleting application namespaces: $NAMESPACES"
    for ns in $NAMESPACES; do
        kubectl delete namespace $ns --ignore-not-found=true
    done
else
    echo "No application namespaces found"
fi

# Delete any remaining Radius CRDs
echo "Deleting any remaining Radius CRDs..."
kubectl delete crd -l radapp.io/name --ignore-not-found=true

# Delete any remaining Radius resources
echo "Deleting any remaining Radius resources..."
kubectl delete deployments -l radapp.io/name --all-namespaces --ignore-not-found=true
kubectl delete services -l radapp.io/name --all-namespaces --ignore-not-found=true
kubectl delete configmaps -l radapp.io/name --all-namespaces --ignore-not-found=true
kubectl delete secrets -l radapp.io/name --all-namespaces --ignore-not-found=true

echo "Radius cleanup completed"