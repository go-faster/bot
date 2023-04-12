package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type GPTDialog struct {
	ent.Schema
}

func (GPTDialog) Fields() []ent.Field {
	return []ent.Field{
		field.String("peer_id").Immutable().
			Comment("Peer ID"),
		field.Int("prompt_msg_id").
			Comment("Telegram message id of prompt message."),
		field.String("prompt_msg").
			Comment("Prompt message."),
		field.Int("gpt_msg_id").
			Comment("Telegram message id of sent message."),
		field.String("gpt_msg").
			Comment("AI-generated message. Does not include prompt."),
		field.Int("thread_top_msg_id").Optional().
			Comment("Telegram thread's top message id."),
		field.Time("created_at").Default(time.Now).Immutable().
			Comment("Message generation time. To simplify cleanup."),
	}
}

func (GPTDialog) Indexes() []ent.Index {
	return []ent.Index{
		// In order to find all thread messages.
		index.Fields("peer_id", "thread_top_msg_id"),
	}
}

func (GPTDialog) Edges() []ent.Edge {
	return []ent.Edge{}
}
