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
	"strings"
)

// dataBlobName is the name of the blob entry within a tree node that stores
// the resource data. Using a reserved leaf name allows a path segment to be
// both a tree (with children) and hold data (as a blob child named __data).
// This solves the problem where a resource like .../resource1 must coexist
// with nested resources like .../resource1/nestedType/nested1.
const dataBlobName = "__data"

// idToTreePath converts a Radius resource ID to a grif tree path.
// The resource ID segments are lowercased and the leading slash is stripped.
// The graph name is NOT included in the returned path.
//
// Example: "/planes/radius/local/resourceGroups/rg1" → "planes/radius/local/resourcegroups/rg1"
func idToTreePath(id string) string {
	trimmed := strings.TrimPrefix(id, "/")
	return strings.ToLower(trimmed)
}

// idToGrifPath converts a Radius resource ID to a full grif path including
// the graph name prefix and the __data leaf blob.
//
// Example: (graphName="radius", id="/planes/radius/local/resourceGroups/rg1")
//
//	→ "radius/planes/radius/local/resourcegroups/rg1/__data"
func (c *Client) idToGrifPath(id string) string {
	return c.graphName + "/" + idToTreePath(id) + "/" + dataBlobName
}

// scopeToGrifPath converts a query root scope to a grif tree path prefix
// for tree walking. The scope is lowercased and trimmed of leading/trailing
// slashes, then prefixed with the graph name.
func (c *Client) scopeToGrifPath(rootScope string) string {
	trimmed := strings.Trim(rootScope, "/")
	return c.graphName + "/" + strings.ToLower(trimmed)
}
