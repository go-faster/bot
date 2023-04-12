package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type TelegramSession struct {
	ent.Schema
}

func (TelegramSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()),
		field.Bytes("data"),
	}
}

func (TelegramSession) Edges() []ent.Edge {
	return []ent.Edge{}
}
