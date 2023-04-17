package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LastChannelMessage holds the last message ID of Telegram channel.
//
// We use it to compute how many messages were sent between PR event notification
// and last messsage, since Telegram does not allow to bots to query messages in a channel.
//
// The number of messages is used to find out if old event notification is out of context
// and we should send a new message for a new event instead of editing.
type LastChannelMessage struct {
	ent.Schema
}

func (LastChannelMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable().Comment("Telegram channel ID."),
		field.Int("message_id").Comment("Telegram message ID of last observed message in channel."),
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
		field.Int64("repo_id").Comment("Github repository ID."),
		field.Int("pull_request_id").Comment("Pull request number."),
		field.String("pull_request_title").Default("").Comment("Pull request title."),
		field.String("pull_request_body").Default("").Comment("Pull request body."),
		field.String("pull_request_author_login").Default("").Comment("Pull request author's login."),
		// TODO(tdakkota): store notify peer_id.
		field.Int("message_id").Comment("Telegram message ID. Belongs to notify channel."),
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
