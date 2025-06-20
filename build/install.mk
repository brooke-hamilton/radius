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

##@ Install

RAD_LOCATION := /usr/local/bin/rad

.PHONY: install
install: build-binaries ## Installs a local build for development
	@echo "$(ARROW) Installing rad"
	sudo cp $(OUT_DIR)/$(GOOS)_$(GOARCH)/$(BUILDTYPE_DIR)/rad$(BINARY_EXT) $(RAD_LOCATION)

.PHONY: install-latest
install-latest: ## Installs the latest release from GitHub
	@echo "$(ARROW) Installing latest rad release"
	@bash ./deploy/install.sh

FLUX_VERSION ?= 2.5.1
.PHONY: install-flux
install-flux: ## Installs flux using the install script
	@echo "$(ARROW) Installing flux"
	@export FLUX_VERSION=$(FLUX_VERSION) && curl -s https://fluxcd.io/install.sh | sudo -E bash
	@./.github/actions/install-flux/install-flux.sh $(FLUX_VERSION)

GITEA_VERSION ?= v11.0.0
GITEA_USERNAME ?= "testuser"
GITEA_EMAIL ?= "testuser@radapp.io"
GITEA_ACCESS_TOKEN_NAME ?= "radius-functional-test"
GITEA_PASSWORD ?= ""
.PHONY: install-gitea
install-gitea: ## Installs gitea
	@echo "$(ARROW) Installing gitea"
	@export GITEA_PASSWORD=$(GITEA_PASSWORD) && .github/actions/install-gitea/install-gitea.sh $(GITEA_VERSION) $(GITEA_USERNAME) $(GITEA_EMAIL) $(GITEA_ACCESS_TOKEN_NAME)