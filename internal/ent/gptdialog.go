// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/bot/internal/ent/gptdialog"
)

// GPTDialog is the model entity for the GPTDialog schema.
type GPTDialog struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Peer ID
	PeerID string `json:"peer_id,omitempty"`
	// Telegram message id of prompt message.
	PromptMsgID int `json:"prompt_msg_id,omitempty"`
	// Prompt message.
	PromptMsg string `json:"prompt_msg,omitempty"`
	// Telegram message id of sent message.
	GptMsgID int `json:"gpt_msg_id,omitempty"`
	// AI-generated message. Does not include prompt.
	GptMsg string `json:"gpt_msg,omitempty"`
	// Telegram thread's top message id.
	ThreadTopMsgID int `json:"thread_top_msg_id,omitempty"`
	// Message generation time. To simplify cleanup.
	CreatedAt    time.Time `json:"created_at,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*GPTDialog) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case gptdialog.FieldID, gptdialog.FieldPromptMsgID, gptdialog.FieldGptMsgID, gptdialog.FieldThreadTopMsgID:
			values[i] = new(sql.NullInt64)
		case gptdialog.FieldPeerID, gptdialog.FieldPromptMsg, gptdialog.FieldGptMsg:
			values[i] = new(sql.NullString)
		case gptdialog.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the GPTDialog fields.
func (gd *GPTDialog) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case gptdialog.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			gd.ID = int(value.Int64)
		case gptdialog.FieldPeerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field peer_id", values[i])
			} else if value.Valid {
				gd.PeerID = value.String
			}
		case gptdialog.FieldPromptMsgID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field prompt_msg_id", values[i])
			} else if value.Valid {
				gd.PromptMsgID = int(value.Int64)
			}
		case gptdialog.FieldPromptMsg:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field prompt_msg", values[i])
			} else if value.Valid {
				gd.PromptMsg = value.String
			}
		case gptdialog.FieldGptMsgID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field gpt_msg_id", values[i])
			} else if value.Valid {
				gd.GptMsgID = int(value.Int64)
			}
		case gptdialog.FieldGptMsg:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field gpt_msg", values[i])
			} else if value.Valid {
				gd.GptMsg = value.String
			}
		case gptdialog.FieldThreadTopMsgID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field thread_top_msg_id", values[i])
			} else if value.Valid {
				gd.ThreadTopMsgID = int(value.Int64)
			}
		case gptdialog.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				gd.CreatedAt = value.Time
			}
		default:
			gd.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the GPTDialog.
// This includes values selected through modifiers, order, etc.
func (gd *GPTDialog) Value(name string) (ent.Value, error) {
	return gd.selectValues.Get(name)
}

// Update returns a builder for updating this GPTDialog.
// Note that you need to call GPTDialog.Unwrap() before calling this method if this GPTDialog
// was returned from a transaction, and the transaction was committed or rolled back.
func (gd *GPTDialog) Update() *GPTDialogUpdateOne {
	return NewGPTDialogClient(gd.config).UpdateOne(gd)
}

// Unwrap unwraps the GPTDialog entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (gd *GPTDialog) Unwrap() *GPTDialog {
	_tx, ok := gd.config.driver.(*txDriver)
	if !ok {
		panic("ent: GPTDialog is not a transactional entity")
	}
	gd.config.driver = _tx.drv
	return gd
}

// String implements the fmt.Stringer.
func (gd *GPTDialog) String() string {
	var builder strings.Builder
	builder.WriteString("GPTDialog(")
	builder.WriteString(fmt.Sprintf("id=%v, ", gd.ID))
	builder.WriteString("peer_id=")
	builder.WriteString(gd.PeerID)
	builder.WriteString(", ")
	builder.WriteString("prompt_msg_id=")
	builder.WriteString(fmt.Sprintf("%v", gd.PromptMsgID))
	builder.WriteString(", ")
	builder.WriteString("prompt_msg=")
	builder.WriteString(gd.PromptMsg)
	builder.WriteString(", ")
	builder.WriteString("gpt_msg_id=")
	builder.WriteString(fmt.Sprintf("%v", gd.GptMsgID))
	builder.WriteString(", ")
	builder.WriteString("gpt_msg=")
	builder.WriteString(gd.GptMsg)
	builder.WriteString(", ")
	builder.WriteString("thread_top_msg_id=")
	builder.WriteString(fmt.Sprintf("%v", gd.ThreadTopMsgID))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(gd.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// GPTDialogs is a parsable slice of GPTDialog.
type GPTDialogs []*GPTDialog