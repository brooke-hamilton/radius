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

package graphstore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_idToTreePath(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "full resource ID",
			id:       "/planes/radius/local/resourceGroups/rg1/providers/Applications.Core/applications/my-app",
			expected: "planes/radius/local/resourcegroups/rg1/providers/applications.core/applications/my-app",
		},
		{
			name:     "scope ID",
			id:       "/planes/radius/local/resourceGroups/rg1",
			expected: "planes/radius/local/resourcegroups/rg1",
		},
		{
			name:     "mixed case",
			id:       "/Planes/Radius/Local",
			expected: "planes/radius/local",
		},
		{
			name:     "no leading slash",
			id:       "planes/radius/local",
			expected: "planes/radius/local",
		},
		{
			name:     "single segment",
			id:       "/planes",
			expected: "planes",
		},
		{
			name:     "empty string",
			id:       "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := idToTreePath(tt.id)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_Client_idToGrifPath(t *testing.T) {
	c := &Client{graphName: "radius"}

	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "full resource ID",
			id:       "/planes/radius/local/resourceGroups/rg1/providers/Applications.Core/applications/my-app",
			expected: "radius/planes/radius/local/resourcegroups/rg1/providers/applications.core/applications/my-app/__data",
		},
		{
			name:     "scope ID",
			id:       "/planes/radius/local",
			expected: "radius/planes/radius/local/__data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.idToGrifPath(tt.id)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_Client_scopeToGrifPath(t *testing.T) {
	c := &Client{graphName: "radius"}

	tests := []struct {
		name     string
		scope    string
		expected string
	}{
		{
			name:     "plane scope with trailing slash",
			scope:    "/planes/radius/local/",
			expected: "radius/planes/radius/local",
		},
		{
			name:     "resource group scope",
			scope:    "/planes/radius/local/resourceGroups/group1",
			expected: "radius/planes/radius/local/resourcegroups/group1",
		},
		{
			name:     "planes only",
			scope:    "/planes",
			expected: "radius/planes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.scopeToGrifPath(tt.scope)
			require.Equal(t, tt.expected, result)
		})
	}
}
