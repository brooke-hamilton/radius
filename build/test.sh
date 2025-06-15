#!/bin/bash

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

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if Azure CLI is installed and authenticated
check_azure_cli() {
    if ! command -v az &> /dev/null; then
        print_error "Azure CLI (az) is required but not installed."
        exit 1
    fi

    # Check if user is authenticated
    if ! az account show &> /dev/null; then
        print_error "Azure CLI is not authenticated."
        exit 1
    fi
}

# All environment variables required for the tests
get_env_vars() {

    local env_vars=(
        "AZURE_SUBSCRIPTION_ID"
        "AZURE_SP_TESTS_APPID"
        "AZURE_SP_TESTS_TENANTID"
        "TEST_BICEP_TYPES_REGISTRY"
        "RAD_VERSION"
        "BICEP_RECIPE_TAG_VERSION"
        "BICEP_RECIPE_REGISTRY"
        "TEST_RESOURCE_GROUP"
        "TEST_RESOURCE_GROUP_LOCATION"
    )

    printf '%s\n' "${env_vars[@]}"
}

set_env_defaults(){
    # Set default values for environment variables if not already set
    export AZURE_SUBSCRIPTION_ID="${AZURE_SUBSCRIPTION_ID:-$(az account show --query id -o tsv)}"
    export AZURE_SP_TESTS_APPID="${AZURE_SP_TESTS_APPID:-960d45e2-3636-46f7-ac3a-57262dc5e9c5}"
    export AZURE_SP_TESTS_TENANTID="${AZURE_SP_TESTS_TENANTID:-$(az account show --query tenantId -o tsv)}"
    export TEST_BICEP_TYPES_REGISTRY="${TEST_BICEP_TYPES_REGISTRY:-testuserdefinedbiceptypes.azurecr.io}"
    if [[ -n "${RAD_VERSION:-}" ]]; then
        export BICEP_RECIPE_TAG_VERSION="${BICEP_RECIPE_TAG_VERSION:-$RAD_VERSION}" # Use the RAD version as the tag for recipes by default
    fi
    export BICEP_RECIPE_REGISTRY="${BICEP_RECIPE_REGISTRY:-ghcr.io/radius-project/dev/test/recipes}"
    
    # Resource group where tests will deploy resources
    local id="${GITHUB_RUN_NUMBER:-$(whoami)}"
    TEST_RESOURCE_GROUP="${TEST_RESOURCE_GROUP:-$(whoami)-test-$(echo "$id" | sha1sum | head -c 5)}"
    export TEST_RESOURCE_GROUP
    export TEST_RESOURCE_GROUP_LOCATION="${TEST_RESOURCE_GROUP_LOCATION:-westus3}"
}

# Create a test.env file with the required environment variables from the get_env_vars function
create_env_file() {
    print_info " Creating test.env file with default values. Check the file contents to ensure accuracy."
    set_env_defaults
    local env_file_path="./build/test.env"
    local env_vars
    mapfile -t env_vars < <(get_env_vars)
    {
        echo "# Environment variables for tests"

        for var in "${env_vars[@]}"; do
            echo "$var=\"${!var:-}\""
        done
    } > $env_file_path || {
        print_error "Failed to create test.env file."
        exit 1
    }

    print_success "test.env file created successfully"
    cat $env_file_path
}

# Load environment variables from the test.env file if it exists.
# Validate that all required environment variables have values.
load_environment() {
    
    # Source the test environment file if it exists
    if [[ -f "./build/test.env" ]]; then
        # shellcheck source=/dev/null
        source "./build/test.env"
    fi

    # Validate that all required environment variables have values
    print_info "Validating environment variables..."
    local env_vars
    mapfile -t env_vars < <(get_env_vars)

    local missing_vars=()
    for var in "${env_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            print_error "$var is not set or is empty"
            missing_vars+=("$var")
        else
            print_success "$var is set"
        fi
    done
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        print_error "Missing ${#missing_vars[@]} required environment variable(s). Exiting."
        exit 1
    fi

    print_success "All required environment variables are properly set"
}

# Deploy long-running test cluster
deploy_aks_cluster() {

    load_environment

    # Install latest Radius CLI release
    make install-latest

    print_info "Deploying long-running test cluster to Azure..."
    
    # Default values (can be overridden by environment variables)
    local location="${LRT_AZURE_LOCATION:-westus3}"
    local resource_group="${LRT_RG:-$(whoami)-radius-lrt}"

    print_info "Configuration:"
    print_info "  Subscription: $(az account show --query "[name,id]" --output tsv | paste -sd,)"
    print_info "  Resource Group: $resource_group"
    print_info "  Location: $location"
    
    local feature_state
    feature_state=$(az feature show --namespace "Microsoft.ContainerService" --name "EnableImageCleanerPreview" --query properties.state -o tsv 2>/dev/null || echo "NotRegistered")
    if [[ "$feature_state" != "Registered" ]]; then
        print_warning "Feature flag is not registered. Registering now..."
        az feature register --namespace "Microsoft.ContainerService" --name "EnableImageCleanerPreview"
        az provider register --namespace Microsoft.ContainerService
    fi
    
    # Check if resource group exists and delete it if it does
    if az group exists --name "$resource_group" --output tsv | grep -q "true"; then
        print_warning "Resource group '$resource_group' already exists. Deleting it..."
        az group delete --name "$resource_group" --yes
        print_success "Resource group '$resource_group' deleted successfully."
    fi

    print_info "Creating resource group '$resource_group' in location '$location'..."
    az group create --location "$location" --resource-group "$resource_group" --output none
    print_success "Resource group created successfully."
    
    local template_file="./test/infra/azure/main.bicep"
    print_info "Deploying Bicep template..."
    if az deployment group create \
        --resource-group "$resource_group" \
        --template-file "$template_file"; then
        print_success "Deployment completed successfully!"
        
        # Connect to the AKS cluster
        local aks_cluster
        aks_cluster=$(az aks list --resource-group "$resource_group" --query '[0].name' -o tsv 2>/dev/null || echo "")
        az aks get-credentials --resource-group "$resource_group" --name "$aks_cluster" --admin --overwrite-existing --output none

        if [[ -n "$aks_cluster" ]]; then
        # Connect to cluster
            print_success "AKS cluster created and connected: '$aks_cluster'."
            print_info "To connect to the cluster, run:"
            print_info "  az aks get-credentials --resource-group $resource_group --name $aks_cluster --admin"
            print_info "To delete the cluster, run:"
            print_info "  az group delete --name $resource_group --yes --no-wait"
        fi
    else
        print_error "Deployment failed!"
        exit 1
    fi

    print_info "Publishing Bicep types to registry..."
    rad bicep publish-extension -f ./test/functional-portable/dynamicrp/noncloud/resources/testdata/testresourcetypes.yaml --target types.tgz --force
    make publish-test-bicep-recipes

    print_info "Creating resource group for tests '$TEST_RESOURCE_GROUP' in location '$TEST_RESOURCE_GROUP_LOCATION'..."
    az group create --name "$TEST_RESOURCE_GROUP" --location "$TEST_RESOURCE_GROUP_LOCATION" --output none

    print_info "Installing Radius CLI..."
    rad install kubernetes --reinstall
    
    # Ensure that all pods are running before proceeding
    kubectl wait --for=condition=available --all deployments --namespace radius-system --timeout=120s

    rad workspace create kubernetes --force
    rad group create default
    rad group switch default
    rad env create default --namespace default
    rad env switch default
    rad env update default --azure-subscription-id "$AZURE_SUBSCRIPTION_ID" --azure-resource-group "$TEST_RESOURCE_GROUP"
    rad credential register azure wi --client-id "$AZURE_SP_TESTS_APPID" --tenant-id "$AZURE_SP_TESTS_TENANTID"

    # rad env update ${{ env.RADIUS_TEST_ENVIRONMENT_NAME }} --aws-region ${{ env.AWS_REGION }} --aws-account-id ${{ secrets.FUNCTEST_AWS_ACCOUNT_ID }}
    # rad credential register aws access-key \
    #     --access-key-id ${{ secrets.FUNCTEST_AWS_ACCESS_KEY_ID }} --secret-access-key ${{ secrets.FUNCTEST_AWS_SECRET_ACCESS_KEY }}

    make publish-test-terraform-recipes

    # FUNC_TEST_OIDC_ISSUER not used
    # export FUNCTEST_OIDC_ISSUER=$(az aks show -n $AKS_CLUSTER_NAME -g $AZURE_RESOURCE_GROUP --query "oidcIssuerProfile.issuerUrl" -otsv)

    # Restore Radius Bicep types
    bicep restore ./test/functional-portable/corerp/cloud/resources/testdata/corerp-azure-connection-database-service.bicep --force
    # Restore AWS Bicep types 
    bicep restore ./test/functional-portable/corerp/cloud/resources/testdata/aws-s3-bucket.bicep --force
    make install-flux
    make install-gitea
}

run_lrt() {
    
    print_info "Running long-running tests against the AKS cluster..."
    
    load_environment

    # make test-functional-all
    #make test-functional-ucp
    # make test-functional-kubernetes
    
    # Time out and one failure
    # make test-functional-corerp
    
    # No return
    #make test-functional-cli
    
    # 4 tests, 4 failures
    #make test-functional-msgrp
    
    # 8 failures
    #make test-functional-daprrp
    
    # 6 failures
    #make test-functional-datastoresrp
    
    # Needs path to samples repo
    #make test-functional-samples
    
    # 4 failures
    #make test-functional-dynamicrp-noncloud

    #Delete the resource group after tests
    #print_info "Deleting resource group '$TEST_RESOURCE_GROUP'..."
    #az group delete --name "$TEST_RESOURCE_GROUP" --yes --no-wait
}

tear_down_aks_cluster() {
    print_info "Tearing down AKS cluster and resource group..."
    
    load_environment

    if az group exists --name "$TEST_RESOURCE_GROUP" --output tsv | grep -q "true"; then
        print_info "Deleting resource group '$TEST_RESOURCE_GROUP'..."
        az group delete --name "$TEST_RESOURCE_GROUP" --yes --no-wait
        print_success "Resource group '$TEST_RESOURCE_GROUP' deleted successfully."
    else
        print_warning "Resource group '$TEST_RESOURCE_GROUP' does not exist."
    fi
}

main() {
    if [[ $# -eq 0 ]]; then
        print_error "No command specified."
        local available_commands="create-env-file, deploy-aks-cluster, run-lrt"
        print_info "Available commands: $available_commands"
        exit 1
    fi
    
    local command="$1"
    check_azure_cli
    
    case "$command" in
        "load-environment")
            load_environment
            ;;
        "create-env-file")
            create_env_file
            ;;
        "deploy-aks-cluster")
            deploy_aks_cluster
            ;;
        "run-lrt")
            run_lrt
            ;;
        *)
            print_error "Unknown command: $command"
            print_info "Available commands: $available_commands"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
