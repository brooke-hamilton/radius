package devcli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeEnvironment_ReplacesKeysCaseInsensitively(t *testing.T) {
	t.Parallel()

	merged := mergeEnvironment(
		[]string{"PATH=old", "KEEP=value", "Malformed"},
		map[string]string{"Path": "new", "CGO_ENABLED": "0"},
	)

	require.ElementsMatch(t, []string{"KEEP=value", "Path=new", "CGO_ENABLED=0"}, merged)
}
