package devcli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/radius-project/radius/internal/tooling"
	"github.com/stretchr/testify/require"
)

func TestFileBicepStager_DownloadsVerifiedPayloadAndConfig(t *testing.T) {
	t.Parallel()

	payload := []byte("linux bicep payload")
	hash := sha256.Sum256(payload)
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestedPath = request.URL.Path
		_, err := response.Write(payload)
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	outputDir := t.TempDir()
	stager := fileBicepStager{client: server.Client()}
	err := stager.Stage(context.Background(), bicepStageConfig{
		outputDir: outputDir,
		channel:   "edge",
		arch:      "amd64",
		tool: tooling.Tool{
			Version:          "v1.2.3",
			DownloadTemplate: server.URL + "/{repository}/{tag}/{asset}",
			Source:           tooling.Source{Repository: "Azure/bicep"},
			Platforms: map[string]tooling.Platform{
				"linux_amd64": {Asset: "bicep-linux-x64", Checksum: hex.EncodeToString(hash[:])},
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "/Azure/bicep/v1.2.3/bicep-linux-x64", requestedPath)
	require.FileExists(t, filepath.Join(outputDir, "bicep"))
	actualPayload, err := os.ReadFile(filepath.Join(outputDir, "bicep"))
	require.NoError(t, err)
	require.Equal(t, payload, actualPayload)
	config, err := os.ReadFile(filepath.Join(outputDir, "bicepconfig.json"))
	require.NoError(t, err)
	require.Equal(t, "{\n  \"extensions\": {\n    \"radius\": \"br:biceptypes.azurecr.io/radius:latest\",\n    \"aws\": \"br:biceptypes.azurecr.io/aws:latest\"\n  }\n}\n", string(config))
}

func TestFileBicepStager_RejectsChecksumMismatch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, err := response.Write([]byte("payload"))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	err := (fileBicepStager{client: server.Client()}).Stage(context.Background(), bicepStageConfig{
		outputDir: t.TempDir(),
		arch:      "arm",
		tool: tooling.Tool{
			Version:          "v1.2.3",
			DownloadTemplate: server.URL + "/{asset}",
			Platforms: map[string]tooling.Platform{
				"linux_amd64": {Asset: "bicep-linux-x64", Checksum: "0000000000000000000000000000000000000000000000000000000000000000"},
			},
		},
	})
	require.ErrorContains(t, err, "checksum")
}
