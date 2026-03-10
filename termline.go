package termline

import (
	"math"
	"sort"
	"strings"
	"time"
)

type Event struct {
	Label string    `json:"label"`
	At    time.Time `json:"at"`
	Depth int       `json:"depth"`
}

type Timeline struct {
	Width      int           `json:"width"`
	LeftMargin int           `json:"left_margin"`
	TickEvery  time.Duration `json:"tick_every"`
	TimeFormat string        `json:"time_format"`
}

// Options is an alias kept for backward compatibility.
//
// Deprecated: use Timeline instead.
type Options = Timeline

// Render draws the timeline as a string.
func (tl Timeline) Render(events []Event, start, end time.Time) string {
	if tl.Width <= 0 {
		tl.Width = 100
	}
	if tl.TickEvery <= 0 {
		tl.TickEvery = time.Hour
	}
	if tl.TimeFormat == "" {
		tl.TimeFormat = "15:04"
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].Depth != events[j].Depth {
			return events[i].Depth < events[j].Depth
		}
		return events[i].At.Before(events[j].At)
	})

	span := end.Sub(start)
	if span <= 0 {
		return ""
	}

	toX := func(t time.Time) int {
		if t.Before(start) {
			return 0
		}
		if t.After(end) {
			return tl.Width
		}
		ratio := float64(t.Sub(start)) / float64(span)
		return int(math.Round(ratio * float64(tl.Width)))
	}

	margin := strings.Repeat(" ", tl.LeftMargin)
	var b strings.Builder

	// ---- Tick labels
	tickLine := make([]rune, tl.Width+1)
	for i := range tickLine {
		tickLine[i] = ' '
	}

	firstTick := start.Truncate(tl.TickEvery)
	if firstTick.Before(start) {
		firstTick = firstTick.Add(tl.TickEvery)
	}

	for t := firstTick; !t.After(end); t = t.Add(tl.TickEvery) {
		x := toX(t)
		lbl := t.Format(tl.TimeFormat)
		pos := x - len(lbl)/2
		if pos < 0 {
			pos = 0
		}
		if pos+len(lbl) > len(tickLine) {
			pos = len(tickLine) - len(lbl)
		}
		for i, r := range lbl {
			tickLine[pos+i] = r
		}
	}

	b.WriteString(margin)
	b.WriteString(string(tickLine))
	b.WriteString("\n")

	// ---- Axis
	axis := make([]rune, tl.Width+1)
	for i := range axis {
		axis[i] = '-'
	}
	axis[0] = '|'
	axis[tl.Width] = '|'

	for t := firstTick; !t.After(end); t = t.Add(tl.TickEvery) {
		axis[toX(t)] = '|'
	}

	b.WriteString(margin)
	b.WriteString(string(axis))
	b.WriteString("\n\n")

	// ---- Events: greedy row packing
	type slot struct {
		e Event
		x int
		w int
	}
	var rows [][]slot

	for _, e := range events {
		x := toX(e.At)
		w := len("^ " + e.Label + " (" + e.At.Format(tl.TimeFormat) + ")")

		target := -1
		for r, row := range rows {
			fit := true
			for _, s := range row {
				if x < s.x+s.w+1 && s.x < x+w+1 {
					fit = false
					break
				}
			}
			if fit {
				target = r
				break
			}
		}

		p := slot{e, x, w}
		if target == -1 {
			rows = append(rows, []slot{p})
		} else {
			rows[target] = append(rows[target], p)
		}
	}

	for _, row := range rows {
		sort.Slice(row, func(i, j int) bool { return row[i].x < row[j].x })
		pos := 0
		b.WriteString(margin)
		for _, s := range row {
			if s.x > pos {
				b.WriteString(strings.Repeat(" ", s.x-pos))
			}
			b.WriteString("^ ")
			b.WriteString(s.e.Label)
			b.WriteString(" (")
			b.WriteString(s.e.At.Format(tl.TimeFormat))
			b.WriteString(")")
			pos = s.x + s.w
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderTimeline renders a timeline as a string.
//
// Deprecated: use Timeline.Render instead.
func RenderTimeline(events []Event, start, end time.Time, opt Options) string {
	return opt.Render(events, start, end)
}
