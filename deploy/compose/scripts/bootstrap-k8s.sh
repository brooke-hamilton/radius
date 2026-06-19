#!/bin/sh
# ------------------------------------------------------------
# Copyright 2023 The Radius Authors.
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
#
# Prepares the k3s cluster for the Radius control plane:
#   1. Rewrites the k3s kubeconfig so the API server is reachable by the
#      Compose service DNS name "k3s", and publishes it to the shared volume
#      as "config" (the file the Radius binaries read from $HOME/.kube/config).
#   2. Creates the radius-system namespace and the radius-encryption-key secret
#      that dynamic-rp requires to start.
#   3. Installs the Radius CRDs.
#
# Idempotent: safe to run repeatedly. Mirrors the encryption-key bootstrap in
# build/scripts/start-radius.sh.
set -eu

RAW_KUBECONFIG="/work/kubeconfig.yaml"
OUT_KUBECONFIG="/work/config"

echo "Waiting for k3s to write its kubeconfig..."
while [ ! -f "${RAW_KUBECONFIG}" ]; do
  sleep 2
done

# Point the kubeconfig at the k3s service DNS name instead of 127.0.0.1.
sed -e 's#https://127.0.0.1:6443#https://k3s:6443#g' \
    -e 's#https://0.0.0.0:6443#https://k3s:6443#g' \
    "${RAW_KUBECONFIG}" > "${OUT_KUBECONFIG}"
chmod 0644 "${OUT_KUBECONFIG}"
export KUBECONFIG="${OUT_KUBECONFIG}"

echo "Waiting for the Kubernetes API to become ready..."
until kubectl get --raw='/readyz' >/dev/null 2>&1; do
  sleep 2
done
echo "Kubernetes API is ready."

# Namespace used by the control plane for secrets and the encryption key.
kubectl create namespace radius-system --dry-run=client -o yaml | kubectl apply -f -

# Create the encryption key secret that dynamic-rp loads at startup.
if ! kubectl -n radius-system get secret radius-encryption-key >/dev/null 2>&1; then
  echo "Creating radius-encryption-key secret..."
  ENC_KEY="$(head -c 32 /dev/urandom | base64 | tr -d '\n')"
  ENC_NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  # 90 days = 7776000 seconds. busybox `date` has no `-d '+90 days'`, so we
  # compute the expiry epoch and format it with `date -d @<epoch>`.
  ENC_EXP="$(date -u -d "@$(( $(date -u +%s) + 7776000 ))" +%Y-%m-%dT%H:%M:%SZ)"
  ENC_JSON="$(printf '{"currentVersion":1,"keys":{"1":{"key":"%s","version":1,"createdAt":"%s","expiresAt":"%s"}}}' \
    "${ENC_KEY}" "${ENC_NOW}" "${ENC_EXP}")"
  kubectl -n radius-system create secret generic radius-encryption-key \
    --from-literal=keys.json="${ENC_JSON}"
else
  echo "radius-encryption-key secret already present."
fi

# Install the Radius CRDs (radapp.io/* and ucp.dev/*).
if [ -d /crds ]; then
  echo "Applying Radius CRDs..."
  kubectl apply -R -f /crds
fi

echo "Kubernetes bootstrap complete."
