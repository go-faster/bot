package dispatch

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestMessageMux_OnMessage(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mux := NewMessageMux()

	cmdCalls := 0
	mux.HandleFunc("/github", "test", func(ctx context.Context, e MessageEvent) error {
		cmdCalls++
		return nil
	})

	fallbackCalls := 0
	mux.SetFallbackFunc(func(ctx context.Context, e MessageEvent) error {
		fallbackCalls++
		return nil
	})

	send := func(text string) {
		a.NoError(mux.OnMessage(ctx, MessageEvent{
			Message: &tg.Message{
				Message: text,
			},
		}))
	}
	send("github/")
	send("github/gotd")
	send("github/gotd/td")
	a.Zero(cmdCalls)
	a.Equal(fallbackCalls, 3)

	send("/github")
	a.Equal(1, cmdCalls)
	a.Equal(fallbackCalls, 3)
}
