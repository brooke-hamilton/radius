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

package delete

import (
	"context"
	"errors"
	"fmt"

	"github.com/project-radius/radius/pkg/cli"
	"github.com/project-radius/radius/pkg/cli/cmd/commonflags"
	"github.com/project-radius/radius/pkg/cli/framework"
	"github.com/project-radius/radius/pkg/cli/output"
	"github.com/project-radius/radius/pkg/cli/prompt"
	"github.com/project-radius/radius/pkg/cli/workspaces"
	"github.com/spf13/cobra"
)

const (
	deleteConfirmationFmt = "Are you sure you want to delete workspace '%v' from local config? This will update config but will not delete any deployed resources."
)

// NewCommand creates an instance of the command and runner for the `rad workspace delete` command.
func NewCommand(factory framework.Factory) (*cobra.Command, framework.Runner) {
	runner := NewRunner(factory)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete local workspace",
		Long:  `Delete local workspace`,
		Example: `# Delete current workspace
rad workspace delete

# Delete named workspace
rad workspace delete my-workspace`,
		Args: cobra.RangeArgs(0, 1),
		RunE: framework.RunCommand(runner),
	}

	commonflags.AddWorkspaceFlag(cmd)
	commonflags.AddConfirmationFlag(cmd)

	return cmd, runner
}

// Runner is the runner implementation for the `rad workspace delete` command.
type Runner struct {
	ConfigHolder        *framework.ConfigHolder
	ConfigFileInterface framework.ConfigFileInterface
	Output              output.Interface
	InputPrompter       prompt.Interface
	Workspace           *workspaces.Workspace
	Confirm             bool
}

// NewRunner creates a new instance of the `rad workspace delete` runner.
func NewRunner(factory framework.Factory) *Runner {
	return &Runner{
		ConfigFileInterface: factory.GetConfigFileInterface(),
		ConfigHolder:        factory.GetConfigHolder(),
		InputPrompter:       factory.GetPrompter(),
		Output:              factory.GetOutput(),
	}
}

// Validate runs validation for the `rad workspace delete` command.
func (r *Runner) Validate(cmd *cobra.Command, args []string) error {
	workspace, err := cli.RequireWorkspaceArgs(cmd, r.ConfigHolder.Config, args)
	if err != nil {
		return err
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	r.Workspace = workspace
	r.Confirm = yes

	if !r.Workspace.IsNamedWorkspace() {
		// Only workspaces stored in configuration can be deleted.
		return workspaces.ErrNamedWorkspaceRequired
	}

	return nil
}

// Run runs the `rad workspace delete` command.
func (r *Runner) Run(ctx context.Context) error {
	// Prompt user to confirm deletion
	if !r.Confirm {
		message := fmt.Sprintf(deleteConfirmationFmt, r.Workspace.Name)
		confirmed, err := prompt.YesOrNoPrompt(message, prompt.ConfirmNo, r.InputPrompter)
		if err != nil {
			if errors.Is(err, &prompt.ErrExitConsole{}) {
				return &cli.FriendlyError{Message: err.Error()}
			}
			return err
		}

		if !confirmed {
			return nil
		}
	}

	err := r.ConfigFileInterface.DeleteWorkspace(ctx, r.ConfigHolder.Config, r.Workspace.Name)
	if err != nil {
		return err
	}

	return nil
}
