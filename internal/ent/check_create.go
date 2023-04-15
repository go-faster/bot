// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/go-faster/bot/internal/ent/check"
)

// CheckCreate is the builder for creating a Check entity.
type CheckCreate struct {
	config
	mutation *CheckMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetRepoID sets the "repo_id" field.
func (cc *CheckCreate) SetRepoID(i int64) *CheckCreate {
	cc.mutation.SetRepoID(i)
	return cc
}

// SetPullRequestID sets the "pull_request_id" field.
func (cc *CheckCreate) SetPullRequestID(i int) *CheckCreate {
	cc.mutation.SetPullRequestID(i)
	return cc
}

// SetName sets the "name" field.
func (cc *CheckCreate) SetName(s string) *CheckCreate {
	cc.mutation.SetName(s)
	return cc
}

// SetStatus sets the "status" field.
func (cc *CheckCreate) SetStatus(s string) *CheckCreate {
	cc.mutation.SetStatus(s)
	return cc
}

// SetConclusion sets the "conclusion" field.
func (cc *CheckCreate) SetConclusion(s string) *CheckCreate {
	cc.mutation.SetConclusion(s)
	return cc
}

// SetNillableConclusion sets the "conclusion" field if the given value is not nil.
func (cc *CheckCreate) SetNillableConclusion(s *string) *CheckCreate {
	if s != nil {
		cc.SetConclusion(*s)
	}
	return cc
}

// SetID sets the "id" field.
func (cc *CheckCreate) SetID(i int64) *CheckCreate {
	cc.mutation.SetID(i)
	return cc
}

// Mutation returns the CheckMutation object of the builder.
func (cc *CheckCreate) Mutation() *CheckMutation {
	return cc.mutation
}

// Save creates the Check in the database.
func (cc *CheckCreate) Save(ctx context.Context) (*Check, error) {
	return withHooks[*Check, CheckMutation](ctx, cc.sqlSave, cc.mutation, cc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (cc *CheckCreate) SaveX(ctx context.Context) *Check {
	v, err := cc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (cc *CheckCreate) Exec(ctx context.Context) error {
	_, err := cc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (cc *CheckCreate) ExecX(ctx context.Context) {
	if err := cc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (cc *CheckCreate) check() error {
	if _, ok := cc.mutation.RepoID(); !ok {
		return &ValidationError{Name: "repo_id", err: errors.New(`ent: missing required field "Check.repo_id"`)}
	}
	if _, ok := cc.mutation.PullRequestID(); !ok {
		return &ValidationError{Name: "pull_request_id", err: errors.New(`ent: missing required field "Check.pull_request_id"`)}
	}
	if _, ok := cc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Check.name"`)}
	}
	if _, ok := cc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "Check.status"`)}
	}
	return nil
}

func (cc *CheckCreate) sqlSave(ctx context.Context) (*Check, error) {
	if err := cc.check(); err != nil {
		return nil, err
	}
	_node, _spec := cc.createSpec()
	if err := sqlgraph.CreateNode(ctx, cc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != _node.ID {
		id := _spec.ID.Value.(int64)
		_node.ID = int64(id)
	}
	cc.mutation.id = &_node.ID
	cc.mutation.done = true
	return _node, nil
}

func (cc *CheckCreate) createSpec() (*Check, *sqlgraph.CreateSpec) {
	var (
		_node = &Check{config: cc.config}
		_spec = sqlgraph.NewCreateSpec(check.Table, sqlgraph.NewFieldSpec(check.FieldID, field.TypeInt64))
	)
	_spec.OnConflict = cc.conflict
	if id, ok := cc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := cc.mutation.RepoID(); ok {
		_spec.SetField(check.FieldRepoID, field.TypeInt64, value)
		_node.RepoID = value
	}
	if value, ok := cc.mutation.PullRequestID(); ok {
		_spec.SetField(check.FieldPullRequestID, field.TypeInt, value)
		_node.PullRequestID = value
	}
	if value, ok := cc.mutation.Name(); ok {
		_spec.SetField(check.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := cc.mutation.Status(); ok {
		_spec.SetField(check.FieldStatus, field.TypeString, value)
		_node.Status = value
	}
	if value, ok := cc.mutation.Conclusion(); ok {
		_spec.SetField(check.FieldConclusion, field.TypeString, value)
		_node.Conclusion = value
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Check.Create().
//		SetRepoID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.CheckUpsert) {
//			SetRepoID(v+v).
//		}).
//		Exec(ctx)
func (cc *CheckCreate) OnConflict(opts ...sql.ConflictOption) *CheckUpsertOne {
	cc.conflict = opts
	return &CheckUpsertOne{
		create: cc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Check.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (cc *CheckCreate) OnConflictColumns(columns ...string) *CheckUpsertOne {
	cc.conflict = append(cc.conflict, sql.ConflictColumns(columns...))
	return &CheckUpsertOne{
		create: cc,
	}
}

type (
	// CheckUpsertOne is the builder for "upsert"-ing
	//  one Check node.
	CheckUpsertOne struct {
		create *CheckCreate
	}

	// CheckUpsert is the "OnConflict" setter.
	CheckUpsert struct {
		*sql.UpdateSet
	}
)

// SetRepoID sets the "repo_id" field.
func (u *CheckUpsert) SetRepoID(v int64) *CheckUpsert {
	u.Set(check.FieldRepoID, v)
	return u
}

// UpdateRepoID sets the "repo_id" field to the value that was provided on create.
func (u *CheckUpsert) UpdateRepoID() *CheckUpsert {
	u.SetExcluded(check.FieldRepoID)
	return u
}

// AddRepoID adds v to the "repo_id" field.
func (u *CheckUpsert) AddRepoID(v int64) *CheckUpsert {
	u.Add(check.FieldRepoID, v)
	return u
}

// SetPullRequestID sets the "pull_request_id" field.
func (u *CheckUpsert) SetPullRequestID(v int) *CheckUpsert {
	u.Set(check.FieldPullRequestID, v)
	return u
}

// UpdatePullRequestID sets the "pull_request_id" field to the value that was provided on create.
func (u *CheckUpsert) UpdatePullRequestID() *CheckUpsert {
	u.SetExcluded(check.FieldPullRequestID)
	return u
}

// AddPullRequestID adds v to the "pull_request_id" field.
func (u *CheckUpsert) AddPullRequestID(v int) *CheckUpsert {
	u.Add(check.FieldPullRequestID, v)
	return u
}

// SetName sets the "name" field.
func (u *CheckUpsert) SetName(v string) *CheckUpsert {
	u.Set(check.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *CheckUpsert) UpdateName() *CheckUpsert {
	u.SetExcluded(check.FieldName)
	return u
}

// SetStatus sets the "status" field.
func (u *CheckUpsert) SetStatus(v string) *CheckUpsert {
	u.Set(check.FieldStatus, v)
	return u
}

// UpdateStatus sets the "status" field to the value that was provided on create.
func (u *CheckUpsert) UpdateStatus() *CheckUpsert {
	u.SetExcluded(check.FieldStatus)
	return u
}

// SetConclusion sets the "conclusion" field.
func (u *CheckUpsert) SetConclusion(v string) *CheckUpsert {
	u.Set(check.FieldConclusion, v)
	return u
}

// UpdateConclusion sets the "conclusion" field to the value that was provided on create.
func (u *CheckUpsert) UpdateConclusion() *CheckUpsert {
	u.SetExcluded(check.FieldConclusion)
	return u
}

// ClearConclusion clears the value of the "conclusion" field.
func (u *CheckUpsert) ClearConclusion() *CheckUpsert {
	u.SetNull(check.FieldConclusion)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Check.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(check.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *CheckUpsertOne) UpdateNewValues() *CheckUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(check.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Check.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *CheckUpsertOne) Ignore() *CheckUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *CheckUpsertOne) DoNothing() *CheckUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the CheckCreate.OnConflict
// documentation for more info.
func (u *CheckUpsertOne) Update(set func(*CheckUpsert)) *CheckUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&CheckUpsert{UpdateSet: update})
	}))
	return u
}

// SetRepoID sets the "repo_id" field.
func (u *CheckUpsertOne) SetRepoID(v int64) *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.SetRepoID(v)
	})
}

// AddRepoID adds v to the "repo_id" field.
func (u *CheckUpsertOne) AddRepoID(v int64) *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.AddRepoID(v)
	})
}

// UpdateRepoID sets the "repo_id" field to the value that was provided on create.
func (u *CheckUpsertOne) UpdateRepoID() *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.UpdateRepoID()
	})
}

// SetPullRequestID sets the "pull_request_id" field.
func (u *CheckUpsertOne) SetPullRequestID(v int) *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.SetPullRequestID(v)
	})
}

// AddPullRequestID adds v to the "pull_request_id" field.
func (u *CheckUpsertOne) AddPullRequestID(v int) *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.AddPullRequestID(v)
	})
}

// UpdatePullRequestID sets the "pull_request_id" field to the value that was provided on create.
func (u *CheckUpsertOne) UpdatePullRequestID() *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.UpdatePullRequestID()
	})
}

// SetName sets the "name" field.
func (u *CheckUpsertOne) SetName(v string) *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *CheckUpsertOne) UpdateName() *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.UpdateName()
	})
}

// SetStatus sets the "status" field.
func (u *CheckUpsertOne) SetStatus(v string) *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.SetStatus(v)
	})
}

// UpdateStatus sets the "status" field to the value that was provided on create.
func (u *CheckUpsertOne) UpdateStatus() *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.UpdateStatus()
	})
}

// SetConclusion sets the "conclusion" field.
func (u *CheckUpsertOne) SetConclusion(v string) *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.SetConclusion(v)
	})
}

// UpdateConclusion sets the "conclusion" field to the value that was provided on create.
func (u *CheckUpsertOne) UpdateConclusion() *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.UpdateConclusion()
	})
}

// ClearConclusion clears the value of the "conclusion" field.
func (u *CheckUpsertOne) ClearConclusion() *CheckUpsertOne {
	return u.Update(func(s *CheckUpsert) {
		s.ClearConclusion()
	})
}

// Exec executes the query.
func (u *CheckUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for CheckCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *CheckUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *CheckUpsertOne) ID(ctx context.Context) (id int64, err error) {
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *CheckUpsertOne) IDX(ctx context.Context) int64 {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// CheckCreateBulk is the builder for creating many Check entities in bulk.
type CheckCreateBulk struct {
	config
	builders []*CheckCreate
	conflict []sql.ConflictOption
}

// Save creates the Check entities in the database.
func (ccb *CheckCreateBulk) Save(ctx context.Context) ([]*Check, error) {
	specs := make([]*sqlgraph.CreateSpec, len(ccb.builders))
	nodes := make([]*Check, len(ccb.builders))
	mutators := make([]Mutator, len(ccb.builders))
	for i := range ccb.builders {
		func(i int, root context.Context) {
			builder := ccb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*CheckMutation)
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
					_, err = mutators[i+1].Mutate(root, ccb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ccb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ccb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ccb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ccb *CheckCreateBulk) SaveX(ctx context.Context) []*Check {
	v, err := ccb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ccb *CheckCreateBulk) Exec(ctx context.Context) error {
	_, err := ccb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ccb *CheckCreateBulk) ExecX(ctx context.Context) {
	if err := ccb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Check.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.CheckUpsert) {
//			SetRepoID(v+v).
//		}).
//		Exec(ctx)
func (ccb *CheckCreateBulk) OnConflict(opts ...sql.ConflictOption) *CheckUpsertBulk {
	ccb.conflict = opts
	return &CheckUpsertBulk{
		create: ccb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Check.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ccb *CheckCreateBulk) OnConflictColumns(columns ...string) *CheckUpsertBulk {
	ccb.conflict = append(ccb.conflict, sql.ConflictColumns(columns...))
	return &CheckUpsertBulk{
		create: ccb,
	}
}

// CheckUpsertBulk is the builder for "upsert"-ing
// a bulk of Check nodes.
type CheckUpsertBulk struct {
	create *CheckCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Check.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(check.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *CheckUpsertBulk) UpdateNewValues() *CheckUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(check.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Check.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *CheckUpsertBulk) Ignore() *CheckUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *CheckUpsertBulk) DoNothing() *CheckUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the CheckCreateBulk.OnConflict
// documentation for more info.
func (u *CheckUpsertBulk) Update(set func(*CheckUpsert)) *CheckUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&CheckUpsert{UpdateSet: update})
	}))
	return u
}

// SetRepoID sets the "repo_id" field.
func (u *CheckUpsertBulk) SetRepoID(v int64) *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.SetRepoID(v)
	})
}

// AddRepoID adds v to the "repo_id" field.
func (u *CheckUpsertBulk) AddRepoID(v int64) *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.AddRepoID(v)
	})
}

// UpdateRepoID sets the "repo_id" field to the value that was provided on create.
func (u *CheckUpsertBulk) UpdateRepoID() *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.UpdateRepoID()
	})
}

// SetPullRequestID sets the "pull_request_id" field.
func (u *CheckUpsertBulk) SetPullRequestID(v int) *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.SetPullRequestID(v)
	})
}

// AddPullRequestID adds v to the "pull_request_id" field.
func (u *CheckUpsertBulk) AddPullRequestID(v int) *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.AddPullRequestID(v)
	})
}

// UpdatePullRequestID sets the "pull_request_id" field to the value that was provided on create.
func (u *CheckUpsertBulk) UpdatePullRequestID() *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.UpdatePullRequestID()
	})
}

// SetName sets the "name" field.
func (u *CheckUpsertBulk) SetName(v string) *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *CheckUpsertBulk) UpdateName() *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.UpdateName()
	})
}

// SetStatus sets the "status" field.
func (u *CheckUpsertBulk) SetStatus(v string) *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.SetStatus(v)
	})
}

// UpdateStatus sets the "status" field to the value that was provided on create.
func (u *CheckUpsertBulk) UpdateStatus() *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.UpdateStatus()
	})
}

// SetConclusion sets the "conclusion" field.
func (u *CheckUpsertBulk) SetConclusion(v string) *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.SetConclusion(v)
	})
}

// UpdateConclusion sets the "conclusion" field to the value that was provided on create.
func (u *CheckUpsertBulk) UpdateConclusion() *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.UpdateConclusion()
	})
}

// ClearConclusion clears the value of the "conclusion" field.
func (u *CheckUpsertBulk) ClearConclusion() *CheckUpsertBulk {
	return u.Update(func(s *CheckUpsert) {
		s.ClearConclusion()
	})
}

// Exec executes the query.
func (u *CheckUpsertBulk) Exec(ctx context.Context) error {
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the CheckCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for CheckCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *CheckUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}