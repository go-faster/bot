package gpt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultSystemPrompt(t *testing.T) {
	var sb strings.Builder
	defaultSystemPrompt.Execute(&sb, SystemPromptData{
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
