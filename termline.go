package termline

import (
	"math"
	"sort"
	"strings"
	"time"
)

type Event struct {
	Label string
	At    time.Time
	Depth int
}

type Options struct {
	Width      int
	LeftMargin int
	TickEvery  time.Duration
	TimeFormat string
}

func RenderTimeline(events []Event, start, end time.Time, opt Options) string {
	if opt.Width <= 0 {
		opt.Width = 100
	}
	if opt.TickEvery <= 0 {
		opt.TickEvery = time.Hour
	}
	if opt.TimeFormat == "" {
		opt.TimeFormat = "15:04"
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
			return opt.Width
		}
		ratio := float64(t.Sub(start)) / float64(span)
		return int(math.Round(ratio * float64(opt.Width)))
	}

	margin := strings.Repeat(" ", opt.LeftMargin)
	var b strings.Builder

	// ---- Tick labels
	tickLine := make([]rune, opt.Width+1)
	for i := range tickLine {
		tickLine[i] = ' '
	}

	firstTick := start.Truncate(opt.TickEvery)
	if firstTick.Before(start) {
		firstTick = firstTick.Add(opt.TickEvery)
	}

	for t := firstTick; !t.After(end); t = t.Add(opt.TickEvery) {
		x := toX(t)
		lbl := t.Format(opt.TimeFormat)
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
	axis := make([]rune, opt.Width+1)
	for i := range axis {
		axis[i] = '-'
	}
	axis[0] = '|'
	axis[opt.Width] = '|'

	for t := firstTick; !t.After(end); t = t.Add(opt.TickEvery) {
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
		w := len("^ " + e.Label + " (" + e.At.Format(opt.TimeFormat) + ")")

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
			b.WriteString(s.e.At.Format(opt.TimeFormat))
			b.WriteString(")")
			pos = s.x + s.w
		}
		b.WriteString("\n")
	}

	return b.String()
}
