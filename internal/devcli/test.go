package devcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	helmUnitPluginURL     = "https://github.com/helm-unittest/helm-unittest.git"
	helmUnitPluginVersion = "1.1.1"
)

func (w *workflows) Test(ctx context.Context, gotestsumArgs, goTestArgs []string) error {
	kubernetesVersion := environmentDefault(w.lookupEnv, "K8S_VERSION", "1.30.*")
	envtestArch := environmentDefault(w.lookupEnv, "ENVTEST_ARCH", "amd64")

	fmt.Fprintln(w.stdout, "==> Provisioning envtest assets")
	assetsOutput, err := w.runner.Output(ctx, command{
		name: "go",
		args: []string{"tool", "setup-envtest", "use", "-p", "path", kubernetesVersion, "--arch", envtestArch},
		dir:  w.root,
	})
	if err != nil {
		return fmt.Errorf("provision envtest assets: %w", err)
	}
	assetsPath := strings.TrimSpace(assetsOutput)
	if assetsPath == "" {
		return errors.New("provision envtest assets: setup-envtest returned no path")
	}
	if info, err := os.Stat(assetsPath); err != nil {
		return fmt.Errorf("inspect envtest assets at %s: %w", assetsPath, err)
	} else if !info.IsDir() {
		return fmt.Errorf("inspect envtest assets at %s: not a directory", assetsPath)
	}

	fmt.Fprintln(w.stdout, "==> Installing helm-unittest plugin if not already installed")
	pluginOutput, err := w.runner.Output(ctx, command{name: "helm", args: []string{"plugin", "list"}, dir: w.root})
	if err != nil {
		return fmt.Errorf("list Helm plugins: %w", err)
	}
	if !helmPluginInstalled(pluginOutput, "unittest") {
		if err := w.runner.Run(ctx, command{
			name: "helm",
			args: []string{"plugin", "install", helmUnitPluginURL, "--version", helmUnitPluginVersion, "--verify=false"},
			dir:  w.root,
		}); err != nil {
			return fmt.Errorf("install helm-unittest plugin: %w", err)
		}
	}

	fmt.Fprintln(w.stdout, "==> Running Helm unit tests")
	if err := w.runner.Run(ctx, command{
		name: "helm",
		args: []string{"unittest", "."},
		dir:  filepath.Join(w.root, "deploy", "Chart"),
	}); err != nil {
		return fmt.Errorf("run Helm unit tests: %w", err)
	}

	fmt.Fprintln(w.stdout, "==> Running manage-radius-installation behavior checks")
	if err := w.manageChecks(ctx); err != nil {
		return fmt.Errorf("run manage-radius-installation behavior checks: %w", err)
	}

	fmt.Fprintln(w.stdout, "==> Running Go package and validation tests")
	args := []string{"tool", "gotestsum"}
	args = append(args, gotestsumArgs...)
	args = append(args, "--", "./pkg/...", "./test/validation/...")
	args = append(args, goTestArgs...)
	if err := w.runner.Run(ctx, command{
		name: "go",
		args: args,
		dir:  w.root,
		env: map[string]string{
			"KUBEBUILDER_ASSETS": assetsPath,
			"CGO_ENABLED":        "1",
		},
	}); err != nil {
		return fmt.Errorf("run Go tests: %w", err)
	}
	return nil
}

func helmPluginInstalled(output, name string) bool {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && strings.EqualFold(fields[0], name) {
			return true
		}
	}
	return false
}

type installationCommands interface {
	InspectDeployment(context.Context, string) (bool, error)
	Upgrade(context.Context, []string) error
	ListResources(context.Context) ([]byte, error)
}

func reconcileRequiredChartValues(ctx context.Context, commands installationCommands) error {
	enabled, err := awsIRSAVolumesEnabled(ctx, commands)
	if err != nil {
		return err
	}
	if enabled {
		return nil
	}

	args := []string{
		"upgrade", "kubernetes",
		"--set", "global.azureWorkloadIdentity.enabled=true",
		"--set", "global.aws.irsa.enabled=true",
		"--set", "database.enabled=false",
		"--skip-preflight",
	}
	if err := commands.Upgrade(ctx, args); err != nil {
		return fmt.Errorf("reconcile required chart values: %w", err)
	}
	enabled, err = awsIRSAVolumesEnabled(ctx, commands)
	if err != nil {
		return err
	}
	if !enabled {
		return errors.New("AWS IRSA token volumes are still missing after reconciliation")
	}
	return nil
}

func awsIRSAVolumesEnabled(ctx context.Context, commands installationCommands) (bool, error) {
	for _, deployment := range []string{"ucp", "applications-rp", "dynamic-rp"} {
		enabled, err := commands.InspectDeployment(ctx, deployment)
		if err != nil {
			return false, fmt.Errorf("inspect deployment %s for AWS IRSA configuration: %w", deployment, err)
		}
		if !enabled {
			return false, nil
		}
	}
	return true, nil
}

func saveSkipResourcesList(ctx context.Context, commands installationCommands, path string) error {
	resources, err := commands.ListResources(ctx)
	if err != nil {
		return fmt.Errorf("retrieve UCP resources: %w", err)
	}
	if len(resources) == 0 {
		return errors.New("retrieve UCP resources: empty response")
	}
	if err := os.WriteFile(path, resources, 0o644); err != nil {
		return fmt.Errorf("write skip resources list: %w", err)
	}
	return nil
}

func runManageInstallationBehaviorChecks(ctx context.Context, output io.Writer) error {
	scenarios := []struct {
		name          string
		initialState  string
		expectUpgrade bool
		expectError   bool
	}{
		{name: "IRSA absent", initialState: "false", expectUpgrade: true},
		{name: "IRSA enabled", initialState: "true"},
		{name: "inspection error", initialState: "error", expectError: true},
	}
	for _, scenario := range scenarios {
		commands := &fakeInstallationCommands{irsaState: scenario.initialState}
		err := reconcileRequiredChartValues(ctx, commands)
		if scenario.expectError {
			if err == nil {
				return fmt.Errorf("%s: expected deployment inspection failure", scenario.name)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("%s: %w", scenario.name, err)
		}
		if scenario.expectUpgrade != (len(commands.upgradeArgs) > 0) {
			return fmt.Errorf("%s: upgrade expectation was not met", scenario.name)
		}
		for _, required := range []string{
			"--skip-preflight",
			"global.azureWorkloadIdentity.enabled=true",
			"global.aws.irsa.enabled=true",
			"database.enabled=false",
		} {
			if scenario.expectUpgrade && !slices.Contains(commands.upgradeArgs, required) {
				return fmt.Errorf("%s: upgrade arguments do not contain %q", scenario.name, required)
			}
		}

		workDir, err := os.MkdirTemp("", "radius-manage-installation-*")
		if err != nil {
			return fmt.Errorf("%s: create temporary directory: %w", scenario.name, err)
		}
		path := filepath.Join(workDir, "skip-delete-resources-list.txt")
		saveErr := saveSkipResourcesList(ctx, commands, path)
		removeErr := os.RemoveAll(workDir)
		if saveErr != nil {
			return fmt.Errorf("%s: %w", scenario.name, saveErr)
		}
		if removeErr != nil {
			return fmt.Errorf("%s: clean temporary directory: %w", scenario.name, removeErr)
		}
	}
	fmt.Fprintln(output, "manage-radius-installation tests passed")
	return nil
}

type fakeInstallationCommands struct {
	irsaState   string
	upgradeArgs []string
}

func (f *fakeInstallationCommands) InspectDeployment(context.Context, string) (bool, error) {
	switch f.irsaState {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("deployment inspection failed")
	}
}

func (f *fakeInstallationCommands) Upgrade(_ context.Context, args []string) error {
	f.upgradeArgs = append([]string(nil), args...)
	f.irsaState = "true"
	return nil
}

func (f *fakeInstallationCommands) ListResources(context.Context) ([]byte, error) {
	return []byte("resource-id\n"), nil
}
