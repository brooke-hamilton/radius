# Offline/Air-gapped Deployment Guide for Radius on AKS

This guide provides detailed instructions for deploying Radius in a completely offline/air-gapped environment where there is no internet connectivity.

## Architecture Overview

The offline deployment creates:

- **Private AKS Cluster**: No public endpoints, API server accessible only through private network
- **Private Azure Container Registry (ACR)**: For hosting all required container images
- **Virtual Network**: Isolated network environment with controlled outbound access
- **NAT Gateway**: Optional controlled outbound connectivity for critical updates
- **Private DNS Zones**: For internal name resolution

## Pre-deployment Preparation

### 1. Container Image Preparation

Before deploying in an offline environment, you must prepare all required container images:

#### Required Images for Radius

```bash
# Core Radius images (update versions as needed)
radius/applications-rp:latest
radius/ucp:latest
radius/deployment-engine:latest

# Kubernetes dependencies
jetstack/cert-manager-controller:v1.12.0
jetstack/cert-manager-webhook:v1.12.0
jetstack/cert-manager-cainjector:v1.12.0
azure/aad-pod-identity/mic:v1.8.15
azure/aad-pod-identity/nmi:v1.8.15

# Dapr images (if enabled)
daprio/dapr:1.11.2
daprio/dapr-placement-server:1.11.2
daprio/dapr-sentry:1.11.2
daprio/dapr-sidecar-injector:1.11.2
```

#### Image Transfer Process

1. **From a connected environment**, pull and save images:

```bash
# Create a directory for images
mkdir -p offline-images

# Pull and save each image
docker pull jetstack/cert-manager-controller:v1.12.0
docker save jetstack/cert-manager-controller:v1.12.0 -o offline-images/cert-manager-controller.tar

# Repeat for all required images...
```

1. **Transfer images** to your offline environment using secure media

1. **In the offline environment**, load images into your private registry:

```bash
# Load images from files
docker load -i offline-images/cert-manager-controller.tar

# Tag for your private registry
docker tag jetstack/cert-manager-controller:v1.12.0 yourregistry.azurecr.io/cert-manager-controller:v1.12.0

# Push to private registry (must be done before full isolation)
docker push yourregistry.azurecr.io/cert-manager-controller:v1.12.0
```

## Deployment Steps

### 1. Network Connectivity Planning

Before deployment, plan your network connectivity:

- **Management Access**: How will you access the private cluster? (Bastion host, VPN, ExpressRoute)
- **Outbound Connectivity**: Do you need any outbound access? (NAT Gateway vs. completely isolated)
- **DNS Resolution**: How will internal DNS be handled?

### 2. Deploy Infrastructure

```bash
# Deploy with offline-specific parameters
az deployment group create --resource-group [Resource Group Name] --template-file main.bicep \
  --parameters \
  privateClusterEnabled=true \
  enableNatGateway=true \
  installKubernetesDependencies=false \
  grafanaEnabled=false \
  virtualNetworkAddressPrefix='10.0.0.0/8' \
  aksSubnetAddressPrefix='10.240.0.0/16' \
  privateEndpointSubnetAddressPrefix='10.241.0.0/24'
```

**Note**: For custom VNet deployments, Azure AKS requires either `loadBalancer` or `userDefinedRouting` outbound types. The template automatically configures:

- `userDefinedRouting` when `enableNatGateway=true` (creates empty route table, NAT Gateway handles routing)
- `loadBalancer` when `enableNatGateway=false` (uses Azure Load Balancer)

When using `userDefinedRouting`, an empty route table is created and associated with the AKS subnet. The NAT Gateway at the subnet level handles the actual outbound traffic routing. This approach provides controlled outbound connectivity while satisfying AKS requirements.

Key parameters for offline deployment:

- `enableNatGateway=true/false`: Controls outbound connectivity method
  - `true`: Creates NAT Gateway for controlled outbound access (recommended for partial connectivity)
  - `false`: Uses Load Balancer only (for completely isolated environments)
- `installKubernetesDependencies=false`: Skip online dependency installation
- `grafanaEnabled=false`: Reduce complexity in offline environment

### 3. Post-Deployment Configuration

#### Access the Private Cluster

1. **Set up management connectivity** (choose one):
   - Deploy a jump box VM in the same VNet
   - Configure VPN Gateway for remote access
   - Use Azure Bastion for secure access

2. **Configure kubectl**:

```bash
# From management host with VNet access
az aks get-credentials --resource-group [RG] --name [AKS] --admin
kubectl get nodes
```

#### Manual Installation of Dependencies

Since `installKubernetesDependencies=false`, you'll need to manually install:

1. **cert-manager** using images from your private ACR:

```yaml
# cert-manager-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager
  namespace: cert-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cert-manager
  template:
    metadata:
      labels:
        app: cert-manager
    spec:
      containers:
      - name: cert-manager
        image: yourregistry.azurecr.io/cert-manager-controller:v1.12.0
        # ... rest of configuration
```

1. **Configure image pull secrets** if needed:

```bash
kubectl create secret docker-registry acr-secret \
  --docker-server=yourregistry.azurecr.io \
  --docker-username=[ACR-USERNAME] \
  --docker-password=[ACR-PASSWORD] \
  --namespace=cert-manager
```

## Monitoring and Maintenance

### Health Checks

Regular health checks for offline environments:

```bash
# Check cluster health
kubectl get nodes
kubectl get pods --all-namespaces

# Check ACR connectivity
kubectl run test-acr --image=yourregistry.azurecr.io/test:latest --rm -it

# Check DNS resolution
kubectl run -it --rm debug --image=yourregistry.azurecr.io/busybox --restart=Never -- nslookup kubernetes.default
```

### Update Process

For updates in offline environments:

1. **Test updates** in a connected environment first
2. **Prepare new images** following the same offline process
3. **Transfer and load** new images
4. **Rolling updates** using kubectl or Helm

## Security Considerations

### Network Security

- All ingress/egress traffic should be monitored and controlled
- Regular security scanning of container images before import
- Network segmentation between different application tiers

### Image Security

- Implement image scanning pipeline before importing to offline registry
- Use signed images where possible
- Regular vulnerability assessment of stored images

### Access Control

- Strict RBAC policies for cluster access
- Regular rotation of certificates and credentials
- Audit logging for all administrative actions

## Troubleshooting

### Common Issues in Offline Environments

1. **Image Pull Failures**:
   - Verify image exists in private ACR
   - Check image pull secrets
   - Validate network connectivity to ACR private endpoint

2. **DNS Resolution Issues**:
   - Check private DNS zone configuration
   - Verify DNS forwarding rules
   - Test with `nslookup` from within cluster

3. **Certificate Issues**:
   - Ensure cert-manager is properly configured for offline operation
   - Check if certificate authorities are accessible
   - Consider using internal CA for offline scenarios

### Diagnostic Commands

```bash
# Network connectivity test
kubectl run network-test --image=yourregistry.azurecr.io/busybox -it --rm -- /bin/sh

# Check private endpoint status
az network private-endpoint list --resource-group [RG] --output table

# ACR health check
az acr check-health --name [ACR-NAME]

# DNS resolution test
kubectl run dns-test --image=yourregistry.azurecr.io/busybox -it --rm -- nslookup yourregistry.azurecr.io
```

## Maintenance Scripts

Consider creating maintenance scripts for routine operations:

```bash
#!/bin/bash
# check-offline-health.sh

echo "Checking AKS cluster health..."
kubectl get nodes

echo "Checking critical pods..."
kubectl get pods -n kube-system
kubectl get pods -n cert-manager

echo "Checking ACR connectivity..."
kubectl run acr-test --image=$ACR_NAME.azurecr.io/busybox --rm -it --restart=Never -- echo "ACR connectivity OK"

echo "Health check complete."
```

This guide ensures your Radius deployment can operate effectively in a completely offline environment while maintaining security and operational excellence.
