package workspace

import (
	"bytes"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/actions"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/context"
	"github.com/idursun/jjui/internal/ui/dispatch"
	"github.com/idursun/jjui/internal/ui/input"
	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/idursun/jjui/internal/ui/render"
)

type updateWorkspaceMsg struct {
	Rows []row
}

type WorkspaceClickedMsg struct {
	Index int
}

type WorkspaceScrollMsg struct {
	Delta      int
	Horizontal bool
}

func (o WorkspaceScrollMsg) SetDelta(delta int, horizontal bool) tea.Msg {
	return WorkspaceScrollMsg{Delta: delta, Horizontal: horizontal}
}

// SwitchWorkspaceMsg is emitted when the user wants to switch to a different workspace.
// The root UI handles this by re-exec'ing jjui pointed at the workspace root.
type SwitchWorkspaceMsg struct {
	WorkspaceRoot string
}

var _ common.ImmediateModel = (*Model)(nil)

type Model struct {
	context          *context.MainContext
	listRenderer     *render.ListRenderer
	rows             []row
	cursor           int
	textStyle        lipgloss.Style
	selectedStyle    lipgloss.Style
	ensureCursorView bool
	pendingAdd       bool
}

func (m *Model) Len() int {
	if m.rows == nil {
		return 0
	}
	return len(m.rows)
}

func (m *Model) Cursor() int {
	return m.cursor
}

func (m *Model) SetCursor(index int) {
	if index >= 0 && index < len(m.rows) {
		m.cursor = index
		m.ensureCursorView = true
	}
}

func (m *Model) Scopes() []dispatch.Scope {
	return []dispatch.Scope{
		{
			Name:    actions.ScopeWorkspace,
			Leak:    dispatch.LeakAll,
			Handler: m,
		},
	}
}

func (m *Model) Init() tea.Cmd {
	return m.load()
}

func (m *Model) Scroll(delta int) tea.Cmd {
	m.ensureCursorView = false
	currentStart := m.listRenderer.GetScrollOffset()
	desiredStart := currentStart + delta
	m.listRenderer.SetScrollOffset(desiredStart)
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case intents.Intent:
		cmd, _ := m.HandleIntent(msg)
		return cmd
	case updateWorkspaceMsg:
		m.rows = msg.Rows
		return nil
	case input.SelectedMsg:
		if m.pendingAdd {
			m.pendingAdd = false
			path := strings.TrimSpace(msg.Value)
			if path == "" {
				return nil
			}
			return m.runAdd(path)
		}
	case input.CancelledMsg:
		m.pendingAdd = false
	case common.CommandCompletedMsg:
		return m.load()
	case WorkspaceClickedMsg:
		if msg.Index >= 0 && msg.Index < len(m.rows) {
			m.cursor = msg.Index
			m.ensureCursorView = true
			return nil
		}
	case WorkspaceScrollMsg:
		if msg.Horizontal {
			return nil
		}
		return m.Scroll(msg.Delta)
	}
	return nil
}

func (m *Model) HandleIntent(intent intents.Intent) (tea.Cmd, bool) {
	switch intent := intent.(type) {
	case intents.WorkspaceNavigate:
		return m.navigate(intent.Delta, intent.IsPage), true
	case intents.WorkspaceClose:
		return m.close(), true
	case intents.WorkspaceAdd:
		return m.add(), true
	case intents.WorkspaceForget:
		return m.forget(intent), true
	case intents.WorkspaceUpdateStale:
		return m.updateStale(), true
	case intents.WorkspaceSwitch:
		return m.switchWorkspace(intent), true
	}
	return nil, false
}

func (m *Model) navigate(delta int, page bool) tea.Cmd {
	if len(m.rows) == 0 {
		return nil
	}

	step := delta
	if page {
		firstRowIndex := m.listRenderer.GetFirstRowIndex()
		lastRowIndex := m.listRenderer.GetLastRowIndex()
		span := max(lastRowIndex-firstRowIndex-1, 1)
		if step < 0 {
			step = -span
		} else {
			step = span
		}
	}

	totalItems := len(m.rows)
	newCursor := m.cursor + step
	if newCursor < 0 {
		newCursor = 0
	} else if newCursor >= totalItems {
		newCursor = totalItems - 1
	}

	m.SetCursor(newCursor)
	return nil
}

func (m *Model) close() tea.Cmd {
	return tea.Batch(common.Close, common.Refresh)
}

func (m *Model) add() tea.Cmd {
	m.pendingAdd = true
	return func() tea.Msg {
		return common.ShowInputMsg{
			Title:  "Add Workspace",
			Prompt: "Path: ",
		}
	}
}

func (m *Model) runAdd(path string) tea.Cmd {
	return func() tea.Msg {
		return common.ExecMsg{
			Line: fmt.Sprintf("workspace add %s", path),
			Mode: common.ExecJJ,
		}
	}
}

func (m *Model) forget(intent intents.WorkspaceForget) tea.Cmd {
	name := intent.WorkspaceName
	if name == "" {
		if len(m.rows) == 0 {
			return nil
		}
		name = m.rows[m.cursor].Name
	}
	if name == "" {
		return nil
	}
	return m.context.RunCommand(jj.WorkspaceForget(name), func() tea.Msg {
		return common.CommandCompletedMsg{}
	})
}

func (m *Model) updateStale() tea.Cmd {
	return m.context.RunCommand(jj.WorkspaceUpdateStale(), func() tea.Msg {
		return common.CommandCompletedMsg{}
	})
}

func (m *Model) switchWorkspace(intent intents.WorkspaceSwitch) tea.Cmd {
	name := intent.WorkspaceName
	if name == "" {
		if len(m.rows) == 0 {
			return nil
		}
		name = m.rows[m.cursor].Name
	}
	if name == "" {
		return nil
	}

	return func() tea.Msg {
		output, err := m.context.RunCommandImmediate(jj.WorkspaceRoot(name))
		if err != nil {
			return intents.AddMessage{Text: fmt.Sprintf("Failed to get workspace root: %v", err), Err: err}
		}
		root := strings.TrimSpace(string(output))
		if root == "" {
			return intents.AddMessage{Text: "Workspace root is empty", Err: fmt.Errorf("empty workspace root")}
		}
		return SwitchWorkspaceMsg{WorkspaceRoot: root}
	}
}

func (m *Model) ViewRect(dl *render.DisplayContext, box layout.Box) {
	if m.rows == nil {
		content := lipgloss.Place(box.R.Dx(), box.R.Dy(), lipgloss.Center, lipgloss.Center, "loading")
		dl.AddDraw(box.R, content, 0)
		return
	}

	if len(m.rows) == 0 {
		content := lipgloss.Place(box.R.Dx(), box.R.Dy(), lipgloss.Center, lipgloss.Center, "no workspaces")
		dl.AddDraw(box.R, content, 0)
		return
	}

	measure := func(index int) int {
		return len(m.rows[index].Lines)
	}

	renderItem := func(dl *render.DisplayContext, index int, itemRect layout.Rectangle) {
		row := m.rows[index]
		isSelected := index == m.cursor
		styleOverride := m.textStyle
		if isSelected {
			styleOverride = m.selectedStyle
		}

		y := itemRect.Min.Y
		for _, line := range row.Lines {
			var content bytes.Buffer
			for _, segment := range line.Segments {
				text := segment.Text
				style := segment.Style.Inherit(styleOverride)
				content.WriteString(style.Render(text))
			}
			lineContent := lipgloss.PlaceHorizontal(itemRect.Dx(), 0, content.String(), lipgloss.WithWhitespaceStyle(styleOverride))
			lineRect := layout.Rect(itemRect.Min.X, y, itemRect.Dx(), 1)
			dl.AddDraw(lineRect, lineContent, 0)
			y++
		}
	}

	clickMsg := func(index int, _ tea.Mouse) render.ClickMessage {
		return WorkspaceClickedMsg{Index: index}
	}

	m.listRenderer.Render(
		dl,
		box,
		len(m.rows),
		m.cursor,
		m.ensureCursorView,
		measure,
		renderItem,
		clickMsg,
	)
	m.listRenderer.RegisterScroll(dl, box)

	m.ensureCursorView = false
}

func (m *Model) load() tea.Cmd {
	return func() tea.Msg {
		output, err := m.context.RunCommandImmediate(jj.WorkspaceList())
		if err != nil {
			panic(err)
		}

		rows := parseRows(bytes.NewReader(output))
		return updateWorkspaceMsg{Rows: rows}
	}
}

func New(context *context.MainContext) *Model {
	m := &Model{
		context:       context,
		rows:          nil,
		cursor:        0,
		textStyle:     common.DefaultPalette.Get("oplog text"),
		selectedStyle: common.DefaultPalette.Get("oplog selected"),
	}
	m.listRenderer = render.NewListRenderer(WorkspaceScrollMsg{})
	return m
}

// SelectedWorkspaceName returns the name of the workspace currently under the cursor.
func (m *Model) SelectedWorkspaceName() string {
	if m.cursor >= 0 && m.cursor < len(m.rows) {
		return m.rows[m.cursor].Name
	}
	return ""
}
