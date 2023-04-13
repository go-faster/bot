// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/go-faster/bot/internal/ent/gptdialog"
	"github.com/go-faster/bot/internal/ent/predicate"
)

// GPTDialogDelete is the builder for deleting a GPTDialog entity.
type GPTDialogDelete struct {
	config
	hooks    []Hook
	mutation *GPTDialogMutation
}

// Where appends a list predicates to the GPTDialogDelete builder.
func (gdd *GPTDialogDelete) Where(ps ...predicate.GPTDialog) *GPTDialogDelete {
	gdd.mutation.Where(ps...)
	return gdd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (gdd *GPTDialogDelete) Exec(ctx context.Context) (int, error) {
	return withHooks[int, GPTDialogMutation](ctx, gdd.sqlExec, gdd.mutation, gdd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (gdd *GPTDialogDelete) ExecX(ctx context.Context) int {
	n, err := gdd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (gdd *GPTDialogDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(gptdialog.Table, sqlgraph.NewFieldSpec(gptdialog.FieldID, field.TypeInt))
	if ps := gdd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, gdd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	gdd.mutation.done = true
	return affected, err
}

// GPTDialogDeleteOne is the builder for deleting a single GPTDialog entity.
type GPTDialogDeleteOne struct {
	gdd *GPTDialogDelete
}

// Where appends a list predicates to the GPTDialogDelete builder.
func (gddo *GPTDialogDeleteOne) Where(ps ...predicate.GPTDialog) *GPTDialogDeleteOne {
	gddo.gdd.mutation.Where(ps...)
	return gddo
}

// Exec executes the deletion query.
func (gddo *GPTDialogDeleteOne) Exec(ctx context.Context) error {
	n, err := gddo.gdd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{gptdialog.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (gddo *GPTDialogDeleteOne) ExecX(ctx context.Context) {
	if err := gddo.Exec(ctx); err != nil {
		panic(err)
	}
}