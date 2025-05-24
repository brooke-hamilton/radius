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

##@ Workflow

# Default values for workflow dispatch
ifndef WORKFLOW_FILE_NAME
$(error WORKFLOW_FILE_NAME is required but not set)
endif
REPOSITORY ?= $(shell gh repo view --json nameWithOwner --jq '.nameWithOwner')
BRANCH ?= main
.PHONY: workflow
workflow: ## Dispatch a GitHub workflow using the dispatch-workflow script
	@echo "$(ARROW) Dispatching workflow $(WORKFLOW_FILE_NAME) on $(REPOSITORY) for branch $(BRANCH)"
	@./.github/actions/dispatch-workflow/dispatch-workflow.sh $(WORKFLOW_FILE_NAME) $(REPOSITORY) $(BRANCH)
