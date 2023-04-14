package gh

import "testing"

func TestExtractMeta(t *testing.T) {
	_, _ = extractEventMeta([]byte(`{}`))
}
