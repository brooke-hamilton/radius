package devcli

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeWorkflow struct {
	buildCalls int
	sumArgs    []string
	goArgs     []string
}

func (f *fakeWorkflow) Build(context.Context) error {
	f.buildCalls++
	return nil
}

func (f *fakeWorkflow) Test(_ context.Context, sumArgs, goArgs []string) error {
	f.sumArgs = append([]string(nil), sumArgs...)
	f.goArgs = append([]string(nil), goArgs...)
	return nil
}

func TestCLI_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantError string
		assert    func(*testing.T, *fakeWorkflow)
	}{
		{
			name: "build",
			args: []string{"build"},
			assert: func(t *testing.T, workflow *fakeWorkflow) {
				require.Equal(t, 1, workflow.buildCalls)
			},
		},
		{
			name: "test forwards both option channels",
			args: []string{"test", "--junitfile", "results.xml", "--", "-race", "-count=1"},
			assert: func(t *testing.T, workflow *fakeWorkflow) {
				require.Equal(t, []string{"--junitfile", "results.xml"}, workflow.sumArgs)
				require.Equal(t, []string{"-race", "-count=1"}, workflow.goArgs)
			},
		},
		{
			name: "test forwards gotestsum options without separator",
			args: []string{"test", "--format", "testname"},
			assert: func(t *testing.T, workflow *fakeWorkflow) {
				require.Equal(t, []string{"--format", "testname"}, workflow.sumArgs)
				require.Empty(t, workflow.goArgs)
			},
		},
		{name: "missing command", wantError: "a command is required"},
		{name: "unknown command", args: []string{"lint"}, wantError: `unknown command "lint"`},
		{name: "build rejects arguments", args: []string{"build", "--debug"}, wantError: "build does not accept arguments"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			workflow := &fakeWorkflow{}
			err := (&CLI{workflow: workflow}).Run(context.Background(), test.args)
			if test.wantError != "" {
				require.ErrorContains(t, err, test.wantError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, test.assert)
			test.assert(t, workflow)
		})
	}
}

func TestFindRepositoryRoot(t *testing.T) {
	root, err := FindRepositoryRoot()
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(root, "go.mod"))
}

func TestCLI_PropagatesWorkflowError(t *testing.T) {
	t.Parallel()

	expected := errors.New("build failed")
	workflow := errorWorkflow{err: expected}
	err := (&CLI{workflow: workflow}).Run(context.Background(), []string{"build"})
	require.ErrorIs(t, err, expected)
}

type errorWorkflow struct {
	err error
}

func (e errorWorkflow) Build(context.Context) error {
	return e.err
}

func (e errorWorkflow) Test(context.Context, []string, []string) error {
	return e.err
}
