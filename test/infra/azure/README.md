# Build Radius infrastructure to Azure (Offline/Air-gapped Environment)

This directory includes the Bicep templates to deploy the following resources on Azure for running Radius in an offline/air-gapped network environment:

- Virtual Network with private subnets and NAT Gateway for outbound connectivity
- Azure Container Registry (ACR) with private endpoints for container image storage
- Log Analytics Workspace for log
- Azure Monitor Workspace for metric
- Private AKS Cluster
  - Deployed in private subnet with no public endpoints
  - Installed extensions: Azure Keyvault CSI driver, Dapr
- Grafana dashboard (optional)
- Network security groups for secure communication
- Private DNS zones for name resolution

## Key Features for Offline Environment

- **Private AKS Cluster**: API server is only accessible through private endpoints
- **Private Container Registry**: ACR with private endpoints for secure image storage
- **Network Isolation**: All components deployed in private subnets
- **Controlled Outbound Access**: Uses NAT Gateway for controlled internet access
- **Private DNS**: Custom DNS zones for internal name resolution

## Prerequisite

1. [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)
2. [Azure subscription](https://azure.com) to which you have a Owner/Contributor role
3. **Network connectivity**: You'll need a connection to your Azure environment (VPN, ExpressRoute, or Azure Bastion)
4. **Pre-loaded container images**: For a fully offline environment, you'll need to pre-load required container images into the ACR

## Steps

1. Log in to Azure and select your subscription:

   ```bash
   az login
   az account set -s [Subscription Id]
   ```

1. Enable `Microsoft.ContainerService/EnableImageCleanerPreview` feature flag

   This cleans up unused container images in each node, which can cause the security vulnerabilities. Visit <https://aka.ms/aks/image-cleaner> to learn more about image cleaner.

   ```bash
   # Check the feature flag to see if it is 'Registered'. If the status is 'Registered', you can skip this step.
   az feature show --namespace "Microsoft.ContainerService" --name "EnableImageCleanerPreview"
   {
     "id": "/subscriptions/<subscriptionid>/providers/Microsoft.Features/providers/Microsoft.ContainerService/features/EnableImageCleanerPreview",
     "name": "Microsoft.ContainerService/EnableImageCleanerPreview",
     "properties": {
       "state": "Registered"
     },
     "type": "Microsoft.Features/providers/features"
   }

   # Register feature flag.
   az feature register --namespace "Microsoft.ContainerService" --name "EnableImageCleanerPreview"

   # Ensure that the feature flag is 'Registered'.
   az feature show --namespace "Microsoft.ContainerService" --name "EnableImageCleanerPreview"

   # Re-register resource provider.
   az provider register --namespace Microsoft.ContainerService
   ```

   > Note: When you enable the feature flag first in your subscription, it will take some time to be propagated.

1. Create resource group:

   ```bash
   az group create --location [Location Name] --resource-group [Resource Group Name]
   ```

   - **[Location Name]**: Specify the location of the resource group. This location will be used as the default location for the resources in the template.
   - **[Resource Group Name]**: Provide a name for the resource group where the template will be deployed.

1. Deploy main.bicep:

   The template now includes parameters for configuring the private network environment:

   ```bash
   az deployment group create --resource-group [Resource Group Name] --template-file main.bicep \
     --parameters \
     grafanaEnabled=[Grafana Dashboard Enabled] \
     grafanaAdminObjectId='[Grafana Admin Object Id]' \
     privateClusterEnabled=true \
     enableNatGateway=true \
     virtualNetworkAddressPrefix='10.0.0.0/8' \
     aksSubnetAddressPrefix='10.240.0.0/16' \
     privateEndpointSubnetAddressPrefix='10.241.0.0/24'
   ```

   Key parameters for offline deployment:
   - **privateClusterEnabled**: Set to `true` for private AKS cluster (default: true)
   - **enableNatGateway**: Set to `true` to enable NAT Gateway for controlled outbound connectivity (default: true)
     - When `true`: Uses `userDefinedRouting` with custom NAT Gateway
     - When `false`: Uses `loadBalancer` for outbound connectivity
   - **virtualNetworkAddressPrefix**: Address space for the virtual network (default: 10.0.0.0/8)
   - **aksSubnetAddressPrefix**: Subnet for AKS nodes (default: 10.240.0.0/16)
   - **privateEndpointSubnetAddressPrefix**: Subnet for private endpoints (default: 10.241.0.0/24)
   - **[Grafana Dashboard Enabled]**: Set `true` if you want to see metrics and its dashboard with Azure managed Prometheus and Grafana dashboard. Otherwise, `false` is recommended to save the cost.
   - **[Grafana Admin Object Id]**: Set the object ID of the Grafana Admin user or group. To find the object id, search for the admin user or group name on [AAD Portal Overview search box](https://portal.azure.com/#view/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/~/Overview) and get the object id or run `az ad signed-in-user show` to get your own user object id.

## Post-Deployment Steps for Offline Environment

### 1. Access the Private AKS Cluster

Since the AKS cluster is private, you'll need to access it from within the virtual network or through a connection that has access to the private network:

```bash
# If you have a VM in the same VNet or connected network
az aks get-credentials --resource-group [Resource Group Name] --name [AKS Cluster Name] --admin

# Verify connectivity
kubectl get nodes
```

### 2. Pre-load Container Images (For Fully Offline Environment)

For a completely offline environment, you'll need to pre-load all required container images into your private ACR:

```bash
# Get ACR login server (from deployment output)
ACR_NAME="[Your ACR Name]"
az acr login --name $ACR_NAME

# Import required images (examples)
az acr import --name $ACR_NAME --source docker.io/jetstack/cert-manager-controller:v1.12.0 --image cert-manager-controller:v1.12.0
az acr import --name $ACR_NAME --source docker.io/jetstack/cert-manager-webhook:v1.12.0 --image cert-manager-webhook:v1.12.0
az acr import --name $ACR_NAME --source docker.io/jetstack/cert-manager-cainjector:v1.12.0 --image cert-manager-cainjector:v1.12.0

# Add any other required images for your Radius applications
```

### 3. Configure kubectl for Private Cluster Access

You may need to set up additional connectivity options:

- **Azure Bastion**: For secure RDP/SSH access to a jump box VM
- **VPN Gateway**: For site-to-site or point-to-site connectivity
- **ExpressRoute**: For dedicated private connectivity

## Network Architecture

The deployment creates the following network topology:

```text
Internet
    │
    ▼
┌─────────────┐
│ NAT Gateway │ (Outbound only)
└─────────────┘
    │
    ▼
┌───────────────────────────────────────┐
│          Virtual Network              │
│         (10.0.0.0/8)                  │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │     AKS Subnet                  │  │
│  │    (10.240.0.0/16)              │  │
│  │                                 │  │
│  │  ┌─────────────────────────┐    │  │
│  │  │   Private AKS Cluster   │    │  │
│  │  └─────────────────────────┘    │  │
│  └─────────────────────────────────┘  │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │  Private Endpoint Subnet        │  │
│  │   (10.241.0.0/24)               │  │
│  │                                 │  │
│  │  ┌─────────────────────────┐    │  │
│  │  │   ACR Private Endpoint  │    │  │
│  │  └─────────────────────────┘    │  │
│  └─────────────────────────────────┘  │
└───────────────────────────────────────┘
```

## Security Considerations

This offline deployment provides enhanced security through:

1. **Network Isolation**: AKS nodes are in private subnets with no direct internet access
2. **Private API Server**: Kubernetes API server is only accessible through private endpoints
3. **Private Container Registry**: ACR is only accessible through private endpoints
4. **Network Security Groups**: Restrict traffic between subnets
5. **Controlled Outbound Access**: Only necessary outbound traffic through NAT Gateway

## Troubleshooting

### Common Issues

1. **Cannot access AKS cluster**:
   - Ensure you're connecting from a network that has access to the private VNet
   - Check if kubectl is configured correctly with admin credentials

2. **Pod image pull failures**:
   - Verify ACR private endpoint is correctly configured
   - Ensure required images are available in the private ACR
   - Check if AKS has proper permissions to pull from ACR

3. **DNS resolution issues**:
   - Verify private DNS zones are correctly linked to the VNet
   - Check if custom DNS servers are properly configured

### Useful Commands

```bash
# Check AKS cluster connectivity
kubectl get nodes
kubectl get pods --all-namespaces

# Check ACR connectivity
az acr check-health --name [ACR_NAME]

# List images in ACR
az acr repository list --name [ACR_NAME]

# Check private endpoint status
az network private-endpoint list --resource-group [Resource Group Name]
```

## Cost Optimization

For development/testing environments, consider:

- Setting `grafanaEnabled=false` to avoid Grafana costs
- Using smaller VM sizes for AKS nodes
- Implementing auto-scaling to reduce costs during idle periods
- Using Azure Spot instances for non-production workloads (if supported in your scenario)

## References

- [Azure Private AKS Cluster](https://docs.microsoft.com/en-us/azure/aks/private-clusters)
- [Azure Container Registry Private Endpoints](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-private-link)
- [Azure NAT Gateway](https://docs.microsoft.com/en-us/azure/virtual-network/nat-gateway/nat-overview)
- [Azure Private DNS Zones](https://docs.microsoft.com/en-us/azure/dns/private-dns-overview)
