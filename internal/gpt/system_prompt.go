package gpt

import (
	"text/template"

	"github.com/go-faster/bot/internal/dispatch"
)

// PromptUser defines context prompt data of user asking the question.
type PromptUser struct {
	Username  string
	FirstName string
}

// ContextPromptData is a data structure passed to context prompt template.
type ContextPromptData struct {
	Prompter PromptUser
	// ChatTitle is a chat title where prompt was generated.
	ChatTitle string
}

var defaultContextPrompt = template.Must(template.New("context_prompt").Parse(`Chat title is: {{ printf "%q" .ChatTitle }}
My nickname is: {{ printf "%q" .Prompter.Username }}
My name is: {{ printf "%q" .Prompter.FirstName }}
`))

func generateContextPromptData(e dispatch.MessageEvent) (data ContextPromptData) {
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
