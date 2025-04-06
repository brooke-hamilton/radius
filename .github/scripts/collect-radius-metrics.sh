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

# This script collects performance metrics from Radius components
# It focuses on memory usage, CPU usage, and pod restarts to detect memory leaks and performance issues

set -e

OUTPUT_DIR="${1:-./dist/metrics}"
NAMESPACE="${2:-radius-system}"
TIMESTAMP=$(date +%Y%m%d%H%M%S)

mkdir -p "$OUTPUT_DIR"

echo "Collecting Radius performance metrics..."
echo "Output directory: $OUTPUT_DIR"

# Get list of Radius pods
echo "Getting list of Radius pods..."
PODS=$(kubectl get pods -n $NAMESPACE -o jsonpath="{.items[*].metadata.name}")

if [ -z "$PODS" ]; then
    echo "No Radius pods found in namespace $NAMESPACE"
    exit 1
fi

# Collect resource usage metrics for each pod
echo "Collecting pod resource metrics..."
kubectl top pods -n $NAMESPACE > "$OUTPUT_DIR/pod_resources_$TIMESTAMP.txt"

# Collect detailed pod information (including restart counts)
echo "Collecting pod information..."
kubectl get pods -n $NAMESPACE -o wide > "$OUTPUT_DIR/pod_info_$TIMESTAMP.txt"

# Collect logs from each pod
echo "Collecting pod logs..."
for pod in $PODS; do
    echo "Collecting logs from $pod..."
    
    # Get containers in the pod
    CONTAINERS=$(kubectl get pods $pod -n $NAMESPACE -o jsonpath="{.spec.containers[*].name}")
    
    for container in $CONTAINERS; do
        echo "  - Container: $container"
        kubectl logs $pod -c $container -n $NAMESPACE > "$OUTPUT_DIR/${pod}_${container}_${TIMESTAMP}.log"
    done
done

# Collect pod descriptions (which include events)
echo "Collecting pod descriptions..."
for pod in $PODS; do
    kubectl describe pod $pod -n $NAMESPACE > "$OUTPUT_DIR/${pod}_description_${TIMESTAMP}.txt"
done

# Collect memory and CPU metrics for nodes
echo "Collecting node metrics..."
kubectl top nodes > "$OUTPUT_DIR/node_metrics_${TIMESTAMP}.txt"

echo "Metrics collection complete"