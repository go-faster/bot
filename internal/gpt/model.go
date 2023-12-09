package gpt

import (
	"github.com/go-faster/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

const (
	model = openai.GPT4
	// Model token limit.
	modelTokenLimit = 8_192
	// Reserve tokens for model to answer.
	modelAnswerReserve = 1_000
	tokenizerModel     = tokenizer.GPT4
)

func cutDialog(tokenizer tokenizer.Codec, limit int, dialog []openai.ChatCompletionMessage) (_ []openai.ChatCompletionMessage, tokens int, _ error) {
	for i := len(dialog) - 1; i >= 0; i-- {
		msg := dialog[i]
		// FIXME(tdakkota): dramatically inefficient.
		// 	Probably we should fork it and optimize it.
		_, tks, err := tokenizer.Encode(msg.Content)
		if err != nil {
			return nil, 0, errors.Wrap(err, "tokenizer error")
		}
		msgTokens := len(tks)

		if tokens+msgTokens >= limit {
			dialog = dialog[i+1:]
			break
		}
		tokens += msgTokens
	}
	return dialog, tokens, nil
}
