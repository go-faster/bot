package gh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractMeta(t *testing.T) {
	_, _ = extractEventMeta([]byte(`{}`))
	type testCase struct {
		name   string
		output eventMeta
	}
	for _, tc := range []testCase{
		{
			name: "event.status.json",
			output: eventMeta{
				Organization:       "go-faster",
				OrganizationID:     93744681,
				Repository:         "yaml",
				RepositoryID:       512150878,
				RepositoryFullName: "go-faster/yaml",
			},
		},
		{
			name: "event.workflow.job.completed.json",
			output: eventMeta{
				Organization:       "go-faster",
				OrganizationID:     93744681,
				Repository:         "yaml",
				RepositoryID:       512150878,
				RepositoryFullName: "go-faster/yaml",
			},
		},
		{
			name: "event.workflow.run.json",
			output: eventMeta{
				Organization:       "go-faster",
				OrganizationID:     93744681,
				Repository:         "yaml",
				RepositoryID:       512150878,
				RepositoryFullName: "go-faster/yaml",
			},
		},
		{
			name: "event.check.run.completed.json",
			output: eventMeta{
				Organization:       "go-faster",
				OrganizationID:     93744681,
				Repository:         "yaml",
				RepositoryID:       512150878,
				RepositoryFullName: "go-faster/yaml",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("_testdata", tc.name))
			require.NoError(t, err, "no file")
			v, err := extractEventMeta(data)
			require.NoError(t, err, "no error")
			require.Equal(t, tc.output, *v, "output")
		})
	}
}
