package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Repository struct {
	ent.Schema
}

func (Repository) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable().Comment("GitHub repository ID."),
		field.String("owner").Comment("GitHub repository owner."),
		field.String("name").Comment("GitHub repository name."),
		field.String("full_name").Comment("GitHub repository full name."),
		field.String("html_url").Comment("GitHub repository URL."),
		field.String("description").Default("").Comment("GitHub repository description."),
		field.Time("last_pushed_at").Optional(),
		field.Time("last_event_at").Optional(),
	}
}
