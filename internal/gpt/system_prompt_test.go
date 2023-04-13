package gpt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultContextPrompt(t *testing.T) {
	var sb strings.Builder
	require.NoError(t, defaultContextPrompt.Execute(&sb, ContextPromptData{
		Prompter: PromptUser{
			Username:  "catent",
			FirstName: "Aleksandr",
		},
		ChatTitle: "go faster chat",
	}))
	require.Equal(t, `Chat title is: "go faster chat"
My nickname is: "catent"
My name is: "Aleksandr"
`, sb.String())
}
