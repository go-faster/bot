// Code generated by ent, DO NOT EDIT.

package telegramsession

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the telegramsession type in the database.
	Label = "telegram_session"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldData holds the string denoting the data field in the database.
	FieldData = "data"
	// Table holds the table name of the telegramsession in the database.
	Table = "telegram_sessions"
)

// Columns holds all SQL columns for telegramsession fields.
var Columns = []string{
	FieldID,
	FieldData,
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

// OrderOption defines the ordering options for the TelegramSession queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}
