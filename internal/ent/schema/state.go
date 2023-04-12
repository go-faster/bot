package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LastChannelMessage struct {
	ent.Schema
}

func (LastChannelMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable().Comment("Channel ID"),
		field.Int("message_id"),
	}
}

func (LastChannelMessage) Edges() []ent.Edge {
	return []ent.Edge{}
}

type PRNotification struct {
	ent.Schema
}

func (PRNotification) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("repo_id"),
		field.Int("pull_request_id"),
		field.Int("message_id"),
	}
}

func (PRNotification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("repo_id", "pull_request_id").
			Unique(),
	}
}

func (PRNotification) Edges() []ent.Edge {
	return []ent.Edge{}
}
