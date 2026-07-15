package jsonbridge

import "encoding/json"

func GoDumps(obj string) string {
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

func GoLoads(s string) string {
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
