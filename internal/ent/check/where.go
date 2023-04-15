// Code generated by ent, DO NOT EDIT.

package check

import (
	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/bot/internal/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int64) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int64) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int64) predicate.Check {
	return predicate.Check(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int64) predicate.Check {
	return predicate.Check(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int64) predicate.Check {
	return predicate.Check(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int64) predicate.Check {
	return predicate.Check(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int64) predicate.Check {
	return predicate.Check(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int64) predicate.Check {
	return predicate.Check(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int64) predicate.Check {
	return predicate.Check(sql.FieldLTE(FieldID, id))
}

// RepoID applies equality check predicate on the "repo_id" field. It's identical to RepoIDEQ.
func RepoID(v int64) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldRepoID, v))
}

// PullRequestID applies equality check predicate on the "pull_request_id" field. It's identical to PullRequestIDEQ.
func PullRequestID(v int) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldPullRequestID, v))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldName, v))
}

// Status applies equality check predicate on the "status" field. It's identical to StatusEQ.
func Status(v string) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldStatus, v))
}

// Conclusion applies equality check predicate on the "conclusion" field. It's identical to ConclusionEQ.
func Conclusion(v string) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldConclusion, v))
}

// RepoIDEQ applies the EQ predicate on the "repo_id" field.
func RepoIDEQ(v int64) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldRepoID, v))
}

// RepoIDNEQ applies the NEQ predicate on the "repo_id" field.
func RepoIDNEQ(v int64) predicate.Check {
	return predicate.Check(sql.FieldNEQ(FieldRepoID, v))
}

// RepoIDIn applies the In predicate on the "repo_id" field.
func RepoIDIn(vs ...int64) predicate.Check {
	return predicate.Check(sql.FieldIn(FieldRepoID, vs...))
}

// RepoIDNotIn applies the NotIn predicate on the "repo_id" field.
func RepoIDNotIn(vs ...int64) predicate.Check {
	return predicate.Check(sql.FieldNotIn(FieldRepoID, vs...))
}

// RepoIDGT applies the GT predicate on the "repo_id" field.
func RepoIDGT(v int64) predicate.Check {
	return predicate.Check(sql.FieldGT(FieldRepoID, v))
}

// RepoIDGTE applies the GTE predicate on the "repo_id" field.
func RepoIDGTE(v int64) predicate.Check {
	return predicate.Check(sql.FieldGTE(FieldRepoID, v))
}

// RepoIDLT applies the LT predicate on the "repo_id" field.
func RepoIDLT(v int64) predicate.Check {
	return predicate.Check(sql.FieldLT(FieldRepoID, v))
}

// RepoIDLTE applies the LTE predicate on the "repo_id" field.
func RepoIDLTE(v int64) predicate.Check {
	return predicate.Check(sql.FieldLTE(FieldRepoID, v))
}

// PullRequestIDEQ applies the EQ predicate on the "pull_request_id" field.
func PullRequestIDEQ(v int) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldPullRequestID, v))
}

// PullRequestIDNEQ applies the NEQ predicate on the "pull_request_id" field.
func PullRequestIDNEQ(v int) predicate.Check {
	return predicate.Check(sql.FieldNEQ(FieldPullRequestID, v))
}

// PullRequestIDIn applies the In predicate on the "pull_request_id" field.
func PullRequestIDIn(vs ...int) predicate.Check {
	return predicate.Check(sql.FieldIn(FieldPullRequestID, vs...))
}

// PullRequestIDNotIn applies the NotIn predicate on the "pull_request_id" field.
func PullRequestIDNotIn(vs ...int) predicate.Check {
	return predicate.Check(sql.FieldNotIn(FieldPullRequestID, vs...))
}

// PullRequestIDGT applies the GT predicate on the "pull_request_id" field.
func PullRequestIDGT(v int) predicate.Check {
	return predicate.Check(sql.FieldGT(FieldPullRequestID, v))
}

// PullRequestIDGTE applies the GTE predicate on the "pull_request_id" field.
func PullRequestIDGTE(v int) predicate.Check {
	return predicate.Check(sql.FieldGTE(FieldPullRequestID, v))
}

// PullRequestIDLT applies the LT predicate on the "pull_request_id" field.
func PullRequestIDLT(v int) predicate.Check {
	return predicate.Check(sql.FieldLT(FieldPullRequestID, v))
}

// PullRequestIDLTE applies the LTE predicate on the "pull_request_id" field.
func PullRequestIDLTE(v int) predicate.Check {
	return predicate.Check(sql.FieldLTE(FieldPullRequestID, v))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.Check {
	return predicate.Check(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.Check {
	return predicate.Check(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.Check {
	return predicate.Check(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.Check {
	return predicate.Check(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.Check {
	return predicate.Check(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.Check {
	return predicate.Check(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.Check {
	return predicate.Check(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.Check {
	return predicate.Check(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.Check {
	return predicate.Check(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.Check {
	return predicate.Check(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.Check {
	return predicate.Check(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.Check {
	return predicate.Check(sql.FieldContainsFold(FieldName, v))
}

// StatusEQ applies the EQ predicate on the "status" field.
func StatusEQ(v string) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldStatus, v))
}

// StatusNEQ applies the NEQ predicate on the "status" field.
func StatusNEQ(v string) predicate.Check {
	return predicate.Check(sql.FieldNEQ(FieldStatus, v))
}

// StatusIn applies the In predicate on the "status" field.
func StatusIn(vs ...string) predicate.Check {
	return predicate.Check(sql.FieldIn(FieldStatus, vs...))
}

// StatusNotIn applies the NotIn predicate on the "status" field.
func StatusNotIn(vs ...string) predicate.Check {
	return predicate.Check(sql.FieldNotIn(FieldStatus, vs...))
}

// StatusGT applies the GT predicate on the "status" field.
func StatusGT(v string) predicate.Check {
	return predicate.Check(sql.FieldGT(FieldStatus, v))
}

// StatusGTE applies the GTE predicate on the "status" field.
func StatusGTE(v string) predicate.Check {
	return predicate.Check(sql.FieldGTE(FieldStatus, v))
}

// StatusLT applies the LT predicate on the "status" field.
func StatusLT(v string) predicate.Check {
	return predicate.Check(sql.FieldLT(FieldStatus, v))
}

// StatusLTE applies the LTE predicate on the "status" field.
func StatusLTE(v string) predicate.Check {
	return predicate.Check(sql.FieldLTE(FieldStatus, v))
}

// StatusContains applies the Contains predicate on the "status" field.
func StatusContains(v string) predicate.Check {
	return predicate.Check(sql.FieldContains(FieldStatus, v))
}

// StatusHasPrefix applies the HasPrefix predicate on the "status" field.
func StatusHasPrefix(v string) predicate.Check {
	return predicate.Check(sql.FieldHasPrefix(FieldStatus, v))
}

// StatusHasSuffix applies the HasSuffix predicate on the "status" field.
func StatusHasSuffix(v string) predicate.Check {
	return predicate.Check(sql.FieldHasSuffix(FieldStatus, v))
}

// StatusEqualFold applies the EqualFold predicate on the "status" field.
func StatusEqualFold(v string) predicate.Check {
	return predicate.Check(sql.FieldEqualFold(FieldStatus, v))
}

// StatusContainsFold applies the ContainsFold predicate on the "status" field.
func StatusContainsFold(v string) predicate.Check {
	return predicate.Check(sql.FieldContainsFold(FieldStatus, v))
}

// ConclusionEQ applies the EQ predicate on the "conclusion" field.
func ConclusionEQ(v string) predicate.Check {
	return predicate.Check(sql.FieldEQ(FieldConclusion, v))
}

// ConclusionNEQ applies the NEQ predicate on the "conclusion" field.
func ConclusionNEQ(v string) predicate.Check {
	return predicate.Check(sql.FieldNEQ(FieldConclusion, v))
}

// ConclusionIn applies the In predicate on the "conclusion" field.
func ConclusionIn(vs ...string) predicate.Check {
	return predicate.Check(sql.FieldIn(FieldConclusion, vs...))
}

// ConclusionNotIn applies the NotIn predicate on the "conclusion" field.
func ConclusionNotIn(vs ...string) predicate.Check {
	return predicate.Check(sql.FieldNotIn(FieldConclusion, vs...))
}

// ConclusionGT applies the GT predicate on the "conclusion" field.
func ConclusionGT(v string) predicate.Check {
	return predicate.Check(sql.FieldGT(FieldConclusion, v))
}

// ConclusionGTE applies the GTE predicate on the "conclusion" field.
func ConclusionGTE(v string) predicate.Check {
	return predicate.Check(sql.FieldGTE(FieldConclusion, v))
}

// ConclusionLT applies the LT predicate on the "conclusion" field.
func ConclusionLT(v string) predicate.Check {
	return predicate.Check(sql.FieldLT(FieldConclusion, v))
}

// ConclusionLTE applies the LTE predicate on the "conclusion" field.
func ConclusionLTE(v string) predicate.Check {
	return predicate.Check(sql.FieldLTE(FieldConclusion, v))
}

// ConclusionContains applies the Contains predicate on the "conclusion" field.
func ConclusionContains(v string) predicate.Check {
	return predicate.Check(sql.FieldContains(FieldConclusion, v))
}

// ConclusionHasPrefix applies the HasPrefix predicate on the "conclusion" field.
func ConclusionHasPrefix(v string) predicate.Check {
	return predicate.Check(sql.FieldHasPrefix(FieldConclusion, v))
}

// ConclusionHasSuffix applies the HasSuffix predicate on the "conclusion" field.
func ConclusionHasSuffix(v string) predicate.Check {
	return predicate.Check(sql.FieldHasSuffix(FieldConclusion, v))
}

// ConclusionIsNil applies the IsNil predicate on the "conclusion" field.
func ConclusionIsNil() predicate.Check {
	return predicate.Check(sql.FieldIsNull(FieldConclusion))
}

// ConclusionNotNil applies the NotNil predicate on the "conclusion" field.
func ConclusionNotNil() predicate.Check {
	return predicate.Check(sql.FieldNotNull(FieldConclusion))
}

// ConclusionEqualFold applies the EqualFold predicate on the "conclusion" field.
func ConclusionEqualFold(v string) predicate.Check {
	return predicate.Check(sql.FieldEqualFold(FieldConclusion, v))
}

// ConclusionContainsFold applies the ContainsFold predicate on the "conclusion" field.
func ConclusionContainsFold(v string) predicate.Check {
	return predicate.Check(sql.FieldContainsFold(FieldConclusion, v))
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Check) predicate.Check {
	return predicate.Check(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Check) predicate.Check {
	return predicate.Check(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Check) predicate.Check {
	return predicate.Check(func(s *sql.Selector) {
		p(s.Not())
	})
}