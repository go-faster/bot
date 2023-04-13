package gpt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultContextPrompt(t *testing.T) {
	var sb strings.Builder
	defaultContextPrompt.Execute(&sb, ContextPromptData{
		Prompter: PromptUser{
			Username:  "catent",
			FirstName: "Aleksandr",
		},
		ChatTitle: "go faster chat",
	})
	require.Equal(t, `Chat title is: "go faster chat"
User's nickname is: "catent"
User's name is: "Aleksandr"
`, sb.String())
}
