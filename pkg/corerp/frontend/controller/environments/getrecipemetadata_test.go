// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package environments

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	ctrl "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	"github.com/project-radius/radius/pkg/corerp/api/v20220315privatepreview"
	"github.com/project-radius/radius/pkg/ucp/store"
	"github.com/project-radius/radius/test/testutil"
	"github.com/stretchr/testify/require"
)

func TestGetRecipeMetadataRun_20220315PrivatePreview(t *testing.T) {
	mctrl := gomock.NewController(t)
	defer mctrl.Finish()
	mStorageClient := store.NewMockStorageClient(mctrl)
	ctx := context.Background()

	t.Parallel()
	t.Run("get recipe metadata run", func(t *testing.T) {
		envDataModel, expectedOutput := getTestModelsGetRecipeMetadata20220315privatepreview()
		w := httptest.NewRecorder()
		req, _ := testutil.GetARMTestHTTPRequest(ctx, v1.OperationPost.HTTPMethod(), testHeaderfilegetrecipemetadata, nil)

		mStorageClient.
			EXPECT().
			Get(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, id string, _ ...store.GetOptions) (*store.Object, error) {
				return &store.Object{
					Metadata: store.Metadata{ID: id, ETag: "etag"},
					Data:     envDataModel,
				}, nil
			})
		ctx := testutil.ARMTestContextFromRequest(req)

		opts := ctrl.Options{
			StorageClient: mStorageClient,
		}
		ctl, err := NewGetRecipeMetadata(opts)
		require.NoError(t, err)
		resp, err := ctl.Run(ctx, w, req)
		require.NoError(t, err)
		_ = resp.Apply(ctx, w, req)
		require.Equal(t, 200, w.Result().StatusCode)

		actualOutput := &v20220315privatepreview.EnvironmentResource{}
		_ = json.Unmarshal(w.Body.Bytes(), actualOutput)
		require.Equal(t, expectedOutput, actualOutput)
	})

	t.Run("get recipe metadata run non existing environment", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := testutil.GetARMTestHTTPRequest(ctx, v1.OperationPost.HTTPMethod(), testHeaderfilegetrecipemetadata, nil)
		ctx := testutil.ARMTestContextFromRequest(req)

		mStorageClient.
			EXPECT().
			Get(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, id string, _ ...store.GetOptions) (*store.Object, error) {
				return nil, &store.ErrNotFound{}
			})
		opts := ctrl.Options{
			StorageClient: mStorageClient,
		}
		ctl, err := NewGetRecipeMetadata(opts)
		require.NoError(t, err)
		resp, err := ctl.Run(ctx, w, req)
		require.NoError(t, err)
		_ = resp.Apply(ctx, w, req)
		result := w.Result()
		require.Equal(t, 404, result.StatusCode)

		body := result.Body
		defer body.Close()
		payload, err := io.ReadAll(body)
		require.NoError(t, err)

		armerr := v1.ErrorResponse{}
		err = json.Unmarshal(payload, &armerr)
		require.NoError(t, err)
		require.Equal(t, v1.CodeNotFound, armerr.Error.Code)
		require.Contains(t, armerr.Error.Message, "the resource with id")
		require.Contains(t, armerr.Error.Message, "was not found")
	})

	t.Run("get recipe metadata non existing recipe", func(t *testing.T) {
		envDataModel, _ := getTestModelsGetRecipeMetadata20220315privatepreview()
		w := httptest.NewRecorder()
		req, _ := testutil.GetARMTestHTTPRequest(ctx, v1.OperationPost.HTTPMethod(), testHeaderfilegetrecipemetadatanotexisting, nil)
		ctx := testutil.ARMTestContextFromRequest(req)

		mStorageClient.
			EXPECT().
			Get(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, id string, _ ...store.GetOptions) (*store.Object, error) {
				return &store.Object{
					Metadata: store.Metadata{ID: id, ETag: "etag"},
					Data:     envDataModel,
				}, nil
			})

		opts := ctrl.Options{
			StorageClient: mStorageClient,
		}
		ctl, err := NewGetRecipeMetadata(opts)
		require.NoError(t, err)
		resp, err := ctl.Run(ctx, w, req)
		require.NoError(t, err)
		_ = resp.Apply(ctx, w, req)
		result := w.Result()
		require.Equal(t, 404, result.StatusCode)

		body := result.Body
		defer body.Close()
		payload, err := io.ReadAll(body)
		require.NoError(t, err)

		armerr := v1.ErrorResponse{}
		err = json.Unmarshal(payload, &armerr)
		require.NoError(t, err)
		require.Equal(t, v1.CodeNotFound, armerr.Error.Code)
		require.Contains(t, armerr.Error.Message, "Recipe with name \"mongodb\" not found on environment with id")
	})
}

func TestGetRecipeMetadataFromRegistry(t *testing.T) {
	ctx := context.Background()

	t.Run("get recipe metadata from registry", func(t *testing.T) {
		templatePath := "radiusdev.azurecr.io/recipes/functionaltest/parameters/mongodatabases/azure:1.0"
		output, err := getRecipeMetadataFromRegistry(ctx, templatePath, "mongodb")
		require.NoError(t, err)
		expectedOutput := map[string]any{
			"mongodbName": map[string]any{
				"type": "string",
			},
			"documentdbName": map[string]any{
				"type": "string",
			},
			"location": map[string]any{
				"type":         "string",
				"defaultValue": "[resourceGroup().location]",
			},
		}
		require.Equal(t, expectedOutput, output)
	})

	t.Run("get recipe metadata from registry with context parameter", func(t *testing.T) {
		templatePath := "radiusdev.azurecr.io/recipes/functionaltest/context/mongodatabases/azure:1.0"
		output, err := getRecipeMetadataFromRegistry(ctx, templatePath, "mongodb")
		require.NoError(t, err)
		expectedOutput := map[string]any{
			"location": map[string]any{
				"type":         "string",
				"defaultValue": "[resourceGroup().location]",
			},
			"rg": map[string]any{
				"type":         "string",
				"defaultValue": "[resourceGroup().name]",
			},
		}
		require.Equal(t, expectedOutput, output)
	})

	t.Run("get recipe metadata from registry with invalid path", func(t *testing.T) {
		templatePath := "radiusdev.azurecr.io/recipes/functionaltest/test/mongodatabases/azure:1.0"
		_, err := getRecipeMetadataFromRegistry(ctx, templatePath, "mongodb")
		require.Error(t, err, "failed to fetch template from the path \"radiusdev.azurecr.io/recipes/functionaltest/test/mongodatabases/azure:1.0\" for recipe \"mongodb\": radiusdev.azurecr.io/recipes/functionaltest/test/mongodatabases/azure:1.0: not found")
	})

	t.Run("get recipe metadata from registry with no parameters", func(t *testing.T) {
		templatePath := "radiusdev.azurecr.io/pratikshya/recipe/mongodatabases/azure:1.0"
		output, err := getRecipeMetadataFromRegistry(ctx, templatePath, "mongodb")
		require.NoError(t, err)
		expectedOutput := map[string]any{}
		require.Equal(t, expectedOutput, output)
	})
}