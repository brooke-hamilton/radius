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

set -e

WORKFLOW_FILE_NAME=$1
REPOSITORY=$2
BRANCH=$3

# Check if all required parameters are provided
if [ -z "$WORKFLOW_FILE_NAME" ] || [ -z "$REPOSITORY" ] || [ -z "$BRANCH" ]; then
  echo "❌ Error: Missing required parameters"
  echo "Usage: $0 <workflow_file_name> <repository> <branch>"
  exit 1
fi

gh api \
    --method POST \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "/repos/${REPOSITORY}/actions/workflows/${WORKFLOW_FILE_NAME}/dispatches" \
    -f "ref=${BRANCH}"

echo "✅ Dispatched workflow '$WORKFLOW_FILE_NAME' on repository '$REPOSITORY' for branch '$BRANCH'"
