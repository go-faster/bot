package action

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Action struct {
	Entity       string `json:"entity"`
	ID           int    `json:"id"`
	RepositoryID int64  `json:"repository_id"`
	Type         string `json:"action"`
}

func (a Action) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s(%s=%d)", a.Type, a.Entity, a.ID))
	if a.RepositoryID != 0 {
		b.WriteString(fmt.Sprintf(" r=%d", a.RepositoryID))
	}
	return b.String()
}

func Marshal(a Action) []byte {
	b, _ := json.Marshal(a)
	return b
}
