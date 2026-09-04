// Package devcli implements the repository-owned Radius development CLI.
package devcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

type workflow interface {
	Build(context.Context) error
	Test(context.Context, []string, []string) error
}

// CLI dispatches development commands.
type CLI struct {
	workflow workflow
}

// New creates a CLI rooted at the Radius repository.
func New(root string, stdout, stderr io.Writer) *CLI {
	runner := execRunner{stdout: stdout, stderr: stderr}
	workflows := &workflows{
		root:        root,
		stdout:      stdout,
		runner:      runner,
		lookupEnv:   os.LookupEnv,
		bicepStager: fileBicepStager{},
	}
	workflows.manageChecks = func(ctx context.Context) error {
		if runtime.GOOS == "windows" {
			return runManageInstallationBehaviorChecks(ctx, stdout)
		}
		return runner.Run(ctx, command{
			name: "bash",
			args: []string{"./.github/scripts/manage-radius-installation_test.sh"},
			dir:  root,
		})
	}
	return &CLI{workflow: workflows}
}

// Run executes a development command.
func (c *CLI) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("a command is required: build or test")
	}

	switch args[0] {
	case "build":
		if len(args) != 1 {
			return fmt.Errorf("build does not accept arguments: %v", args[1:])
		}
		return c.workflow.Build(ctx)
	case "test":
		gotestsumArgs, goTestArgs := splitTestArgs(args[1:])
		return c.workflow.Test(ctx, gotestsumArgs, goTestArgs)
	default:
		return fmt.Errorf("unknown command %q: expected build or test", args[0])
	}
}

func splitTestArgs(args []string) ([]string, []string) {
	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

// FindRepositoryRoot locates the closest parent directory containing the Radius Go module.
func FindRepositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect repository root: %w", err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find repository root containing go.mod")
		}
		dir = parent
	}
}
