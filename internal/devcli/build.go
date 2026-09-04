package devcli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/radius-project/radius/internal/tooling"
)

const basePackageName = "github.com/radius-project/radius"

var binaries = []struct {
	name       string
	entrypoint string
}{
	{name: "docgen", entrypoint: "cmd/docgen"},
	{name: "rad", entrypoint: "cmd/rad"},
	{name: "applications-rp", entrypoint: "cmd/applications-rp"},
	{name: "dynamic-rp", entrypoint: "cmd/dynamic-rp"},
	{name: "ucpd", entrypoint: "cmd/ucpd"},
	{name: "controller", entrypoint: "cmd/controller"},
	{name: "testrp", entrypoint: "test/testrp"},
	{name: "magpiego", entrypoint: "test/magpiego"},
	{name: "pre-upgrade", entrypoint: "cmd/pre-upgrade"},
}

type bicepStageConfig struct {
	outputDir string
	channel   string
	arch      string
	tool      tooling.Tool
}

type bicepStager interface {
	Stage(context.Context, bicepStageConfig) error
}

type workflows struct {
	root         string
	stdout       io.Writer
	runner       runner
	lookupEnv    func(string) (string, bool)
	bicepStager  bicepStager
	manageChecks func(context.Context) error
}

type buildConfig struct {
	goos             string
	goarch           string
	buildType        string
	gcflags          string
	ldflags          string
	outputDir        string
	commandEnv       map[string]string
	releaseChannel   string
	bicepTool        tooling.Tool
	terraformVersion string
}

func (w *workflows) Build(ctx context.Context) error {
	config, err := w.resolveBuildConfig(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintln(w.stdout, "==> Building all packages")
	if err := w.runner.Run(ctx, command{
		name: "go",
		args: []string{"build", "-v", "-gcflags", config.gcflags, "-ldflags=" + config.ldflags, "./..."},
		dir:  w.root,
		env:  config.commandEnv,
	}); err != nil {
		return fmt.Errorf("build packages: %w", err)
	}

	extension := ""
	if config.goos == "windows" {
		extension = ".exe"
	}
	for _, binary := range binaries {
		output := filepath.Join(config.outputDir, binary.name+extension)
		fmt.Fprintf(w.stdout, "==> Building %s on %s/%s to %s\n", binary.name, config.goos, config.goarch, output)
		if err := w.runner.Run(ctx, command{
			name: "go",
			args: []string{"build", "-v", "-gcflags", config.gcflags, "-ldflags=" + config.ldflags, "-o", output},
			dir:  filepath.Join(w.root, filepath.FromSlash(binary.entrypoint)),
			env:  config.commandEnv,
		}); err != nil {
			return fmt.Errorf("build %s: %w", binary.name, err)
		}
	}

	fmt.Fprintf(w.stdout, "==> Building bicep container payload on %s/%s to %s\n", config.goos, config.goarch, filepath.Join(config.outputDir, "bicep"))
	if err := w.bicepStager.Stage(ctx, bicepStageConfig{
		outputDir: filepath.Join(config.outputDir, "bicep"),
		channel:   config.releaseChannel,
		arch:      config.goarch,
		tool:      config.bicepTool,
	}); err != nil {
		return fmt.Errorf("build bicep payload: %w", err)
	}
	return nil
}

func (w *workflows) resolveBuildConfig(ctx context.Context) (buildConfig, error) {
	goos, err := w.targetEnvironmentOrOutput(ctx, "RADIUS_DEV_GOOS", "GOOS", "go", "env", "GOOS")
	if err != nil {
		return buildConfig{}, err
	}
	goarch, err := w.targetEnvironmentOrOutput(ctx, "RADIUS_DEV_GOARCH", "GOARCH", "go", "env", "GOARCH")
	if err != nil {
		return buildConfig{}, err
	}
	gitCommit, err := w.environmentOrOutput(ctx, "GIT_COMMIT", "git", "rev-list", "-1", "HEAD")
	if err != nil {
		return buildConfig{}, err
	}
	gitVersion, err := w.environmentOrOutput(ctx, "GIT_VERSION", "git", "describe", "--always", "--abbrev=7", "--dirty", "--tags")
	if err != nil {
		return buildConfig{}, err
	}

	manifest, err := tooling.LoadManifest(filepath.Join(w.root, "build", "tools.yaml"))
	if err != nil {
		return buildConfig{}, fmt.Errorf("load tool metadata: %w", err)
	}
	bicepTool, err := findTool(manifest, "bicep")
	if err != nil {
		return buildConfig{}, err
	}
	terraformTool, err := findTool(manifest, "terraform")
	if err != nil {
		return buildConfig{}, err
	}
	if value, ok := w.lookupEnv("BICEP_VERSION"); ok {
		bicepTool.Version = normalizeBicepVersion(value, bicepTool.Version)
	}
	applyChecksumOverride(&bicepTool, w.lookupEnv, "linux_amd64", "BICEP_CHECKSUM_LINUX_AMD64")
	applyChecksumOverride(&bicepTool, w.lookupEnv, "linux_arm64", "BICEP_CHECKSUM_LINUX_ARM64")
	terraformVersion := environmentDefault(w.lookupEnv, "TERRAFORM_VERSION", terraformTool.Version)

	buildType, gcflags := "release", ""
	if debug, ok := w.lookupEnv("DEBUG"); ok && debug != "0" {
		buildType, gcflags = "debug", "all=-N -l"
	}
	releaseChannel := environmentDefault(w.lookupEnv, "REL_CHANNEL", "edge")
	releaseVersion := environmentDefault(w.lookupEnv, "REL_VERSION", "edge")
	chartVersion := environmentDefault(w.lookupEnv, "CHART_VERSION", "0.42.42-dev")
	ldflags := strings.Join([]string{
		"-s", "-w",
		"-X", basePackageName + "/pkg/version.channel=" + releaseChannel,
		"-X", basePackageName + "/pkg/version.release=" + releaseVersion,
		"-X", basePackageName + "/pkg/version.commit=" + gitCommit,
		"-X", basePackageName + "/pkg/version.version=" + gitVersion,
		"-X", basePackageName + "/pkg/version.chartVersion=" + chartVersion,
		"-X", basePackageName + "/pkg/recipes/terraform.terraformVersion=" + terraformVersion,
	}, " ")

	outputRoot := environmentDefault(w.lookupEnv, "OUT_DIR", filepath.Join(w.root, "dist"))
	if !filepath.IsAbs(outputRoot) {
		outputRoot = filepath.Join(w.root, outputRoot)
	}
	outputDir := filepath.Join(outputRoot, goos+"_"+goarch, buildType)

	commandEnv := map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        goos,
		"GOARCH":      goarch,
	}
	setEnvironmentDefault(commandEnv, w.lookupEnv, "GO111MODULE", "on")
	setEnvironmentDefault(commandEnv, w.lookupEnv, "GOPROXY", "https://proxy.golang.org")
	setEnvironmentDefault(commandEnv, w.lookupEnv, "GOSUMDB", "sum.golang.org")

	return buildConfig{
		goos:             goos,
		goarch:           goarch,
		buildType:        buildType,
		gcflags:          gcflags,
		ldflags:          ldflags,
		outputDir:        outputDir,
		commandEnv:       commandEnv,
		releaseChannel:   releaseChannel,
		bicepTool:        bicepTool,
		terraformVersion: terraformVersion,
	}, nil
}

func (w *workflows) targetEnvironmentOrOutput(ctx context.Context, launcherKey, key, name string, args ...string) (string, error) {
	if value, ok := w.lookupEnv(launcherKey); ok {
		return value, nil
	}
	return w.environmentOrOutput(ctx, key, name, args...)
}

func (w *workflows) environmentOrOutput(ctx context.Context, key, name string, args ...string) (string, error) {
	if value, ok := w.lookupEnv(key); ok {
		return value, nil
	}
	output, err := w.runner.Output(ctx, command{name: name, args: args, dir: w.root})
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", key, err)
	}
	value := strings.TrimSpace(output)
	if value == "" {
		return "", fmt.Errorf("resolve %s: command returned no output", key)
	}
	return value, nil
}

func environmentDefault(lookup func(string) (string, bool), key, fallback string) string {
	if value, ok := lookup(key); ok {
		return value
	}
	return fallback
}

func setEnvironmentDefault(environment map[string]string, lookup func(string) (string, bool), key, fallback string) {
	environment[key] = environmentDefault(lookup, key, fallback)
}

func findTool(manifest tooling.Manifest, name string) (tooling.Tool, error) {
	for _, tool := range manifest.Tools {
		if tool.Name == name {
			return tool, nil
		}
	}
	return tooling.Tool{}, fmt.Errorf("tool %q is missing from build/tools.yaml", name)
}

func normalizeBicepVersion(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if value[0] >= '0' && value[0] <= '9' {
		return "v" + value
	}
	return value
}

func applyChecksumOverride(tool *tooling.Tool, lookup func(string) (string, bool), platform, key string) {
	value, ok := lookup(key)
	if !ok {
		return
	}
	entry, ok := tool.Platforms[platform]
	if !ok {
		return
	}
	entry.Checksum = value
	tool.Platforms[platform] = entry
}
