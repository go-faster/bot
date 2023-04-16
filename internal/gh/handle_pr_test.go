package gh

import (
	"context"
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/go-faster/bot/internal/ent/enttest"
	"github.com/go-faster/errors"
	"github.com/go-faster/simon/sdk/zctx"
	"github.com/google/go-github/v50/github"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zaptest"
)

type mockResolver map[string]tg.InputPeerClass

func (m mockResolver) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	f, ok := m[domain]
	if !ok {
		return nil, tgerr.New(400, tg.ErrUsernameInvalid)
	}
	return f, nil
}

func (m mockResolver) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	f, ok := m[phone]
	if !ok {
		return nil, tgerr.New(400, tg.ErrUsernameInvalid)
	}
	return f, nil
}

func prEvent(prID int, orgID int64) *github.PullRequestEvent {
	return &github.PullRequestEvent{
		PullRequest: &github.PullRequest{
			Merged: github.Bool(true),
			Number: &prID,
		},
		Repo: &github.Repository{
			ID:   &orgID,
			Name: github.String("test"),
		},
	}
}

type mockInvoker struct {
	lastReq *tg.MessagesEditMessageRequest
}

func (m *mockInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	req, ok := input.(*tg.MessagesEditMessageRequest)
	if !ok {
		return errors.Errorf("unexpected type %T", input)
	}
	m.lastReq = req
	return nil
}

func TestWebhook(t *testing.T) {
	lg := zaptest.NewLogger(t)

	ctx := context.Background()
	a := require.New(t)

	msgID, lastMsgID := 10, 11
	prID, orgID := 13, int64(37)
	channel := &tg.InputPeerChannel{
		ChannelID:  69,
		AccessHash: 42,
	}
	event := prEvent(prID, orgID)

	invoker := &mockInvoker{}
	raw := tg.NewClient(invoker)
	sender := message.NewSender(raw).WithResolver(mockResolver{
		"test": channel,
	})

	db := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&_fk=1")
	defer db.Close()

	hook := NewWebhook(
		db,
		sender,
		metric.NewNoopMeterProvider(),
		trace.NewNoopTracerProvider(),
	).WithNotifyGroup("test")

	a.NoError(hook.updateLastMsgID(ctx, channel.ChannelID, lastMsgID))
	a.NoError(hook.setPRNotification(ctx, event.GetRepo(), event.GetPullRequest(), msgID))

	a.NoError(hook.handlePRClosed(zctx.With(ctx, lg), event))
	a.NotNil(invoker.lastReq)
	a.Contains(invoker.lastReq.Message, "opened")
}

func Test_generateChecksStatus(t *testing.T) {
	tests := []struct {
		checks []Check
		want   string
	}{
		{nil, ""},
		{[]Check{}, ""},

		{
			[]Check{
				{Status: "created"},
				{Status: "created"},
				{Status: "created"},
				{Status: "completed", Conclusion: "success"},
			},
			"3🟡,1🟢/4",
		},
		{
			[]Check{
				{Status: "completed", Conclusion: "failure"},
				{Status: "completed", Conclusion: "timed_out"},
				{Status: "completed", Conclusion: "cancelled"},
				{Status: "completed", Conclusion: "success"},
			},
			"3🔴,1🟢/4",
		},
		{
			[]Check{
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "success"},
				{Status: "completed", Conclusion: "success"},
			},
			"3🟢/3",
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.Equal(t, tt.want, generateChecksStatus(tt.checks))
		})
	}
}
