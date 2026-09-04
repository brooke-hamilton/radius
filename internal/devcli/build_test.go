package devcli

import (
	"context"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeRunner struct {
	outputs  map[string]string
	commands []command
}

func (f *fakeRunner) Run(_ context.Context, invocation command) error {
	f.commands = append(f.commands, invocation)
	return nil
}

func (f *fakeRunner) Output(_ context.Context, invocation command) (string, error) {
	f.commands = append(f.commands, invocation)
	return f.outputs[commandKey(invocation)], nil
}

func commandKey(invocation command) string {
	return invocation.name + " " + strings.Join(invocation.args, " ")
}

type fakeBicepStager struct {
	config bicepStageConfig
}

func (f *fakeBicepStager) Stage(_ context.Context, config bicepStageConfig) error {
	f.config = config
	return nil
}

func TestWorkflows_BuildPlansEquivalentCommands(t *testing.T) {
	t.Parallel()

	root, err := FindRepositoryRoot()
	require.NoError(t, err)
	runner := &fakeRunner{outputs: map[string]string{
		"go env GOOS":          "windows\n",
		"go env GOARCH":        "amd64\n",
		"git rev-list -1 HEAD": "abc123\n",
		"git describe --always --abbrev=7 --dirty --tags": "v1.2.3-4-gabc123\n",
	}}
	stager := &fakeBicepStager{}
	environment := map[string]string{
		"DEBUG":             "1",
		"REL_CHANNEL":       "rc",
		"REL_VERSION":       "1.2.3",
		"CHART_VERSION":     "1.2.3",
		"TERRAFORM_VERSION": "9.9.9",
	}
	workflows := &workflows{
		root:        root,
		stdout:      io.Discard,
		runner:      runner,
		lookupEnv:   mapLookup(environment),
		bicepStager: stager,
	}

	require.NoError(t, workflows.Build(context.Background()))

	require.Len(t, runner.commands, 14)
	packageBuild := runner.commands[4]
	require.Equal(t, "go", packageBuild.name)
	require.Equal(t, []string{"build", "-v", "-gcflags", "all=-N -l", packageBuild.args[4], "./..."}, packageBuild.args)
	require.Contains(t, packageBuild.args[4], "-ldflags=-s -w")
	require.Contains(t, packageBuild.args[4], "pkg/version.channel=rc")
	require.Contains(t, packageBuild.args[4], "pkg/version.release=1.2.3")
	require.Contains(t, packageBuild.args[4], "pkg/version.commit=abc123")
	require.Contains(t, packageBuild.args[4], "pkg/recipes/terraform.terraformVersion=9.9.9")
	require.Equal(t, "0", packageBuild.env["CGO_ENABLED"])
	require.Equal(t, "windows", packageBuild.env["GOOS"])
	require.Equal(t, "amd64", packageBuild.env["GOARCH"])

	binaryBuilds := runner.commands[5:]
	require.Len(t, binaryBuilds, len(binaries))
	for i, binary := range binaries {
		require.Equal(t, filepath.Join(root, filepath.FromSlash(binary.entrypoint)), binaryBuilds[i].dir)
		outputIndex := slices.Index(binaryBuilds[i].args, "-o")
		require.NotEqual(t, -1, outputIndex)
		require.Equal(t, filepath.Join(root, "dist", "windows_amd64", "debug", binary.name+".exe"), binaryBuilds[i].args[outputIndex+1])
	}
	require.Equal(t, filepath.Join(root, "dist", "windows_amd64", "debug", "bicep"), stager.config.outputDir)
	require.Equal(t, "rc", stager.config.channel)
	require.Equal(t, "amd64", stager.config.arch)
	require.Equal(t, "v0.42.1", stager.config.tool.Version)
}

func TestResolveBuildConfig_ReleaseDefaultsAndOverrides(t *testing.T) {
	t.Parallel()

	root, err := FindRepositoryRoot()
	require.NoError(t, err)
	runner := &fakeRunner{outputs: map[string]string{
		"git rev-list -1 HEAD":                            "commit\n",
		"git describe --always --abbrev=7 --dirty --tags": "description\n",
	}}
	workflows := &workflows{
		root:   root,
		runner: runner,
		lookupEnv: mapLookup(map[string]string{
			"GOOS":                       "linux",
			"GOARCH":                     "arm64",
			"DEBUG":                      "0",
			"BICEP_VERSION":              "9.8.7",
			"BICEP_CHECKSUM_LINUX_ARM64": "override-checksum",
		}),
	}

	config, err := workflows.resolveBuildConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "release", config.buildType)
	require.Empty(t, config.gcflags)
	require.Equal(t, filepath.Join(root, "dist", "linux_arm64", "release"), config.outputDir)
	require.Equal(t, "v9.8.7", config.bicepTool.Version)
	require.Equal(t, "override-checksum", config.bicepTool.Platforms["linux_arm64"].Checksum)
	require.Equal(t, "edge", config.releaseChannel)
	require.Equal(t, "0", config.commandEnv["CGO_ENABLED"])
	require.Equal(t, "on", config.commandEnv["GO111MODULE"])
}

func TestNormalizeBicepVersion(t *testing.T) {
	t.Parallel()

	require.Equal(t, "v1.2.3", normalizeBicepVersion("1.2.3", "v0.1.0"))
	require.Equal(t, "v1.2.3", normalizeBicepVersion(" v1.2.3 ", "v0.1.0"))
	require.Equal(t, "v0.1.0", normalizeBicepVersion("", "v0.1.0"))
}

func TestResolveBuildConfig_PrefersLauncherTarget(t *testing.T) {
	t.Parallel()

	root, err := FindRepositoryRoot()
	require.NoError(t, err)
	runner := &fakeRunner{outputs: map[string]string{
		"git rev-list -1 HEAD":                            "commit\n",
		"git describe --always --abbrev=7 --dirty --tags": "description\n",
	}}
	workflows := &workflows{
		root:   root,
		runner: runner,
		lookupEnv: mapLookup(map[string]string{
			"RADIUS_DEV_GOOS":   "linux",
			"RADIUS_DEV_GOARCH": "arm64",
			"GOOS":              "windows",
			"GOARCH":            "amd64",
		}),
	}

	config, err := workflows.resolveBuildConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "linux", config.goos)
	require.Equal(t, "arm64", config.goarch)
	require.Equal(t, filepath.Join(root, "dist", "linux_arm64", "release"), config.outputDir)
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
