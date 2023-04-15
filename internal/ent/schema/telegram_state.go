package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TelegramUserState struct {
	ent.Schema
}

func (TelegramUserState) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Comment("User ID"),
		field.Int("qts").Default(0),
		field.Int("pts").Default(0),
		field.Int("date").Default(0),
		field.Int("seq").Default(0),
	}
}

func (TelegramUserState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("channels", TelegramChannelState.Type),
	}
}

type TelegramChannelState struct {
	ent.Schema
}

func (TelegramChannelState) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("channel_id").Comment("Channel id"),
		field.Int64("user_id").Comment("User id"),
		field.Int("pts").Default(0),
	}
}

func (TelegramChannelState) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "channel_id").Unique(),
	}
}

func (TelegramChannelState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", TelegramUserState.Type).
			Ref("channels").
			Field("user_id").
			Unique().
			Required(),
	}
}
