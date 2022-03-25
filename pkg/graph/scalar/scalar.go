package scalar

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

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

// if the type referenced in .gqlgen.yml is a function that returns a marshaller we can use it to encode and decode
// onto any existing go type.
func MarshalTimestamp(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatInt(t.Unix(), 10))
	})
}

// Unmarshal{Typename} is only required if the scalar appears as an input. The raw values have already been decoded
// from json into int/float64/bool/nil/map[string]interface/[]interface
func UnmarshalTimestamp(v interface{}) (time.Time, error) {
	if tmpStr, ok := v.(int64); ok {
		return time.Unix(tmpStr, 0), nil
	}
	return time.Time{}, errors.New("time should be a unix timestamp")
}
