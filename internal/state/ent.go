package state

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/go-faster/bot/internal/ent"
)

type Ent struct {
	db     *ent.Client
	tracer trace.Tracer
}

func NewEnt(db *ent.Client, t trace.TracerProvider) *Ent {
	return &Ent{db: db, tracer: t.Tracer("state")}
}

var _ Storage = (*Ent)(nil)
