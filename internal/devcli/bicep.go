package devcli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/radius-project/radius/internal/tooling"
)

type fileBicepStager struct {
	client *http.Client
}

func (s fileBicepStager) Stage(ctx context.Context, config bicepStageConfig) error {
	platform := "linux_" + config.arch
	if config.arch == "arm" {
		platform = "linux_amd64"
	}
	asset, ok := config.tool.Platforms[platform]
	if !ok {
		return fmt.Errorf("bicep does not provide a payload for architecture %q", config.arch)
	}
	if asset.Checksum == "" {
		return fmt.Errorf("bicep payload %s has no pinned checksum", platform)
	}

	if err := os.MkdirAll(config.outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	target := filepath.Join(config.outputDir, "bicep")
	matches, err := fileMatchesChecksum(target, asset.Checksum)
	if err != nil {
		return err
	}
	if !matches {
		values, err := config.tool.TemplateValues(platform, config.tool.Version)
		if err != nil {
			return fmt.Errorf("resolve bicep download metadata: %w", err)
		}
		url, err := tooling.ExpandTemplate(config.tool.DownloadTemplate, values)
		if err != nil {
			return fmt.Errorf("resolve bicep download URL: %w", err)
		}
		if err := s.download(ctx, url, target, asset.Checksum); err != nil {
			return err
		}
	}
	if err := os.Chmod(target, 0o755); err != nil {
		return fmt.Errorf("set bicep payload permissions: %w", err)
	}

	channel := config.channel
	if channel == "edge" {
		channel = "latest"
	}
	contents := fmt.Sprintf("{\n  \"extensions\": {\n    \"radius\": \"br:biceptypes.azurecr.io/radius:%s\",\n    \"aws\": \"br:biceptypes.azurecr.io/aws:%s\"\n  }\n}\n", channel, channel)
	if err := os.WriteFile(filepath.Join(config.outputDir, "bicepconfig.json"), []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write bicepconfig.json: %w", err)
	}
	return nil
}

func (s fileBicepStager) download(ctx context.Context, url, target, expectedChecksum string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create bicep download request: %w", err)
	}
	client := s.client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("download bicep payload: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download bicep payload: server returned %s", response.Status)
	}

	temp, err := os.CreateTemp(filepath.Dir(target), ".bicep-*")
	if err != nil {
		return fmt.Errorf("create temporary bicep payload: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(temp, hash), response.Body); err != nil {
		temp.Close()
		return fmt.Errorf("write bicep payload: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close bicep payload: %w", err)
	}
	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("verify bicep payload: checksum %s does not match expected %s", actualChecksum, expectedChecksum)
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace bicep payload: %w", err)
	}
	if err := os.Rename(tempName, target); err != nil {
		return fmt.Errorf("install bicep payload: %w", err)
	}
	return nil
}

func fileMatchesChecksum(path, expected string) (bool, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open existing bicep payload: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("hash existing bicep payload: %w", err)
	}
	return strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expected), nil
}
