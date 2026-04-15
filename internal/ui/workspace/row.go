package workspace

import (
	"github.com/idursun/jjui/internal/screen"
)

type row struct {
	Name    string
	Root    string
	Current bool
	Lines   []*rowLine
}

func (r *row) GetSearchableLines() []screen.SearchableLine {
	lines := make([]screen.SearchableLine, len(r.Lines))
	for i, line := range r.Lines {
		lines[i] = line
	}
	return lines
}

type rowLine struct {
	Segments []*screen.Segment
}

func (rl *rowLine) GetSegments() []*screen.Segment {
	return rl.Segments
}

func newRowLine(segments []*screen.Segment) rowLine {
	return rowLine{Segments: segments}
}
