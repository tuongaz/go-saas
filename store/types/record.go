package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Record represents a database record as a map of field names to values
type Record map[string]any

// String returns the value for the given key as a string.
// Returns empty string if the key doesn't exist or the value cannot be converted to string.
func (r Record) String(key string) string {
	if v, ok := r[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case []byte:
			return string(val)
		case fmt.Stringer:
			return val.String()
		default:
			return fmt.Sprint(val)
		}
	}
	return ""
}

// Int returns the value for the given key as an int.
// Returns 0 if the key doesn't exist or the value cannot be converted to int.
func (r Record) Int(key string) int {
	if v, ok := r[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
	}
	return 0
}

// Int64 returns the value for the given key as an int64.
// Returns 0 if the key doesn't exist or the value cannot be converted to int64.
func (r Record) Int64(key string) int64 {
	if v, ok := r[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		case string:
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				return i
			}
		}
	}
	return 0
}

// Float64 returns the value for the given key as a float64.
// Returns 0 if the key doesn't exist or the value cannot be converted to float64.
func (r Record) Float64(key string) float64 {
	if v, ok := r[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

// Bool returns the value for the given key as a bool.
// Returns false if the key doesn't exist or the value cannot be converted to bool.
func (r Record) Bool(key string) bool {
	if v, ok := r[key]; ok {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			if b, err := strconv.ParseBool(val); err == nil {
				return b
			}
		}
	}
	return false
}

// Time returns the value for the given key as a time.Time.
// Returns zero time if the key doesn't exist or the value cannot be converted to time.Time.
func (r Record) Time(key string) time.Time {
	if v, ok := r[key]; ok {
		switch val := v.(type) {
		case time.Time:
			return val
		case string:
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func (r Record) Normalise() {
	// Handle special cases like JSONB and timestamps
	for key, value := range r {
		switch v := value.(type) {
		case []byte:
			var jsonData any
			if err := json.Unmarshal(v, &jsonData); err == nil {
				r[key] = jsonData
			}
		case time.Time:
			// Convert time.Time to string for consistency
			r[key] = v.Format(time.RFC3339)
		}
	}
}

func (r Record) PrepareForDB() (keys []string, values []any, placeholders []string, err error) {
	keys = make([]string, 0, len(r))
	values = make([]any, 0, len(r))
	placeholders = make([]string, 0, len(r))

	for k, v := range r {
		keys = append(keys, k)

		switch value := v.(type) {
		case string, int, int64, float64, bool:
			values = append(values, v)
		case nil:
			// Explicitly handle nil as NULL
			values = append(values, nil)
		case *string, *int, *int64, *float64, *bool:
			rv := reflect.ValueOf(value)
			if rv.IsNil() {
				values = append(values, nil)
			} else {
				values = append(values, rv.Elem().Interface())
			}
		default:
			jsonBytes, marshalErr := json.Marshal(value)
			if marshalErr != nil {
				return nil, nil, nil, fmt.Errorf("failed to marshal value to JSON for key %s: %w", k, marshalErr)
			}
			values = append(values, string(jsonBytes))
		}

		placeholders = append(placeholders, fmt.Sprintf("$%d", len(placeholders)+1))
	}

	return keys, values, placeholders, nil
}

func (r Record) Get(key string) any {
	return r[key]
}

func (r Record) Decode(obj any) error {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("encode to json: %w", err)
	}
	if err := json.Unmarshal(jsonData, obj); err != nil {
		return fmt.Errorf("decode json to struct: %w", err)
	}
	return nil
}
