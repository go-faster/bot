// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/go-faster/bot/internal/ent/organization"
	"github.com/go-faster/bot/internal/ent/repository"
)

// OrganizationCreate is the builder for creating a Organization entity.
type OrganizationCreate struct {
	config
	mutation *OrganizationMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetName sets the "name" field.
func (oc *OrganizationCreate) SetName(s string) *OrganizationCreate {
	oc.mutation.SetName(s)
	return oc
}

// SetHTMLURL sets the "html_url" field.
func (oc *OrganizationCreate) SetHTMLURL(s string) *OrganizationCreate {
	oc.mutation.SetHTMLURL(s)
	return oc
}

// SetNillableHTMLURL sets the "html_url" field if the given value is not nil.
func (oc *OrganizationCreate) SetNillableHTMLURL(s *string) *OrganizationCreate {
	if s != nil {
		oc.SetHTMLURL(*s)
	}
	return oc
}

// SetID sets the "id" field.
func (oc *OrganizationCreate) SetID(i int64) *OrganizationCreate {
	oc.mutation.SetID(i)
	return oc
}

// AddRepositoryIDs adds the "repositories" edge to the Repository entity by IDs.
func (oc *OrganizationCreate) AddRepositoryIDs(ids ...int64) *OrganizationCreate {
	oc.mutation.AddRepositoryIDs(ids...)
	return oc
}

// AddRepositories adds the "repositories" edges to the Repository entity.
func (oc *OrganizationCreate) AddRepositories(r ...*Repository) *OrganizationCreate {
	ids := make([]int64, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return oc.AddRepositoryIDs(ids...)
}

// Mutation returns the OrganizationMutation object of the builder.
func (oc *OrganizationCreate) Mutation() *OrganizationMutation {
	return oc.mutation
}

// Save creates the Organization in the database.
func (oc *OrganizationCreate) Save(ctx context.Context) (*Organization, error) {
	return withHooks[*Organization, OrganizationMutation](ctx, oc.sqlSave, oc.mutation, oc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (oc *OrganizationCreate) SaveX(ctx context.Context) *Organization {
	v, err := oc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (oc *OrganizationCreate) Exec(ctx context.Context) error {
	_, err := oc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (oc *OrganizationCreate) ExecX(ctx context.Context) {
	if err := oc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (oc *OrganizationCreate) check() error {
	if _, ok := oc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Organization.name"`)}
	}
	return nil
}

func (oc *OrganizationCreate) sqlSave(ctx context.Context) (*Organization, error) {
	if err := oc.check(); err != nil {
		return nil, err
	}
	_node, _spec := oc.createSpec()
	if err := sqlgraph.CreateNode(ctx, oc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != _node.ID {
		id := _spec.ID.Value.(int64)
		_node.ID = int64(id)
	}
	oc.mutation.id = &_node.ID
	oc.mutation.done = true
	return _node, nil
}

func (oc *OrganizationCreate) createSpec() (*Organization, *sqlgraph.CreateSpec) {
	var (
		_node = &Organization{config: oc.config}
		_spec = sqlgraph.NewCreateSpec(organization.Table, sqlgraph.NewFieldSpec(organization.FieldID, field.TypeInt64))
	)
	_spec.OnConflict = oc.conflict
	if id, ok := oc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := oc.mutation.Name(); ok {
		_spec.SetField(organization.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := oc.mutation.HTMLURL(); ok {
		_spec.SetField(organization.FieldHTMLURL, field.TypeString, value)
		_node.HTMLURL = value
	}
	if nodes := oc.mutation.RepositoriesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   organization.RepositoriesTable,
			Columns: []string{organization.RepositoriesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(repository.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Organization.Create().
//		SetName(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.OrganizationUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (oc *OrganizationCreate) OnConflict(opts ...sql.ConflictOption) *OrganizationUpsertOne {
	oc.conflict = opts
	return &OrganizationUpsertOne{
		create: oc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Organization.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (oc *OrganizationCreate) OnConflictColumns(columns ...string) *OrganizationUpsertOne {
	oc.conflict = append(oc.conflict, sql.ConflictColumns(columns...))
	return &OrganizationUpsertOne{
		create: oc,
	}
}

type (
	// OrganizationUpsertOne is the builder for "upsert"-ing
	//  one Organization node.
	OrganizationUpsertOne struct {
		create *OrganizationCreate
	}

	// OrganizationUpsert is the "OnConflict" setter.
	OrganizationUpsert struct {
		*sql.UpdateSet
	}
)

// SetName sets the "name" field.
func (u *OrganizationUpsert) SetName(v string) *OrganizationUpsert {
	u.Set(organization.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *OrganizationUpsert) UpdateName() *OrganizationUpsert {
	u.SetExcluded(organization.FieldName)
	return u
}

// SetHTMLURL sets the "html_url" field.
func (u *OrganizationUpsert) SetHTMLURL(v string) *OrganizationUpsert {
	u.Set(organization.FieldHTMLURL, v)
	return u
}

// UpdateHTMLURL sets the "html_url" field to the value that was provided on create.
func (u *OrganizationUpsert) UpdateHTMLURL() *OrganizationUpsert {
	u.SetExcluded(organization.FieldHTMLURL)
	return u
}

// ClearHTMLURL clears the value of the "html_url" field.
func (u *OrganizationUpsert) ClearHTMLURL() *OrganizationUpsert {
	u.SetNull(organization.FieldHTMLURL)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Organization.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(organization.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *OrganizationUpsertOne) UpdateNewValues() *OrganizationUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(organization.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Organization.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *OrganizationUpsertOne) Ignore() *OrganizationUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *OrganizationUpsertOne) DoNothing() *OrganizationUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the OrganizationCreate.OnConflict
// documentation for more info.
func (u *OrganizationUpsertOne) Update(set func(*OrganizationUpsert)) *OrganizationUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&OrganizationUpsert{UpdateSet: update})
	}))
	return u
}

// SetName sets the "name" field.
func (u *OrganizationUpsertOne) SetName(v string) *OrganizationUpsertOne {
	return u.Update(func(s *OrganizationUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *OrganizationUpsertOne) UpdateName() *OrganizationUpsertOne {
	return u.Update(func(s *OrganizationUpsert) {
		s.UpdateName()
	})
}

// SetHTMLURL sets the "html_url" field.
func (u *OrganizationUpsertOne) SetHTMLURL(v string) *OrganizationUpsertOne {
	return u.Update(func(s *OrganizationUpsert) {
		s.SetHTMLURL(v)
	})
}

// UpdateHTMLURL sets the "html_url" field to the value that was provided on create.
func (u *OrganizationUpsertOne) UpdateHTMLURL() *OrganizationUpsertOne {
	return u.Update(func(s *OrganizationUpsert) {
		s.UpdateHTMLURL()
	})
}

// ClearHTMLURL clears the value of the "html_url" field.
func (u *OrganizationUpsertOne) ClearHTMLURL() *OrganizationUpsertOne {
	return u.Update(func(s *OrganizationUpsert) {
		s.ClearHTMLURL()
	})
}

// Exec executes the query.
func (u *OrganizationUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for OrganizationCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *OrganizationUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *OrganizationUpsertOne) ID(ctx context.Context) (id int64, err error) {
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *OrganizationUpsertOne) IDX(ctx context.Context) int64 {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// OrganizationCreateBulk is the builder for creating many Organization entities in bulk.
type OrganizationCreateBulk struct {
	config
	builders []*OrganizationCreate
	conflict []sql.ConflictOption
}

// Save creates the Organization entities in the database.
func (ocb *OrganizationCreateBulk) Save(ctx context.Context) ([]*Organization, error) {
	specs := make([]*sqlgraph.CreateSpec, len(ocb.builders))
	nodes := make([]*Organization, len(ocb.builders))
	mutators := make([]Mutator, len(ocb.builders))
	for i := range ocb.builders {
		func(i int, root context.Context) {
			builder := ocb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*OrganizationMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, ocb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ocb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ocb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil && nodes[i].ID == 0 {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int64(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, ocb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ocb *OrganizationCreateBulk) SaveX(ctx context.Context) []*Organization {
	v, err := ocb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ocb *OrganizationCreateBulk) Exec(ctx context.Context) error {
	_, err := ocb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ocb *OrganizationCreateBulk) ExecX(ctx context.Context) {
	if err := ocb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Organization.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.OrganizationUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (ocb *OrganizationCreateBulk) OnConflict(opts ...sql.ConflictOption) *OrganizationUpsertBulk {
	ocb.conflict = opts
	return &OrganizationUpsertBulk{
		create: ocb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Organization.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ocb *OrganizationCreateBulk) OnConflictColumns(columns ...string) *OrganizationUpsertBulk {
	ocb.conflict = append(ocb.conflict, sql.ConflictColumns(columns...))
	return &OrganizationUpsertBulk{
		create: ocb,
	}
}

// OrganizationUpsertBulk is the builder for "upsert"-ing
// a bulk of Organization nodes.
type OrganizationUpsertBulk struct {
	create *OrganizationCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Organization.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(organization.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *OrganizationUpsertBulk) UpdateNewValues() *OrganizationUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(organization.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Organization.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *OrganizationUpsertBulk) Ignore() *OrganizationUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *OrganizationUpsertBulk) DoNothing() *OrganizationUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the OrganizationCreateBulk.OnConflict
// documentation for more info.
func (u *OrganizationUpsertBulk) Update(set func(*OrganizationUpsert)) *OrganizationUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&OrganizationUpsert{UpdateSet: update})
	}))
	return u
}

// SetName sets the "name" field.
func (u *OrganizationUpsertBulk) SetName(v string) *OrganizationUpsertBulk {
	return u.Update(func(s *OrganizationUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *OrganizationUpsertBulk) UpdateName() *OrganizationUpsertBulk {
	return u.Update(func(s *OrganizationUpsert) {
		s.UpdateName()
	})
}

// SetHTMLURL sets the "html_url" field.
func (u *OrganizationUpsertBulk) SetHTMLURL(v string) *OrganizationUpsertBulk {
	return u.Update(func(s *OrganizationUpsert) {
		s.SetHTMLURL(v)
	})
}

// UpdateHTMLURL sets the "html_url" field to the value that was provided on create.
func (u *OrganizationUpsertBulk) UpdateHTMLURL() *OrganizationUpsertBulk {
	return u.Update(func(s *OrganizationUpsert) {
		s.UpdateHTMLURL()
	})
}

// ClearHTMLURL clears the value of the "html_url" field.
func (u *OrganizationUpsertBulk) ClearHTMLURL() *OrganizationUpsertBulk {
	return u.Update(func(s *OrganizationUpsert) {
		s.ClearHTMLURL()
	})
}

// Exec executes the query.
func (u *OrganizationUpsertBulk) Exec(ctx context.Context) error {
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the OrganizationCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for OrganizationCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *OrganizationUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}