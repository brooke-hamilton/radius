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

@description('Specifies the name of the virtual network.')
param name string

@description('Specifies the location of the virtual network.')
param location string = resourceGroup().location

@description('Specifies the address prefix of the virtual network.')
param addressPrefix string = '10.0.0.0/8'

@description('Specifies the name of the subnet for AKS nodes.')
param aksSubnetName string = 'aks-subnet'

@description('Specifies the address prefix of the AKS subnet.')
param aksSubnetAddressPrefix string = '10.240.0.0/16'

@description('Specifies the name of the subnet for private endpoints.')
param privateEndpointSubnetName string = 'pe-subnet'

@description('Specifies the address prefix of the private endpoint subnet.')
param privateEndpointSubnetAddressPrefix string = '10.241.0.0/24'

@description('Specifies whether to create NAT Gateway for outbound connectivity.')
param enableNatGateway bool = true

@description('Specifies the resource tags.')
param tags object = {}

// NAT Gateway Public IP
resource natGatewayPublicIP 'Microsoft.Network/publicIPAddresses@2023-06-01' = if (enableNatGateway) {
  name: '${name}-natgw-pip'
  location: location
  tags: tags
  sku: {
    name: 'Standard'
    tier: 'Regional'
  }
  properties: {
    publicIPAllocationMethod: 'Static'
    publicIPAddressVersion: 'IPv4'
    idleTimeoutInMinutes: 4
  }
}

// NAT Gateway
resource natGateway 'Microsoft.Network/natGateways@2023-06-01' = if (enableNatGateway) {
  name: '${name}-natgw'
  location: location
  tags: tags
  sku: {
    name: 'Standard'
  }
  properties: {
    publicIpAddresses: [
      {
        id: natGatewayPublicIP.id
      }
    ]
    idleTimeoutInMinutes: 4
  }
}

// Route Table for AKS subnet (required for userDefinedRouting)
resource aksRouteTable 'Microsoft.Network/routeTables@2023-06-01' = if (enableNatGateway) {
  name: '${aksSubnetName}-rt'
  location: location
  tags: tags
  properties: {
    routes: []
    disableBgpRoutePropagation: false
  }
}

// Network Security Group for AKS subnet
resource aksNsg 'Microsoft.Network/networkSecurityGroups@2023-06-01' = {
  name: '${aksSubnetName}-nsg'
  location: location
  tags: tags
  properties: {
    securityRules: [
      {
        name: 'AllowAKSInternalTraffic'
        properties: {
          protocol: '*'
          sourcePortRange: '*'
          destinationPortRange: '*'
          sourceAddressPrefix: aksSubnetAddressPrefix
          destinationAddressPrefix: aksSubnetAddressPrefix
          access: 'Allow'
          priority: 100
          direction: 'Inbound'
        }
      }
      {
        name: 'AllowAzureLoadBalancer'
        properties: {
          protocol: '*'
          sourcePortRange: '*'
          destinationPortRange: '*'
          sourceAddressPrefix: 'AzureLoadBalancer'
          destinationAddressPrefix: '*'
          access: 'Allow'
          priority: 200
          direction: 'Inbound'
        }
      }
      {
        name: 'DenyAllInbound'
        properties: {
          protocol: '*'
          sourcePortRange: '*'
          destinationPortRange: '*'
          sourceAddressPrefix: '*'
          destinationAddressPrefix: '*'
          access: 'Deny'
          priority: 1000
          direction: 'Inbound'
        }
      }
    ]
  }
}

// Network Security Group for Private Endpoints subnet
resource peNsg 'Microsoft.Network/networkSecurityGroups@2023-06-01' = {
  name: '${privateEndpointSubnetName}-nsg'
  location: location
  tags: tags
  properties: {
    securityRules: [
      {
        name: 'AllowVnetInbound'
        properties: {
          protocol: '*'
          sourcePortRange: '*'
          destinationPortRange: '*'
          sourceAddressPrefix: 'VirtualNetwork'
          destinationAddressPrefix: 'VirtualNetwork'
          access: 'Allow'
          priority: 100
          direction: 'Inbound'
        }
      }
      {
        name: 'DenyAllInbound'
        properties: {
          protocol: '*'
          sourcePortRange: '*'
          destinationPortRange: '*'
          sourceAddressPrefix: '*'
          destinationAddressPrefix: '*'
          access: 'Deny'
          priority: 1000
          direction: 'Inbound'
        }
      }
    ]
  }
}

// Virtual Network
resource virtualNetwork 'Microsoft.Network/virtualNetworks@2023-06-01' = {
  name: name
  location: location
  tags: tags
  properties: {
    addressSpace: {
      addressPrefixes: [
        addressPrefix
      ]
    }
    subnets: [
      {
        name: aksSubnetName
        properties: {
          addressPrefix: aksSubnetAddressPrefix
          networkSecurityGroup: {
            id: aksNsg.id
          }
          routeTable: enableNatGateway ? {
            id: aksRouteTable.id
          } : null
          natGateway: enableNatGateway ? {
            id: natGateway.id
          } : null
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
        }
      }
      {
        name: privateEndpointSubnetName
        properties: {
          addressPrefix: privateEndpointSubnetAddressPrefix
          networkSecurityGroup: {
            id: peNsg.id
          }
          privateEndpointNetworkPolicies: 'Disabled'
          privateLinkServiceNetworkPolicies: 'Disabled'
        }
      }
    ]
  }
}

// Outputs
output vnetId string = virtualNetwork.id
output vnetName string = virtualNetwork.name
output aksSubnetId string = resourceId('Microsoft.Network/virtualNetworks/subnets', virtualNetwork.name, aksSubnetName)
output privateEndpointSubnetId string = resourceId('Microsoft.Network/virtualNetworks/subnets', virtualNetwork.name, privateEndpointSubnetName)
output natGatewayId string = enableNatGateway ? natGateway.id : ''
output routeTableId string = enableNatGateway ? aksRouteTable.id : ''
