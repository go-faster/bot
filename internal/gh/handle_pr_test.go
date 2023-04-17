package gh

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateChecksStatus(t *testing.T) {
	tests := []struct {
		checks []Check
		want   string
	}{
		{nil, ""},
		{[]Check{}, ""},

		{
			[]Check{
				{Status: "created"},
				{Status: "created"},
				{Status: "created"},
				{Status: "completed", Conclusion: "success"},
			},
			"3🟡,1🟢/4",
		},
		{
			[]Check{
				{Status: "completed", Conclusion: "failure"},
				{Status: "completed", Conclusion: "timed_out"},
				{Status: "completed", Conclusion: "cancelled"},
				{Status: "completed", Conclusion: "success"},
			},
			"3🔴,1🟢/4",
		},
		{
			[]Check{
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "success"},
			},
			"3🟢/3",
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.Equal(t, tt.want, generateChecksStatus(tt.checks))
		})
	}
}
