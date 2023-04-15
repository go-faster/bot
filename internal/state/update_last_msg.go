package state

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/trace"
)

func (e Ent) UpdateLastMsgID(ctx context.Context, channelID int64, msgID int) error {
	ctx, span := e.tracer.Start(ctx, "UpdateLastMsgID",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	if err := e.db.LastChannelMessage.Create().
		SetID(channelID).
		SetMessageID(msgID).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateMessageID().
		Exec(ctx); err != nil {
		return errors.Wrap(err, "upsert")
	}
	return nil
}
