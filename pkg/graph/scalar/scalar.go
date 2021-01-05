package scalar

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

// Map implements the below interfaces to satisfy both
// sql parser and graphql parser
var _ sql.Scanner = (*Map)(nil)
var _ driver.Valuer = (*Map)(nil)
var _ graphql.Marshaler = (*Map)(nil)
var _ graphql.Unmarshaler = (*Map)(nil)

// Map is an costom scalar type used for graphql and orm model
type Map map[string]interface{}

func (t Map) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *Map) Scan(src interface{}) error {
	buf, ok := src.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(buf, t)
}

func (t Map) MarshalGQL(w io.Writer) {
	buf, _ := json.Marshal(t)
	w.Write(buf)
}

func (t *Map) UnmarshalGQL(v interface{}) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, t)
}
