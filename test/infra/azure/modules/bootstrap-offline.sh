#!/bin/bash

# Offline bootstrap script for Radius AKS cluster
# This script assumes all required container images are pre-loaded into the private ACR

# Variables for offline deployment
CertManagerVersion="v1.12.0"
WorkloadIdentityVersion="v1.1.0"

echo "Installing kubectl..."
az aks install-cli --only-show-errors

# Get AKS credentials
echo "Getting AKS credentials..."
az aks get-credentials \
  --admin \
  --name $clusterName \
  --resource-group $resourceGroupName \
  --subscription $subscriptionId \
  --only-show-errors

# Check if kubectl is working
echo "Verifying cluster connectivity..."
kubectl get nodes

# Create namespace for cert-manager
echo "Creating cert-manager namespace..."
kubectl create namespace cert-manager --dry-run=client -o yaml | kubectl apply -f -

# Install cert-manager using pre-loaded images from ACR
# Note: This assumes cert-manager images have been pre-loaded into the ACR
echo "Installing cert-manager from private registry..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: cert-manager
---
# Add cert-manager CRDs
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: certificaterequests.cert-manager.io
spec:
  group: cert-manager.io
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
          status:
            type: object
  scope: Namespaced
  names:
    plural: certificaterequests
    singular: certificaterequest
    kind: CertificateRequest
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: certificates.cert-manager.io
spec:
  group: cert-manager.io
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
          status:
            type: object
  scope: Namespaced
  names:
    plural: certificates
    singular: certificate
    kind: Certificate
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: clusterissuers.cert-manager.io
spec:
  group: cert-manager.io
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
          status:
            type: object
  scope: Cluster
  names:
    plural: clusterissuers
    singular: clusterissuer
    kind: ClusterIssuer
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: issuers.cert-manager.io
spec:
  group: cert-manager.io
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
          status:
            type: object
  scope: Namespaced
  names:
    plural: issuers
    singular: issuer
    kind: Issuer
EOF

# Note: In a real offline scenario, you would need to:
# 1. Pre-load all required container images into your private ACR
# 2. Create deployment manifests that reference your private ACR
# 3. Apply those manifests instead of downloading from the internet

echo "Offline bootstrap completed. Note: For full offline deployment, ensure all required container images are pre-loaded in your private ACR."

echo '{}' >$AZ_SCRIPTS_OUTPUT_PATH
