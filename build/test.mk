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

##@ Test

# Will be set by our build workflow, this is just a default
TEST_TIMEOUT ?=1h
RADIUS_CONTAINER_LOG_PATH ?=./dist/container_logs
REL_VERSION ?=latest
DOCKER_REGISTRY ?=ghcr.io/radius-project/dev
ENVTEST_ASSETS_DIR=$(shell pwd)/bin
K8S_VERSION=1.30.*
ENV_SETUP=$(GOBIN)/setup-envtest$(BINARY_EXT)

# Use gotestsum if available, otherwise use go test. We want to enable testing with just 'make test'
# without external dependencies, but want to use gotestsum in our CI pipelines for the improved
# reporting.
#
# See: https://github.com/gotestyourself/gotestsum
#
# Gotestsum is a drop-in replacement for go test, but it provides a much nicer formatted output
# and it can also generate JUnit XML reports.
ifeq (, $(shell which gotestsum))
GOTEST_TOOL ?= go test
else
# Use these options by default but allow an override via env-var
GOTEST_OPTS ?=
# We need the double dash here to separate the 'gotestsum' options from the 'go test' options
GOTEST_TOOL ?= gotestsum $(GOTESTSUM_OPTS) --
endif

.PHONY: test
test: test-get-envtools ## Runs unit tests, excluding kubernetes controller tests
	KUBEBUILDER_ASSETS="$(shell $(ENV_SETUP) use -p path ${K8S_VERSION} --arch amd64)" CGO_ENABLED=1 $(GOTEST_TOOL) -v ./pkg/... $(GOTEST_OPTS)

.PHONY: test-get-envtools
test-get-envtools:
	@echo "$(ARROW) Installing Kubebuilder test tools..."
	$(call go-install-tool,$(ENV_SETUP),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)
	@echo "$(ARROW) Instructions:"
	@echo "$(ARROW) Set environment variable KUBEBUILDER_ASSETS for tests."
	@echo "$(ARROW) KUBEBUILDER_ASSETS=\"$(shell $(ENV_SETUP) use -p path ${K8S_VERSION} --arch amd64)\""

.PHONY: test-validate-cli
test-validate-cli: ## Run cli integration tests
	CGO_ENABLED=1 $(GOTEST_TOOL) -coverpkg= ./pkg/cli/cmd/... ./cmd/rad/... -timeout ${TEST_TIMEOUT} -v -parallel 5 $(GOTEST_OPTS)

test-functional-all: test-functional-ucp test-functional-kubernetes test-functional-corerp test-functional-cli test-functional-msgrp test-functional-daprrp test-functional-datastoresrp test-functional-samples ## Runs all functional tests

# Run all functional tests that do not require cloud resources
test-functional-all-noncloud: test-functional-ucp-noncloud test-functional-kubernetes-noncloud test-functional-corerp-noncloud test-functional-cli-noncloud test-functional-msgrp-noncloud test-functional-daprrp-noncloud test-functional-datastoresrp-noncloud test-functional-samples-noncloud ## Runs all functional tests that do not require cloud resources

# Run all functional tests that require cloud resources
test-functional-all-cloud: test-functional-ucp-cloud test-functional-corerp-cloud

test-functional-ucp: test-functional-ucp-noncloud test-functional-ucp-cloud ## Runs all UCP functional tests (both cloud and non-cloud)

test-functional-ucp-noncloud: ## Runs UCP functional tests that do not require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/ucp/noncloud/... -timeout ${TEST_TIMEOUT} -v -parallel 5 $(GOTEST_OPTS)

test-functional-ucp-cloud: ## Runs UCP functional tests that require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/ucp/cloud/... -timeout ${TEST_TIMEOUT} -v -parallel 5 $(GOTEST_OPTS)

test-functional-kubernetes: test-functional-kubernetes-noncloud ## Runs all Kubernetes functional tests
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/kubernetes/... -timeout ${TEST_TIMEOUT} -v -parallel 5 $(GOTEST_OPTS)

test-functional-kubernetes-noncloud: ## Runs Kubernetes functional tests that do not require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/kubernetes/noncloud/... -timeout ${TEST_TIMEOUT} -v -parallel 5 $(GOTEST_OPTS)

test-functional-corerp: test-functional-corerp-noncloud test-functional-corerp-cloud ## Runs all Core RP functional tests (both cloud and non-cloud)

test-functional-corerp-noncloud: ## Runs corerp functional tests that do not require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/corerp/noncloud/... -timeout ${TEST_TIMEOUT} -v -parallel 10 $(GOTEST_OPTS)

test-functional-corerp-cloud: ## Runs corerp functional tests that require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/corerp/cloud/... -timeout ${TEST_TIMEOUT} -v -parallel 10 $(GOTEST_OPTS)

test-functional-msgrp: test-functional-msgrp-noncloud ## Runs all Messaging RP functional tests (both cloud and non-cloud)

test-functional-msgrp-noncloud: ## Runs Messaging RP functional tests that do not require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/messagingrp/noncloud/... -timeout ${TEST_TIMEOUT} -v -parallel 2 $(GOTEST_OPTS)

test-functional-cli: test-functional-cli-noncloud ## Runs all cli functional tests (both cloud and non-cloud)

test-functional-cli-noncloud: ## Runs cli functional tests that do not require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/cli/noncloud/... -timeout ${TEST_TIMEOUT} -v -parallel 10 $(GOTEST_OPTS)

test-functional-daprrp: test-functional-daprrp-noncloud ## Runs all Dapr RP functional tests (both cloud and non-cloud)

test-functional-daprrp-noncloud: ## Runs Dapr RP functional tests that do not require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/daprrp/noncloud/... -timeout ${TEST_TIMEOUT} -v -parallel 3 $(GOTEST_OPTS)

test-functional-datastoresrp: test-functional-datastoresrp-noncloud ## Runs all Datastores RP functional tests (non-cloud)

test-functional-datastoresrp-noncloud: ## Runs Datastores RP functional tests that do not require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/datastoresrp/noncloud/... -timeout ${TEST_TIMEOUT} -v -parallel 3 $(GOTEST_OPTS)

test-functional-samples: test-functional-samples-noncloud ## Runs all Samples functional tests

test-functional-samples-noncloud: ## Runs Samples functional tests that do not require cloud resources
	CGO_ENABLED=1 $(GOTEST_TOOL) ./test/functional-portable/samples/noncloud/... -timeout ${TEST_TIMEOUT} -v -parallel 5 $(GOTEST_OPTS)

test-validate-bicep: ## Validates that all .bicep files compile cleanly
	BICEP_PATH="${HOME}/.rad/bin/rad-bicep" ./build/validate-bicep.sh

.PHONY: oav-installed
oav-installed:
	@echo "$(ARROW) Detecting oav (https://github.com/Azure/oav)..."
	@which oav > /dev/null || { echo "run 'npm install -g oav' to install oav"; exit 1; }
	@echo "$(ARROW) OK"

# TODO re-enable https://github.com/radius-project/radius/issues/5091
.PHONY: test-ucp-spec-examples 
test-ucp-spec-examples: oav-installed ## Validates UCP examples conform to UCP OpenAPI Spec
	# @echo "$(ARROW) Testing x-ms-examples conform to ucp spec..."
	# oav validate-example swagger/specification/ucp/resource-manager/UCP/preview/2023-10-01-preview/openapi.json

##@ Functional Test Environment Setup

.PHONY: setup-functional-test-env
setup-functional-test-env: ## Setup environment for functional tests
	@echo "$(ARROW) Setting up functional test environment..."
	$(eval TEMP_CERT_DIR := $(shell mktemp -d))
	@echo "Created temporary certificate directory: $(TEMP_CERT_DIR)"

.PHONY: create-local-registry
create-local-registry: ## Create a local Docker registry for testing
	@echo "$(ARROW) Creating secure local registry..."
	@mkdir -p $(TEMP_CERT_DIR)/certs/$(LOCAL_REGISTRY_SERVER)
	@openssl req -x509 -newkey rsa:4096 -days 1 -nodes \
		-subj "/CN=$(LOCAL_REGISTRY_SERVER)" \
		-out $(TEMP_CERT_DIR)/certs/$(LOCAL_REGISTRY_SERVER)/client.crt \
		-keyout $(TEMP_CERT_DIR)/certs/$(LOCAL_REGISTRY_SERVER)/client.key

	@docker network create radius-test-network || true
	@docker run -d --name $(LOCAL_REGISTRY_NAME) \
		--network=radius-test-network \
		-p $(LOCAL_REGISTRY_PORT):5000 \
		-v $(TEMP_CERT_DIR)/certs:/certs \
		-e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/$(LOCAL_REGISTRY_SERVER)/client.crt \
		-e REGISTRY_HTTP_TLS_KEY=/certs/$(LOCAL_REGISTRY_SERVER)/client.key \
		registry:2

.PHONY: create-kind-cluster
create-kind-cluster: ## Create a KinD cluster with a local registry
	@echo "$(ARROW) Creating KinD cluster with local registry..."
	@kind create cluster --name radius-test --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraMounts:
  - hostPath: $(TEMP_CERT_DIR)/certs
    containerPath: /etc/docker/certs.d/$(LOCAL_REGISTRY_SERVER):$(LOCAL_REGISTRY_PORT)
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."$(LOCAL_REGISTRY_SERVER):$(LOCAL_REGISTRY_PORT)"]
    endpoint = ["https://$(LOCAL_REGISTRY_SERVER):$(LOCAL_REGISTRY_PORT)"]
EOF
	@kubectl cluster-info --context kind-radius-test

.PHONY: install-radius-for-test
install-radius-for-test: ## Install Radius in the test cluster
	@echo "$(ARROW) Installing Radius to Kubernetes..."
	rad install kubernetes \
		--chart $(RADIUS_CHART_LOCATION) \
		--set rp.image=$(LOCAL_REGISTRY_NAME):$(LOCAL_REGISTRY_PORT)/applications-rp,rp.tag=$(REL_VERSION) \
		--set dynamicrp.image=$(LOCAL_REGISTRY_NAME):$(LOCAL_REGISTRY_PORT)/dynamic-rp,dynamicrp.tag=$(REL_VERSION) \
		--set controller.image=$(LOCAL_REGISTRY_NAME):$(LOCAL_REGISTRY_PORT)/controller,controller.tag=$(REL_VERSION) \
		--set ucp.image=$(LOCAL_REGISTRY_NAME):$(LOCAL_REGISTRY_PORT)/ucpd,ucp.tag=$(REL_VERSION) \
		--set de.image=$(DE_IMAGE),de.tag=$(DE_TAG) \
		--set-file global.rootCA.cert=$(TEMP_CERT_DIR)/certs/$(LOCAL_REGISTRY_SERVER)/client.crt
	
	@echo "$(ARROW) Creating workspace, group and environment for test..."
	rad workspace create kubernetes
	rad group create kind-radius
	rad group switch kind-radius
	rad env create kind-radius --namespace default
	rad env switch kind-radius

.PHONY: setup-test-recipes
setup-test-recipes: ## Set up test recipes for functional tests
	@echo "$(ARROW) Publishing test recipes..."
	make publish-test-terraform-recipes
	make publish-test-bicep-recipes

.PHONY: generate-test-bicepconfig
generate-test-bicepconfig: ## Generate bicepconfig.json for testing
	@echo "$(ARROW) Generating test bicepconfig.json..."
	@if [[ "$(REL_VERSION)" == "edge" ]]; then \
		RADIUS_VERSION="latest"; \
	else \
		RADIUS_VERSION="$(REL_VERSION)"; \
	fi; \
	cat > ./test/bicepconfig.json << EOF
{
  "experimentalFeaturesEnabled": {
    "extensibility": true
  },
  "extensions": {
    "radius": "br:$(LOCAL_REGISTRY_SERVER):$(LOCAL_REGISTRY_PORT)/radius:$$RADIUS_VERSION",
    "aws": "br:$(BICEP_TYPES_REGISTRY)/aws:latest"
  }
}
EOF

.PHONY: collect-radius-logs
collect-radius-logs: ## Collect Radius logs for debugging
	@echo "$(ARROW) Collecting Radius logs and events..."
	@mkdir -p $(RADIUS_CONTAINER_LOG_BASE)/radius-logs-events
	@for pod_name in $$(kubectl get pods -n radius-system -o jsonpath='{.items[*].metadata.name}'); do \
		kubectl logs $$pod_name -n radius-system > $(RADIUS_CONTAINER_LOG_BASE)/radius-logs-events/$$pod_name.txt; \
	done
	@kubectl get events -n radius-system > $(RADIUS_CONTAINER_LOG_BASE)/radius-logs-events/events.txt

.PHONY: collect-pod-details
collect-pod-details: ## Collect pod details for debugging
	@echo "$(ARROW) Collecting pod details..."
	@mkdir -p $(RADIUS_CONTAINER_LOG_BASE)
	@echo "kubectl get pods -A" >> $(RADIUS_CONTAINER_LOG_BASE)/pod-states.log
	@kubectl get pods -A >> $(RADIUS_CONTAINER_LOG_BASE)/pod-states.log
	@echo "kubectl describe pods -A" >> $(RADIUS_CONTAINER_LOG_BASE)/pod-states.log
	@kubectl describe pods -A >> $(RADIUS_CONTAINER_LOG_BASE)/pod-states.log

.PHONY: collect-recipe-logs
collect-recipe-logs: ## Collect recipe publishing logs
	@echo "$(ARROW) Collecting recipe logs..."
	@mkdir -p $(RADIUS_CONTAINER_LOG_BASE)/recipe-logs
	@for pod_name in $$(kubectl get pods -l app.kubernetes.io/name=tf-module-server -n radius-test-tf-module-server -o jsonpath='{.items[*].metadata.name}'); do \
		kubectl logs $$pod_name -n radius-test-tf-module-server > $(RADIUS_CONTAINER_LOG_BASE)/recipe-logs/$$pod_name.txt; \
	done
	@kubectl get events -n radius-test-tf-module-server > $(RADIUS_CONTAINER_LOG_BASE)/recipe-logs/events.txt

.PHONY: cleanup-functional-test-env
cleanup-functional-test-env: ## Clean up functional test environment
	@echo "$(ARROW) Cleaning up functional test environment..."
	@kind delete cluster --name radius-test || true
	@docker rm -f $(LOCAL_REGISTRY_NAME) || true
	@docker network rm radius-test-network || true
	@rm -rf $(TEMP_CERT_DIR) || true


