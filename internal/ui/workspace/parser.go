package workspace

import (
	"io"
	"strings"

	"github.com/idursun/jjui/internal/screen"
)

func parseRows(reader io.Reader, roots map[string]string) []row {
	var rows []row
	rawSegments := screen.ParseFromReader(reader)

	for segmentedLine := range screen.BreakNewLinesIter(rawSegments) {
		rl := newRowLine(segmentedLine)
		name := extractWorkspaceName(segmentedLine)
		root := roots[name]
		rows = append(rows, row{
			Name:  name,
			Root:  root,
			Lines: []*rowLine{&rl},
		})
	}
	return rows
}

func extractWorkspaceName(segments []*screen.Segment) string {
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

// parseRoots parses the output of WorkspaceListRoots into a name→root map.
// Each line is "name\troot". Lines where root starts with "<Error:" are skipped.
func parseRoots(output []byte) map[string]string {
	roots := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			root := strings.TrimSpace(parts[1])
			if name != "" && root != "" && !strings.HasPrefix(root, "<Error:") {
				roots[name] = root
			}
		}
	}
	return roots
}
