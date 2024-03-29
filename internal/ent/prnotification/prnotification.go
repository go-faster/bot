// Code generated by ent, DO NOT EDIT.

package prnotification

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the prnotification type in the database.
	Label = "pr_notification"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldRepoID holds the string denoting the repo_id field in the database.
	FieldRepoID = "repo_id"
	// FieldPullRequestID holds the string denoting the pull_request_id field in the database.
	FieldPullRequestID = "pull_request_id"
	// FieldPullRequestTitle holds the string denoting the pull_request_title field in the database.
	FieldPullRequestTitle = "pull_request_title"
	// FieldPullRequestBody holds the string denoting the pull_request_body field in the database.
	FieldPullRequestBody = "pull_request_body"
	// FieldPullRequestAuthorLogin holds the string denoting the pull_request_author_login field in the database.
	FieldPullRequestAuthorLogin = "pull_request_author_login"
	// FieldMessageID holds the string denoting the message_id field in the database.
	FieldMessageID = "message_id"
	// Table holds the table name of the prnotification in the database.
	Table = "pr_notifications"
)

// Columns holds all SQL columns for prnotification fields.
var Columns = []string{
	FieldID,
	FieldRepoID,
	FieldPullRequestID,
	FieldPullRequestTitle,
	FieldPullRequestBody,
	FieldPullRequestAuthorLogin,
	FieldMessageID,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultPullRequestTitle holds the default value on creation for the "pull_request_title" field.
	DefaultPullRequestTitle string
	// DefaultPullRequestBody holds the default value on creation for the "pull_request_body" field.
	DefaultPullRequestBody string
	// DefaultPullRequestAuthorLogin holds the default value on creation for the "pull_request_author_login" field.
	DefaultPullRequestAuthorLogin string
)

// OrderOption defines the ordering options for the PRNotification queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByRepoID orders the results by the repo_id field.
func ByRepoID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRepoID, opts...).ToFunc()
}

// ByPullRequestID orders the results by the pull_request_id field.
func ByPullRequestID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPullRequestID, opts...).ToFunc()
}

// ByPullRequestTitle orders the results by the pull_request_title field.
func ByPullRequestTitle(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPullRequestTitle, opts...).ToFunc()
}

// ByPullRequestBody orders the results by the pull_request_body field.
func ByPullRequestBody(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPullRequestBody, opts...).ToFunc()
}

// ByPullRequestAuthorLogin orders the results by the pull_request_author_login field.
func ByPullRequestAuthorLogin(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPullRequestAuthorLogin, opts...).ToFunc()
}

// ByMessageID orders the results by the message_id field.
func ByMessageID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldMessageID, opts...).ToFunc()
}
