package termline

import (
	"encoding/json"
	"fmt"
	"time"
)

type jsonData struct {
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	Events []Event   `json:"events"`
}

type jsonConfig struct {
	Width      int    `json:"width"`
	LeftMargin int    `json:"left_margin"`
	TickEvery  string `json:"tick_every"`
	TimeFormat string `json:"time_format"`
}

// NewFromJSON creates a Timeline from a JSON config blob.
//
// Expected format:
//
//	{ "width": 80, "left_margin": 2, "tick_every": "1h", "time_format": "15:04" }
//
// tick_every accepts any string valid for time.ParseDuration (e.g. "1h", "30m").
func NewFromJSON(data []byte) (Timeline, error) {
	var c jsonConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return Timeline{}, err
	}

	tl := Timeline{
		Width:      c.Width,
		LeftMargin: c.LeftMargin,
		TimeFormat: c.TimeFormat,
	}

	if c.TickEvery != "" {
		d, err := time.ParseDuration(c.TickEvery)
		if err != nil {
			return Timeline{}, fmt.Errorf("tick_every: %w", err)
		}
		tl.TickEvery = d
	}

	return tl, nil
}

// RenderJSON parses event data from JSON and renders the timeline.
//
// Expected format:
//
//	{
//	  "start":  "2025-06-10T08:00:00Z",
//	  "end":    "2025-06-10T19:00:00Z",
//	  "events": [
//	    { "label": "Standup", "at": "2025-06-10T09:00:00Z", "depth": 0 }
//	  ]
//	}
//
// Timestamps must be RFC3339.
func (tl Timeline) RenderJSON(data []byte) (string, error) {
	var d jsonData
	if err := json.Unmarshal(data, &d); err != nil {
		return "", err
	}
	return tl.Render(d.Events, d.Start, d.End), nil
}
