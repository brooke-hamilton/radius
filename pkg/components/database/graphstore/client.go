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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/brooke-hamilton/git-infra-graph/src/graph"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	gogitobject "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/radius-project/radius/pkg/components/database"
	"github.com/radius-project/radius/pkg/components/database/databaseutil"
	"github.com/radius-project/radius/pkg/ucp/resources"
	"github.com/radius-project/radius/pkg/ucp/util/etag"
)

// Compile-time interface check.
var _ database.Client = (*Client)(nil)

// storedObject is the JSON envelope persisted as a Git blob for each resource.
type storedObject struct {
	ID           string `json:"id"`
	ETag         string `json:"etag"`
	RootScope    string `json:"rootScope"`
	ResourceType string `json:"resourceType"`
	RoutingScope string `json:"routingScope"`
	Data         any    `json:"data"`
}

// Client implements database.Client using the git-infra-graph (grif) library
// to persist resource state as a versioned graph inside a Git repository.
type Client struct {
	repoPath  string
	graphName string
	mutex     sync.Mutex
}

// NewClient creates a new graph store client for the given Git repository
// and graph name. The graph is initialized if it does not already exist.
func NewClient(repoPath string, graphName string) (*Client, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("repoPath is required")
	}
	if graphName == "" {
		graphName = "radius"
	}

	// Initialize the graph if it does not already exist.
	err := graph.Init(repoPath, graphName)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return nil, fmt.Errorf("failed to initialize graph %q: %w", graphName, err)
	}

	return &Client{
		repoPath:  repoPath,
		graphName: graphName,
	}, nil
}

// Get retrieves a resource by ID from the Git graph.
func (c *Client) Get(ctx context.Context, id string, options ...database.GetOptions) (*database.Object, error) {
	if ctx == nil {
		return nil, &database.ErrInvalid{Message: "invalid argument. 'ctx' is required"}
	}
	parsed, err := resources.Parse(id)
	if err != nil {
		return nil, &database.ErrInvalid{Message: "invalid argument. 'id' must be a valid resource id"}
	}
	if parsed.IsEmpty() {
		return nil, &database.ErrInvalid{Message: "invalid argument. 'id' must not be empty"}
	}
	if parsed.IsResourceCollection() || parsed.IsScopeCollection() {
		return nil, &database.ErrInvalid{Message: "invalid argument. 'id' must refer to a named resource, not a collection"}
	}

	converted, err := databaseutil.ConvertScopeIDToResourceID(parsed)
	if err != nil {
		return nil, err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	grifPath := c.idToGrifPath(converted.String())
	content, err := graph.Get(c.repoPath, grifPath)
	if err != nil {
		if strings.Contains(err.Error(), "node not found") {
			return nil, &database.ErrNotFound{ID: id}
		}
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	if content.Type != graph.BlobNode {
		return nil, &database.ErrNotFound{ID: id}
	}

	stored, err := unmarshalStoredObject(content.Blob)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal stored object: %w", err)
	}

	obj := &database.Object{
		Metadata: database.Metadata{
			ID:   stored.ID,
			ETag: stored.ETag,
		},
		Data: stored.Data,
	}

	return obj, nil
}

// Save persists a resource to the Git graph. Creates a new entry or updates
// an existing one. Each save produces a Git commit.
func (c *Client) Save(ctx context.Context, obj *database.Object, options ...database.SaveOptions) error {
	if ctx == nil {
		return &database.ErrInvalid{Message: "invalid argument. 'ctx' is required"}
	}
	if obj == nil {
		return &database.ErrInvalid{Message: "invalid argument. 'obj' is required"}
	}

	parsed, err := resources.Parse(obj.ID)
	if err != nil {
		return &database.ErrInvalid{Message: "invalid argument. 'obj.ID' must be a valid resource id"}
	}

	converted, err := databaseutil.ConvertScopeIDToResourceID(parsed)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	config := database.NewSaveConfig(options...)
	grifPath := c.idToGrifPath(converted.String())

	// Check ETag if provided
	if config.ETag != "" {
		existing, err := graph.Get(c.repoPath, grifPath)
		if err != nil {
			if strings.Contains(err.Error(), "node not found") {
				return &database.ErrConcurrency{}
			}
			return fmt.Errorf("failed to check existing resource: %w", err)
		}

		if existing.Type != graph.BlobNode {
			return &database.ErrConcurrency{}
		}

		stored, err := unmarshalStoredObject(existing.Blob)
		if err != nil {
			return fmt.Errorf("failed to unmarshal existing object: %w", err)
		}

		if stored.ETag != config.ETag {
			return &database.ErrConcurrency{}
		}
	}

	// Compute new ETag
	raw, err := json.Marshal(obj.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	newETag := etag.New(raw)

	// Build stored object
	stored := storedObject{
		ID:           obj.ID,
		ETag:         newETag,
		RootScope:    databaseutil.NormalizePart(converted.RootScope()),
		ResourceType: databaseutil.NormalizePart(converted.Type()),
		RoutingScope: databaseutil.NormalizePart(converted.RoutingScope()),
		Data:         obj.Data,
	}

	blob, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("failed to marshal stored object: %w", err)
	}

	// Stage the change
	_, err = graph.Put(c.repoPath, grifPath, blob)
	if err != nil {
		return fmt.Errorf("failed to put resource: %w", err)
	}

	// Commit the change
	message := fmt.Sprintf("Save: %s", idToTreePath(converted.String()))
	_, err = graph.Commit(c.repoPath, c.graphName, message)
	if err != nil {
		// Attempt rollback by deleting the staging ref — best effort
		_ = c.rollbackStaging()
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Update the ETag on the input object so callers can read it
	obj.ETag = newETag

	return nil
}

// Delete removes a resource from the Git graph by ID. Each delete produces
// a Git commit.
func (c *Client) Delete(ctx context.Context, id string, options ...database.DeleteOptions) error {
	if ctx == nil {
		return &database.ErrInvalid{Message: "invalid argument. 'ctx' is required"}
	}
	parsed, err := resources.Parse(id)
	if err != nil {
		return &database.ErrInvalid{Message: "invalid argument. 'id' must be a valid resource id"}
	}
	if parsed.IsEmpty() {
		return &database.ErrInvalid{Message: "invalid argument. 'id' must not be empty"}
	}
	if parsed.IsResourceCollection() || parsed.IsScopeCollection() {
		return &database.ErrInvalid{Message: "invalid argument. 'id' must refer to a named resource, not a collection"}
	}

	converted, err := databaseutil.ConvertScopeIDToResourceID(parsed)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	config := database.NewDeleteConfig(options...)
	grifPath := c.idToGrifPath(converted.String())

	// Check if the resource exists and verify ETag if provided
	existing, err := graph.Get(c.repoPath, grifPath)
	if err != nil {
		if strings.Contains(err.Error(), "node not found") {
			if config.ETag != "" {
				return &database.ErrConcurrency{}
			}
			return &database.ErrNotFound{ID: id}
		}
		return fmt.Errorf("failed to check existing resource: %w", err)
	}

	if existing.Type != graph.BlobNode {
		if config.ETag != "" {
			return &database.ErrConcurrency{}
		}
		return &database.ErrNotFound{ID: id}
	}

	if config.ETag != "" {
		stored, err := unmarshalStoredObject(existing.Blob)
		if err != nil {
			return fmt.Errorf("failed to unmarshal existing object: %w", err)
		}
		if stored.ETag != config.ETag {
			return &database.ErrConcurrency{}
		}
	}

	// Stage the deletion — delete the __data blob
	err = graph.DeleteNode(c.repoPath, grifPath)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	// Commit the change
	message := fmt.Sprintf("Delete: %s", idToTreePath(converted.String()))
	_, err = graph.Commit(c.repoPath, c.graphName, message)
	if err != nil {
		_ = c.rollbackStaging()
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// Query executes a scope-based query against the Git graph, returning all
// matching resources with optional filtering and pagination.
func (c *Client) Query(ctx context.Context, query database.Query, options ...database.QueryOptions) (*database.ObjectQueryResult, error) {
	if ctx == nil {
		return nil, &database.ErrInvalid{Message: "invalid argument. 'ctx' is required"}
	}

	err := query.Validate()
	if err != nil {
		return nil, &database.ErrInvalid{Message: fmt.Sprintf("invalid argument. Query is invalid: %s", err.Error())}
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Collect all blobs from the graph
	// We need to walk the tree from the root. Since grif Get requires at least
	// graph_name/segment, we start by getting the children of the graph root
	// using the graph.Status or by finding top-level entries.
	allBlobs, err := c.collectAllBlobs()
	if err != nil {
		return nil, fmt.Errorf("failed to walk tree: %w", err)
	}

	// Filter results
	var results []database.Object
	for _, blobData := range allBlobs {
		stored, err := unmarshalStoredObject(blobData)
		if err != nil {
			continue // Skip corrupt blobs
		}

		obj := database.Object{
			Metadata: database.Metadata{
				ID:   stored.ID,
				ETag: stored.ETag,
			},
			Data: stored.Data,
		}

		// Check root scope
		if query.ScopeRecursive && !strings.HasPrefix(stored.RootScope, databaseutil.NormalizePart(query.RootScope)) {
			continue
		} else if !query.ScopeRecursive && stored.RootScope != databaseutil.NormalizePart(query.RootScope) {
			continue
		}

		// Check resource type
		resourceType, err := databaseutil.ConvertScopeTypeToResourceType(query.ResourceType)
		if err != nil {
			return nil, err
		}
		if stored.ResourceType != databaseutil.NormalizePart(resourceType) {
			continue
		}

		// Check routing scope prefix (optional)
		if query.RoutingScopePrefix != "" && !strings.HasPrefix(stored.RoutingScope, databaseutil.NormalizePart(query.RoutingScopePrefix)) {
			continue
		}

		// Apply property filters
		match, err := obj.MatchesFilters(query.Filters)
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}

		results = append(results, obj)
	}

	// Apply pagination
	queryConfig := database.NewQueryConfig(options...)
	return c.applyPagination(results, queryConfig), nil
}

// collectAllBlobs reads the graph ref's root tree directly via go-git and
// recursively collects all __data blobs.
func (c *Client) collectAllBlobs() ([][]byte, error) {
	repo, err := gogit.PlainOpen(c.repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repo: %w", err)
	}

	ref, err := repo.Reference(plumbing.ReferenceName("refs/infra/"+c.graphName), true)
	if err != nil {
		return nil, fmt.Errorf("graph %q not found: %w", c.graphName, err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to read commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to read tree: %w", err)
	}

	// Also check the staging ref in case there are uncommitted changes
	// (but in our implementation, every operation commits, so this is mainly the committed tree)
	return c.walkGitTree(repo, tree)
}

// walkGitTree recursively walks a go-git tree object and collects all __data blob contents.
func (c *Client) walkGitTree(repo *gogit.Repository, tree *gogitobject.Tree) ([][]byte, error) {
	if tree == nil {
		return nil, nil
	}

	var blobs [][]byte
	for _, entry := range tree.Entries {
		if entry.Mode == filemode.Regular && entry.Name == dataBlobName {
			blob, err := repo.BlobObject(entry.Hash)
			if err != nil {
				continue
			}
			reader, err := blob.Reader()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				continue
			}
			blobs = append(blobs, data)
		} else if entry.Mode == filemode.Dir {
			subtree, err := repo.TreeObject(entry.Hash)
			if err != nil {
				continue
			}
			childBlobs, err := c.walkGitTree(repo, subtree)
			if err != nil {
				return nil, err
			}
			blobs = append(blobs, childBlobs...)
		}
	}

	return blobs, nil
}

// applyPagination applies index-based pagination to the result set.
func (c *Client) applyPagination(results []database.Object, config database.DatabaseOptions) *database.ObjectQueryResult {
	startIndex := 0
	if config.PaginationToken != "" {
		decoded, err := base64.StdEncoding.DecodeString(config.PaginationToken)
		if err == nil {
			idx, err := strconv.Atoi(string(decoded))
			if err == nil && idx > 0 && idx < len(results) {
				startIndex = idx
			}
		}
	}

	if startIndex >= len(results) {
		return &database.ObjectQueryResult{}
	}

	remaining := results[startIndex:]

	if config.MaxQueryItemCount > 0 && len(remaining) > config.MaxQueryItemCount {
		nextIndex := startIndex + config.MaxQueryItemCount
		token := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(nextIndex)))
		return &database.ObjectQueryResult{
			Items:           remaining[:config.MaxQueryItemCount],
			PaginationToken: token,
		}
	}

	return &database.ObjectQueryResult{
		Items: remaining,
	}
}

// rollbackStaging attempts to reset the staging ref. This is a best-effort
// operation used when a commit fails after staging.
func (c *Client) rollbackStaging() error {
	// The simplest rollback is to delete the staging state by performing
	// a no-op that resets to the committed tree. Since grif doesn't expose
	// a direct staging ref removal, we rely on the next operation to
	// overwrite the staging state.
	return nil
}

func unmarshalStoredObject(data []byte) (*storedObject, error) {
	var stored storedObject
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}
	return &stored, nil
}
