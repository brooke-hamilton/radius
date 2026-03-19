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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"github.com/radius-project/radius/pkg/components/database"
	"github.com/radius-project/radius/test/testcontext"
	"github.com/radius-project/radius/test/ucp/storetest"
)

const testdataDir = "testdata"

// setupTestRepo creates a new Git repository in a unique subdirectory for
// integration testing. The repo is initialized with an initial commit so HEAD
// is valid (required by grif).
func setupTestRepo(t *testing.T) string {
	t.Helper()

	if err := os.MkdirAll(testdataDir, 0o755); err != nil {
		t.Fatalf("failed to create testdata dir: %v", err)
	}

	dir, err := os.MkdirTemp(testdataDir, sanitizeTestName(t.Name())+"-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	absDir, _ := filepath.Abs(dir)
	t.Logf("test repo: %s", absDir)

	t.Cleanup(func() {
		if !t.Failed() {
			os.RemoveAll(dir)
		}
	})

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("failed to init test repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("failed to write readme: %v", err)
	}

	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	_, err = wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@test.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	return absDir
}

func sanitizeTestName(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ' ' {
			return '_'
		}
		return r
	}, name)
}

func TestNewClient(t *testing.T) {
	t.Run("creates client with default graph name", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "")
		require.NoError(t, err)
		require.NotNil(t, client)
		require.Equal(t, "radius", client.graphName)
	})

	t.Run("creates client with custom graph name", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "my-graph")
		require.NoError(t, err)
		require.NotNil(t, client)
		require.Equal(t, "my-graph", client.graphName)
	})

	t.Run("fails for empty repo path", func(t *testing.T) {
		_, err := NewClient("", "test")
		require.Error(t, err)
	})

	t.Run("fails for invalid repo path", func(t *testing.T) {
		_, err := NewClient("/nonexistent/path", "test")
		require.Error(t, err)
	})

	t.Run("succeeds when graph already exists", func(t *testing.T) {
		dir := setupTestRepo(t)
		client1, err := NewClient(dir, "radius")
		require.NoError(t, err)
		require.NotNil(t, client1)

		// Creating another client for the same graph should succeed
		client2, err := NewClient(dir, "radius")
		require.NoError(t, err)
		require.NotNil(t, client2)
	})
}

func TestGet(t *testing.T) {
	t.Run("returns not found for missing resource", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj, err := client.Get(ctx, "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1")
		require.ErrorIs(t, err, &database.ErrNotFound{})
		require.Nil(t, obj)
	})

	t.Run("returns error for invalid ID", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		_, err = client.Get(ctx, "")
		require.ErrorIs(t, err, &database.ErrInvalid{})
	})
}

func TestSave(t *testing.T) {
	t.Run("saves and retrieves a resource", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{
				"value": "1",
				"properties": map[string]any{
					"resource": "1",
				},
			},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)
		require.NotEmpty(t, obj.ETag)

		got, err := client.Get(ctx, obj.ID)
		require.NoError(t, err)
		require.Equal(t, obj.ID, got.ID)
		require.Equal(t, obj.ETag, got.ETag)
	})

	t.Run("updates an existing resource", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)
		etag1 := obj.ETag

		obj.Data = map[string]any{"value": "2"}
		err = client.Save(ctx, obj)
		require.NoError(t, err)
		require.NotEqual(t, etag1, obj.ETag)

		got, err := client.Get(ctx, obj.ID)
		require.NoError(t, err)
		require.Equal(t, obj.ETag, got.ETag)
	})

	t.Run("returns ETag on save", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)
		require.NotEmpty(t, obj.ETag)
	})
}

func TestDelete(t *testing.T) {
	t.Run("deletes existing resource", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)

		err = client.Delete(ctx, obj.ID)
		require.NoError(t, err)

		got, err := client.Get(ctx, obj.ID)
		require.ErrorIs(t, err, &database.ErrNotFound{})
		require.Nil(t, got)
	})

	t.Run("returns not found for missing resource", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		err = client.Delete(ctx, "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1")
		require.ErrorIs(t, err, &database.ErrNotFound{})
	})
}

func TestSaveWithETag(t *testing.T) {
	t.Run("succeeds with matching ETag", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)

		obj.Data = map[string]any{"value": "2"}
		err = client.Save(ctx, obj, database.WithETag(obj.ETag))
		require.NoError(t, err)
	})

	t.Run("fails with stale ETag", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)

		obj.Data = map[string]any{"value": "2"}
		err = client.Save(ctx, obj, database.WithETag("stale-etag"))
		require.ErrorIs(t, err, &database.ErrConcurrency{})
	})

	t.Run("fails with ETag when resource does not exist", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj, database.WithETag("some-etag"))
		require.ErrorIs(t, err, &database.ErrConcurrency{})
	})
}

func TestDeleteWithETag(t *testing.T) {
	t.Run("succeeds with matching ETag", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)

		err = client.Delete(ctx, obj.ID, database.WithETag(obj.ETag))
		require.NoError(t, err)
	})

	t.Run("fails with stale ETag", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)

		err = client.Delete(ctx, obj.ID, database.WithETag("stale-etag"))
		require.ErrorIs(t, err, &database.ErrConcurrency{})
	})

	t.Run("fails with ETag when resource deleted", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		err = client.Delete(ctx, "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1", database.WithETag("some-etag"))
		require.ErrorIs(t, err, &database.ErrConcurrency{})
	})
}

func TestCommitPerMutation(t *testing.T) {
	t.Run("each save creates a Git commit", func(t *testing.T) {
		dir := setupTestRepo(t)
		client, err := NewClient(dir, "radius")
		require.NoError(t, err)

		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()

		obj := &database.Object{
			Metadata: database.Metadata{
				ID: "/planes/radius/local/resourceGroups/rg1/providers/System.Resources/resourceType1/resource1",
			},
			Data: map[string]any{"value": "1"},
		}

		err = client.Save(ctx, obj)
		require.NoError(t, err)

		obj.Data = map[string]any{"value": "2"}
		err = client.Save(ctx, obj)
		require.NoError(t, err)

		// Verify we have commits in the graph ref log
		logResult, err := logGraphCommits(dir, "radius")
		require.NoError(t, err)
		// Init commit + 2 saves = 3 commits
		require.GreaterOrEqual(t, len(logResult), 3)
	})
}

// logGraphCommits returns the commit messages for the graph ref.
func logGraphCommits(repoPath string, graphName string) ([]string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Reference(plumbing.ReferenceName("refs/infra/"+graphName), true)
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	var messages []string
	iter := object.NewCommitPreorderIter(commit, nil, nil)
	err = iter.ForEach(func(c *object.Commit) error {
		messages = append(messages, c.Message)
		return nil
	})

	return messages, err
}

// TestConformance runs the shared conformance tests from the storetest package.
func TestConformance(t *testing.T) {
	dir := setupTestRepo(t)
	client, err := NewClient(dir, "radius")
	require.NoError(t, err)

	clear := func(t *testing.T) {
		// To clear state, we create a fresh client with a new graph name
		// since we can't easily clear all data from an existing graph.
		// The conformance tests use a single client instance, so we need
		// to reinitialize the graph.
		//
		// We achieve this by creating a new test repo for each sub-test clear.
		newDir := setupTestRepo(t)
		newClient, err := NewClient(newDir, "radius")
		require.NoError(t, err)

		// Swap the client's internal state to point at the new repo
		client.mutex.Lock()
		client.repoPath = newClient.repoPath
		client.graphName = newClient.graphName
		client.mutex.Unlock()
	}

	storetest.RunTest(t, client, clear)
}
