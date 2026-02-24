package termline

import (
	"encoding/json"
	"fmt"
	"time"
)

type jsonPayload struct {
	Start   time.Time   `json:"start"`
	End     time.Time   `json:"end"`
	Options jsonOptions `json:"options"`
	Events  []Event     `json:"events"`
}

type jsonOptions struct {
	Width      int    `json:"width"`
	LeftMargin int    `json:"left_margin"`
	TickEvery  string `json:"tick_every"`
	TimeFormat string `json:"time_format"`
}

// ParseJSON parses a full timeline payload from JSON and returns
// the events, start/end times, and rendering options.
//
// Expected format:
//
//	{
//	  "start":   "2025-06-10T08:00:00Z",
//	  "end":     "2025-06-10T19:00:00Z",
//	  "options": { "width": 80, "left_margin": 2, "tick_every": "1h", "time_format": "15:04" },
//	  "events":  [
//	    { "label": "Standup", "at": "2025-06-10T09:00:00Z", "depth": 0 }
//	  ]
//	}
//
// Timestamps must be RFC3339. tick_every accepts any string valid for time.ParseDuration (e.g. "1h", "30m").
func ParseJSON(data []byte) (events []Event, start, end time.Time, opt Options, err error) {
	var p jsonPayload
	if err = json.Unmarshal(data, &p); err != nil {
		return
	}

	events = p.Events
	start = p.Start
	end = p.End
	opt = Options{
		Width:      p.Options.Width,
		LeftMargin: p.Options.LeftMargin,
		TimeFormat: p.Options.TimeFormat,
	}

	if p.Options.TickEvery != "" {
		opt.TickEvery, err = time.ParseDuration(p.Options.TickEvery)
		if err != nil {
			err = fmt.Errorf("tick_every: %w", err)
			return
		}
	}

	return
}
