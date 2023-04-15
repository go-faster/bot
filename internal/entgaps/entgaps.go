package entgaps

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/td/telegram/updates"

	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/telegramchannelstate"
)

var _ updates.StateStorage = (*State)(nil)

type State struct {
	db     *ent.Client
	tracer trace.Tracer
}

func New(client *ent.Client, traceProvider trace.TracerProvider) *State {
	return &State{
		db:     client,
		tracer: traceProvider.Tracer("gaps"),
	}
}

func (s *State) GetState(ctx context.Context, userID int64) (state updates.State, found bool, err error) {
	ctx, span := s.tracer.Start(ctx, "GetState",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()
	v, err := s.db.TelegramUserState.Get(ctx, userID)
	if ent.IsNotFound(err) {
		return updates.State{}, false, nil
	}
	if v.Pts == 0 || v.Qts == 0 {
		return updates.State{}, false, nil
	}
	return updates.State{
		Pts:  v.Pts,
		Qts:  v.Qts,
		Date: v.Date,
		Seq:  v.Seq,
	}, true, nil
}

func (s *State) SetState(ctx context.Context, userID int64, state updates.State) error {
	ctx, span := s.tracer.Start(ctx, "SetState",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
			attribute.Int("pts", state.Pts),
			attribute.Int("qts", state.Qts),
			attribute.Int("date", state.Date),
			attribute.Int("seq", state.Seq),
		),
	)
	defer span.End()
	if err := s.db.TelegramUserState.Create().
		SetID(userID).
		SetPts(state.Pts).SetQts(state.Qts).SetDate(state.Date).SetSeq(state.Seq).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx); err != nil {
		return fmt.Errorf("upsert state: %w", err)
	}
	return nil
}

func (s *State) SetPts(ctx context.Context, userID int64, pts int) error {
	ctx, span := s.tracer.Start(ctx, "SetPts",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()
	return s.db.TelegramUserState.Create().SetID(userID).SetPts(pts).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx)

}

func (s *State) SetQts(ctx context.Context, userID int64, qts int) error {
	ctx, span := s.tracer.Start(ctx, "SetQts",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()
	return s.db.TelegramUserState.Create().SetID(userID).SetQts(qts).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx)
}

func (s *State) SetDate(ctx context.Context, userID int64, date int) error {
	ctx, span := s.tracer.Start(ctx, "SetDate",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()
	return s.db.TelegramUserState.Create().SetID(userID).SetDate(date).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx)
}

func (s *State) SetSeq(ctx context.Context, userID int64, seq int) error {
	ctx, span := s.tracer.Start(ctx, "SetSeq",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()
	return s.db.TelegramUserState.Create().SetID(userID).SetSeq(seq).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx)
}

func (s *State) SetDateSeq(ctx context.Context, userID int64, date, seq int) error {
	ctx, span := s.tracer.Start(ctx, "SetDateSeq",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()
	return s.db.TelegramUserState.Create().SetID(userID).SetDate(date).SetSeq(seq).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx)
}

func (s *State) GetChannelPts(ctx context.Context, userID, channelID int64) (int, bool, error) {
	ctx, span := s.tracer.Start(ctx, "GetChannelPts",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
			attribute.Int64("channel_id", channelID),
		),
	)
	defer span.End()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, false, errors.Wrap(err, "begin tx")
	}

	defer func() { _ = tx.Rollback() }()

	state, err := tx.TelegramChannelState.Query().Where(
		telegramchannelstate.UserIDEQ(userID),
		telegramchannelstate.ChannelIDEQ(channelID),
	).Only(ctx)
	if ent.IsNotFound(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, errors.Wrap(err, "query state")
	}
	return state.Pts, true, nil
}

func (s *State) SetChannelPts(ctx context.Context, userID, channelID int64, pts int) error {
	ctx, span := s.tracer.Start(ctx, "SetChannelPts",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
			attribute.Int64("channel_id", channelID),
			attribute.Int("pts", pts),
		),
	)
	defer span.End()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}

	defer func() { _ = tx.Rollback() }()

	if err := tx.TelegramUserState.Create().
		SetID(userID).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).Exec(ctx); err != nil {
		return errors.Wrap(err, "upsert state")
	}

	if err := tx.TelegramChannelState.Create().
		SetChannelID(channelID).SetUserID(userID).
		SetPts(pts).
		OnConflict(
			sql.ConflictColumns("channel_id", "user_id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx); err != nil {
		return errors.Wrap(err, "upsert channel state")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit tx")
	}

	return nil
}

func (s *State) ForEachChannels(ctx context.Context, userID int64, f func(ctx context.Context, channelID int64, pts int) error) error {
	ctx, span := s.tracer.Start(ctx, "ForEachChannels",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()

	channels, err := s.db.TelegramChannelState.Query().Where(telegramchannelstate.UserIDEQ(userID)).All(ctx)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if err := f(ctx, channel.ChannelID, channel.Pts); err != nil {
			return err
		}
	}

	return nil
}
