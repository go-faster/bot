package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction(t *testing.T) {
	a := Action{
		Entity:       PullRequest,
		ID:           1,
		RepositoryID: 2,
		Type:         Merge,
	}

	data, err := a.MarshalText()
	require.NoError(t, err)
	require.Less(t, len(data), 50)
	t.Logf("data=%s [%s]", data, a.String())

	for _, buf := range [][]byte{
		data,
		Marshal(a),
	} {
		var out Action
		require.NoError(t, out.UnmarshalText(buf))
		require.Equal(t, a, out)
		require.True(t, out.Is(Merge, PullRequest))
		require.True(t, out.Is(a.Type, a.Entity))
	}
}
