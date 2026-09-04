package devcli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflows_TestOrchestratesAllPhasesAndForwardsArguments(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	assets := filepath.Join(root, "envtest")
	require.NoError(t, ensureDirectory(assets))
	runner := &fakeRunner{outputs: map[string]string{
		"go tool setup-envtest use -p path 1.30.* --arch amd64": assets + "\n",
		"helm plugin list": "NAME\tVERSION\n",
	}}
	manageCalls := 0
	workflows := &workflows{
		root:      root,
		stdout:    &bytes.Buffer{},
		runner:    runner,
		lookupEnv: mapLookup(nil),
		manageChecks: func(context.Context) error {
			manageCalls++
			return nil
		},
	}

	err := workflows.Test(
		context.Background(),
		[]string{"--junitfile", "results.xml"},
		[]string{"-race", "-coverprofile", "coverage.out"},
	)
	require.NoError(t, err)
	require.Equal(t, 1, manageCalls)
	require.Len(t, runner.commands, 5)
	require.Equal(t, "go tool setup-envtest use -p path 1.30.* --arch amd64", commandKey(runner.commands[0]))
	require.Equal(t, "helm plugin list", commandKey(runner.commands[1]))
	require.Equal(t, "helm plugin install "+helmUnitPluginURL+" --version 1.1.1 --verify=false", commandKey(runner.commands[2]))
	require.Equal(t, filepath.Join(root, "deploy", "Chart"), runner.commands[3].dir)
	require.Equal(t, "helm unittest .", commandKey(runner.commands[3]))
	require.Equal(t, "go tool gotestsum --junitfile results.xml -- ./pkg/... ./test/validation/... -race -coverprofile coverage.out", commandKey(runner.commands[4]))
	require.Equal(t, assets, runner.commands[4].env["KUBEBUILDER_ASSETS"])
	require.Equal(t, "1", runner.commands[4].env["CGO_ENABLED"])
}

func TestHelmPluginInstalled(t *testing.T) {
	t.Parallel()

	require.True(t, helmPluginInstalled("NAME VERSION\nunittest 1.1.1\n", "unittest"))
	require.True(t, helmPluginInstalled("UNITTEST 1.1.1\n", "unittest"))
	require.False(t, helmPluginInstalled("NAME VERSION\ndiff 3.0.0\n", "unittest"))
	require.False(t, helmPluginInstalled("my-unittest-helper 1.0.0\n", "unittest"))
}

func TestRunManageInstallationBehaviorChecks(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	require.NoError(t, runManageInstallationBehaviorChecks(context.Background(), &output))
	require.Contains(t, output.String(), "manage-radius-installation tests passed")
}

func ensureDirectory(path string) error {
	return os.MkdirAll(path, 0o755)
}
