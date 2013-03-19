package errplane

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Name      string
	Value     float64
	Context   interface{}
	Timestamp *time.Time
}

func (e *Event) Line() []byte {

	// Default timestamp == 'now'
	ts := ""
	if e.Timestamp != nil {
		// TODO: what's the time format?
		ts = "now"
	} else {
		ts = "now"
	}

	// context
	var ctx string
	if e.Context != nil {
		jsonBytes, err := json.Marshal(e.Context)
		if err == nil {
			ctx = base64.StdEncoding.EncodeToString(jsonBytes)
		}
	}

	line := fmt.Sprintf("%v %v %v", e.Name, e.Value, ts)
	if ctx != "" {
		line = fmt.Sprintf("%s %s", line, ctx)
	}

	return []byte(line)
}
