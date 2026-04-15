package workspace

import (
	"io"
	"strings"

	"github.com/idursun/jjui/internal/screen"
)

func parseRows(reader io.Reader) []row {
	var rows []row
	rawSegments := screen.ParseFromReader(reader)

	for segmentedLine := range screen.BreakNewLinesIter(rawSegments) {
		rl := newRowLine(segmentedLine)
		// Each line in workspace list output represents a workspace.
		// The workspace name is the first segment before the ":"
		name := extractWorkspaceName(segmentedLine)
		rows = append(rows, row{
			Name:  name,
			Lines: []*rowLine{&rl},
		})
	}
	return rows
}

func extractWorkspaceName(segments []*screen.Segment) string {
	// The workspace list output format is: "name: change_id commit_id description"
	// We concatenate all segment text and extract the name before the first ":"
	var sb strings.Builder
	for _, seg := range segments {
		sb.WriteString(seg.Text)
	}
	line := sb.String()
	if idx := strings.Index(line, ":"); idx != -1 {
		return strings.TrimSpace(line[:idx])
	}
	return strings.TrimSpace(line)
}
