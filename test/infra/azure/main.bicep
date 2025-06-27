/*
Copyright 2023 The Radius Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

@description('Specifies the prefix for resource names deployed in this template.')
param prefix string = uniqueString(resourceGroup().id)

@description('Specifies the location where to deploy the resources. Default is the resource group location.')
param location string = resourceGroup().location

@description('Specifies the name of log analytics workspace. Default is {prefix}-workspace.')
param logAnalyticsWorkspaceName string = '${prefix}-workspace'

@description('Specifies the location of log analytics workspace. Default is the resource group location.')
param logAnalyticsWorkspaceLocation string = resourceGroup().location

@description('Specifies the location of azure monitor workspace. Default is {prefix}-azm-workspace.')
param azureMonitorWorkspaceName string = '${prefix}-azm-workspace'

@allowed([
  'eastus2euap'
  'centraluseuap'
  'centralus'
  'eastus'
  'eastus2'
  'northeurope'
  'southcentralus'
  'southeastasia'
  'uksouth'
  'westeurope'
  'westus'
  'westus2'
])
@description('Specifies the location of azure monitor workspace. Default is westus2')
param azureMonitorWorkspaceLocation string = 'westus2'

@description('Specifies the name of aks cluster. Default is {prefix}-aks.')
@minLength(1)
@maxLength(63)
param aksClusterName string = '${prefix}-aks'

@description('Enables Azure Monitoring and Grafana Dashboard. Default is false.')
param grafanaEnabled bool = false

@description('Specifies the object id to assign Grafana administrator role. Can be the object id of AzureAD user or group.')
param grafanaAdminObjectId string = ''

@description('Specifies the name of Grafana dashboard. Default is {prefix}-dashboard.')
param grafanaDashboardName string = '${prefix}-dashboard'

@description('Specifies whether to install the required tools for running Radius. Default is true.')
param installKubernetesDependencies bool = true

@description('Specifies whether the AKS cluster should be private. Default is true for offline environments.')
param privateClusterEnabled bool = true

@description('Specifies the name of the virtual network. Default is {prefix}-vnet.')
param virtualNetworkName string = '${prefix}-vnet'

@description('Specifies the address prefix of the virtual network. Default is 10.0.0.0/8.')
param virtualNetworkAddressPrefix string = '10.0.0.0/8'

@description('Specifies the name of the subnet for AKS nodes. Default is aks-subnet.')
param aksSubnetName string = 'aks-subnet'

@description('Specifies the address prefix of the AKS subnet. Default is 10.240.0.0/16.')
param aksSubnetAddressPrefix string = '10.240.0.0/16'

@description('Specifies the name of the subnet for private endpoints. Default is pe-subnet.')
param privateEndpointSubnetName string = 'pe-subnet'

@description('Specifies the address prefix of the private endpoint subnet. Default is 10.241.0.0/24.')
param privateEndpointSubnetAddressPrefix string = '10.241.0.0/24'

@description('Specifies the name of the Azure Container Registry. Default is {prefix}registry.')
@minLength(5)
@maxLength(50)
param acrName string = '${replace(prefix, '-', '')}registry'

@description('Specifies whether to create NAT Gateway for outbound connectivity. Default is true.')
param enableNatGateway bool = true

param defaultTags object = {
  radius: 'infra'
}

// Deploy Virtual Network for private connectivity
module virtualNetwork './modules/vnet.bicep' = {
  name: virtualNetworkName
  params: {
    name: virtualNetworkName
    location: location
    addressPrefix: virtualNetworkAddressPrefix
    aksSubnetName: aksSubnetName
    aksSubnetAddressPrefix: aksSubnetAddressPrefix
    privateEndpointSubnetName: privateEndpointSubnetName
    privateEndpointSubnetAddressPrefix: privateEndpointSubnetAddressPrefix
    enableNatGateway: enableNatGateway
    tags: defaultTags
  }
}

// Deploy Azure Container Registry for private image storage
module containerRegistry './modules/acr.bicep' = {
  name: acrName
  params: {
    name: acrName
    location: location
    sku: 'Premium'
    privateEndpointSubnetId: virtualNetwork.outputs.privateEndpointSubnetId
    vnetId: virtualNetwork.outputs.vnetId
    tags: defaultTags
  }
}

// Deploy Log Analytics Workspace for log.
module logAnalyticsWorkspace './modules/loganalytics-workspace.bicep' = {
  name: logAnalyticsWorkspaceName
  params: {
    name: logAnalyticsWorkspaceName
    location: logAnalyticsWorkspaceLocation
    sku: 'PerGB2018'
    retentionInDays: 30
    tags: defaultTags
  }
}

// Deploy Azure Monitor Workspace for metrics.
resource azureMonitorWorkspace 'microsoft.monitor/accounts@2023-04-03' = {
  name: azureMonitorWorkspaceName
  location: azureMonitorWorkspaceLocation
  properties: {}
}

// Deploy AKS cluster with OIDC Issuer profile and Dapr.
module aksCluster './modules/akscluster.bicep' = {
  name: aksClusterName
  params: {
    name: aksClusterName
    location: location
    kubernetesVersion: '1.31.8'
    logAnalyticsWorkspaceId: logAnalyticsWorkspace.outputs.id
    systemAgentPoolName: 'agentpool'
    systemAgentPoolVmSize: 'Standard_D4as_v5'
    systemAgentPoolAvailabilityZones: []
    systemAgentPoolOsDiskType: 'Managed'
    systemAgentPoolOsSKU: 'AzureLinux'
    userAgentPoolName: 'userpool'
    userAgentPoolVmSize: 'Standard_D8as_v5'
    userAgentPoolAvailabilityZones: []
    userAgentPoolMaxPods: 50
    userAgentPoolMinCount: 4
    userAgentPoolOsDiskType: 'Managed'
    userAgentPoolOsSKU: 'AzureLinux'
    daprEnabled: true
    daprHaEnabled: false
    oidcIssuerProfileEnabled: true
    workloadIdentityEnabled: true
    imageCleanerEnabled: true
    imageCleanerIntervalHours: 24
    // Private cluster configuration
    privateClusterEnabled: privateClusterEnabled
    vnetSubnetId: virtualNetwork.outputs.aksSubnetId
    outboundType: enableNatGateway ? 'userDefinedRouting' : 'loadBalancer'
    tags: defaultTags
  }
}

// Deploy data collection for log analytics.
module logAnalyticsDataCollection './modules/loganalytics-datacollection.bicep' = if (grafanaEnabled) {
  name: 'loganalytics-datacollection'
  params: {
    logAnalyticsWorkspaceId: logAnalyticsWorkspace.outputs.id
    logAnalyticsWorkspaceLocation: logAnalyticsWorkspace.outputs.location
    clusterResourceId: aksCluster.outputs.id
    clusterLocation: aksCluster.outputs.location
    tags: defaultTags
  }
}

// Deploy Grafana dashboard.
module grafanaDashboard './modules/grafana.bicep' = if (grafanaEnabled) {
  name: grafanaDashboardName
  params: {
    name: grafanaDashboardName
    location: location
    adminObjectId: grafanaAdminObjectId
    azureMonitorWorkspaceId: azureMonitorWorkspace.id
    clusterResourceId: aksCluster.outputs.id
    clusterLocation: aksCluster.outputs.location
    tags: defaultTags
  }
}

// Deploy data collection for metrics.
module dataCollection './modules/datacollection.bicep' = if (grafanaEnabled) {
  name: 'dataCollection'
  params: {
    azureMonitorWorkspaceLocation: azureMonitorWorkspace.location
    azureMonitorWorkspaceId: azureMonitorWorkspace.id
    clusterResourceId: aksCluster.outputs.id
    clusterLocation: aksCluster.outputs.location
    tags: defaultTags
  }
  dependsOn: [
    grafanaDashboard
  ]
}

// Deploy alert rules using prometheus metrics.
module alertManagement './modules/alert-management.bicep' = if (grafanaEnabled) {
  name: 'alertManagement'
  params: {
    azureMonitorWorkspaceLocation: azureMonitorWorkspace.location
    azureMonitorWorkspaceResourceId: azureMonitorWorkspace.id
    clusterResourceId: aksCluster.outputs.id
    tags: defaultTags
  }
  dependsOn: [
    dataCollection
  ]
}

// This is a workaround to get the AKS cluster resource created by aksCluster module
// Note: Accessing admin credentials may fail for private clusters during deployment
resource aks 'Microsoft.ContainerService/managedClusters@2023-10-01' existing = if (grafanaEnabled) {
  name: aksCluster.name
}

// Deploy configmap for prometheus metrics.
module promConfigMap './modules/ama-metrics-setting-configmap.bicep' = if (grafanaEnabled) {
  name: 'metrics-configmap'
  params: {
    kubeConfig: aks.listClusterAdminCredential().kubeconfigs[0].value
  }
  dependsOn: [
    aks, dataCollection, alertManagement
  ]
}

// Run deployment script to bootstrap the cluster for Radius.
module deploymentScript './modules/deployment-script-offline.bicep' = if (installKubernetesDependencies) {
  name: 'offlineDeploymentScript'
  params: {
    name: 'installKubernetesDependencies'
    clusterName: aksCluster.outputs.name
    resourceGroupName: resourceGroup().name
    subscriptionId: subscription().subscriptionId
    tenantId: subscription().tenantId
    location: location
    tags: defaultTags
  }
  dependsOn: [
    containerRegistry
  ]
}

module mongoDB './modules/mongodb.bicep' = {
  name: 'mongodb'
  params: {
    name: '${prefix}-mongodb'
    location: location
  }
}

output mongodbAccountID string = mongoDB.outputs.cosmosMongoAccountID
output aksControlPlaneFQDN string = aksCluster.outputs.controlPlaneFQDN
output grafanaDashboardFQDN string = grafanaEnabled ? grafanaDashboard.outputs.dashboardFQDN : ''
output acrLoginServer string = containerRegistry.outputs.loginServer
output acrName string = containerRegistry.outputs.name
output vnetId string = virtualNetwork.outputs.vnetId
output aksSubnetId string = virtualNetwork.outputs.aksSubnetId
