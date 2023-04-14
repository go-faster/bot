package gpt

import (
	"fmt"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
	"github.com/tiktoken-go/tokenizer"
)

func Test_cutDialog(t *testing.T) {
	codec, err := tokenizer.ForModel(tokenizerModel)
	require.NoError(t, err)

	type dialog = []openai.ChatCompletionMessage
	msg := func(content string) (m openai.ChatCompletionMessage) {
		m.Content = content
		return m
	}

	tests := []struct {
		limit  int
		msgs   dialog
		expect dialog
	}{
		{
			10,
			dialog{
				msg("Hello. How are you?"), // 6
			},
			dialog{
				msg("Hello. How are you?"),
			},
		},
		{
			10,
			dialog{
				msg("Hello. How are you?"),     // 6
				msg("Oh, I am fine. And you?"), // 9
				msg("Same."),                   // 2
				msg("Something."),              // 2
				msg("Goodbye."),                // 2
			},
			dialog{
				msg("Same."),
				msg("Something."),
				msg("Goodbye."),
			},
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			got, tokens, err := cutDialog(codec, tt.limit, tt.msgs)
			a.NoError(err)

			a.Greater(tt.limit, tokens)
			a.Equal(tt.expect, got)
		})
	}
}
