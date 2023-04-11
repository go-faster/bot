package action

import "encoding/json"

type Action struct {
	Entity string `json:"entity"`
	ID     int    `json:"id"`
	Action string `json:"action"`
}

func Marshal(a Action) []byte {
	b, _ := json.Marshal(a)
	return b
}
