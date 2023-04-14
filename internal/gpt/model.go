package gpt

import (
	"github.com/go-faster/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

const (
	model           = openai.GPT3Dot5Turbo
	modelTokenLimit = 4096
	tokenizerModel  = tokenizer.GPT35Turbo
)

// compile time check shenanigans
var _ = map[bool]struct{}{
	model == tokenizerModel: {},
	false:                   {},
}

func cutDialog(tokenizer tokenizer.Codec, limit int, dialog []openai.ChatCompletionMessage) (_ []openai.ChatCompletionMessage, tokens int, _ error) {
	for i := len(dialog) - 1; i >= 0; i-- {
		msg := dialog[i]
		// FIXME(tdakkota): dramatically inefficient.
		// 	Probably we should fork it and optimize it.
		ids, _, err := tokenizer.Encode(msg.Content)
		if err != nil {
			return nil, 0, errors.Wrap(err, "tokenizer error")
		}

		tokens += len(ids)
		if tokens >= limit {
			dialog = dialog[i+1:]
			break
		}
	}
	return dialog, tokens, nil
}
