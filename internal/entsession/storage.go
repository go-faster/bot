package entsession

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"
	"github.com/google/uuid"
	"github.com/gotd/td/session"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/telegramsession"
)

type Storage struct {
	UUID     uuid.UUID
	Database *ent.Client
	Tracer   trace.Tracer
}

func (s Storage) LoadSession(ctx context.Context) ([]byte, error) {
	ctx, span := s.Tracer.Start(ctx, "LoadSession")
	defer span.End()

	list, err := s.Database.TelegramSession.Query().
		Where(telegramsession.ID(s.UUID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		return v.Data, nil
	}
	return nil, session.ErrNotFound
}

func (s Storage) StoreSession(ctx context.Context, data []byte) error {
	ctx, span := s.Tracer.Start(ctx, "StoreSession")
	defer span.End()

	if err := s.Database.TelegramSession.Create().
		SetID(s.UUID).
		SetData(data).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateData().Exec(ctx); err != nil {
		return errors.Wrap(err, "store session")
	}
	return nil
}

var _ session.Storage = (*Storage)(nil)
