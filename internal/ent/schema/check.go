package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Check for workflow.
//
// https://docs.github.com/webhooks-and-events/webhooks/webhook-events-and-payloads#check_run
type Check struct {
	ent.Schema
}

func (Check) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Comment("Value of check_run.id"),
		field.Int("repo_id").Comment("Repository id"),
		field.String("name").Comment("Name of check_run"),
		field.String("status").Comment(`The phase of the lifecycle that the check is currently in. Can be one of: queued, in_progress, completed, pending`),
		field.String("conclusion").Optional().Comment(`The final conclusion of the check. Can be one of: waiting, pending, startup_failure, stale, success, failure, neutral, cancelled, skipped, timed_out, action_required, null`),
	}
}

func (Check) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("repo_id", "id").
			Unique(),
	}
}
