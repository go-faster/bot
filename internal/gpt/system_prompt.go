package gpt

import (
	"text/template"

	"github.com/go-faster/bot/internal/dispatch"
)

// PromptUser defines system prompt data of user asking the question.
type PromptUser struct {
	Username  string
	FirstName string
}

// SystemPromptData is a data structure passed to system prompt template.
type SystemPromptData struct {
	Prompter PromptUser
	// ChatTitle is a chat title where prompt was generated.
	ChatTitle string
}

var defaultSystemPrompt = template.Must(template.New("system_prompt").Parse(`Chat title is: {{ printf "%q" .ChatTitle }}
User's nickname is: {{ printf "%q" .Prompter.Username }}
User's name is: {{ printf "%q" .Prompter.FirstName }}
`))

func generateSystemPromptData(e dispatch.MessageEvent) (data SystemPromptData) {
	if from, ok := e.MessageFrom(); ok {
		data.Prompter = PromptUser{
			Username:  from.Username,
			FirstName: from.Username,
		}
	}

	if ch, ok := e.Channel(); ok {
		data.ChatTitle = ch.Title
	} else if ch, ok := e.Chat(); ok {
		data.ChatTitle = ch.Title
	} else if u, ok := e.User(); ok {
		data.ChatTitle = u.FirstName
		if last := u.LastName; last != "" {
			data.ChatTitle += " " + last
		}
	}
	return data
}
