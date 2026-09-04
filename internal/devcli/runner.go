package devcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/radius-project/radius/pkg/process"
)

type command struct {
	name string
	args []string
	dir  string
	env  map[string]string
}

type runner interface {
	Run(context.Context, command) error
	Output(context.Context, command) (string, error)
}

type execRunner struct {
	stdout io.Writer
	stderr io.Writer
}

func (r execRunner) Run(ctx context.Context, invocation command) error {
	cmd := process.CommandContext(ctx, invocation.name, invocation.args...)
	cmd.Dir = invocation.dir
	cmd.Env = mergeEnvironment(os.Environ(), invocation.env)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", invocation.name, err)
	}
	return nil
}

func (r execRunner) Output(ctx context.Context, invocation command) (string, error) {
	cmd := process.CommandContext(ctx, invocation.name, invocation.args...)
	cmd.Dir = invocation.dir
	cmd.Env = mergeEnvironment(os.Environ(), invocation.env)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = r.stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run %s: %w", invocation.name, err)
	}
	return stdout.String(), nil
}

func mergeEnvironment(base []string, overrides map[string]string) []string {
	merged := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		key, _, found := strings.Cut(entry, "=")
		if !found || containsEnvironmentKey(overrides, key) {
			continue
		}
		merged = append(merged, entry)
	}
	for key, value := range overrides {
		merged = append(merged, key+"="+value)
	}
	return merged
}

func containsEnvironmentKey(environment map[string]string, key string) bool {
	for candidate := range environment {
		if strings.EqualFold(candidate, key) {
			return true
		}
	}
	return false
}
