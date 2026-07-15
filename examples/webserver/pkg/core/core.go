package core

import "encoding/json"

//export Dumps
func Dumps(obj string) string {
	var v any
	if err := json.Unmarshal([]byte(obj), &v); err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	return string(b)
}

//export Loads
func Loads(s string) string {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	return string(b)
}
