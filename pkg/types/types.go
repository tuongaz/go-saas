package types

import (
	"encoding/json"
)

type M map[string]any

func (m *M) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
