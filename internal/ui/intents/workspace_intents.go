package intents

//jjui:bind scope=workspace action=move_up set=Delta:-1
//jjui:bind scope=workspace action=move_down set=Delta:1
//jjui:bind scope=workspace action=page_up set=Delta:-1,IsPage:true
//jjui:bind scope=workspace action=page_down set=Delta:1,IsPage:true
type WorkspaceNavigate struct {
	Delta  int
	IsPage bool
}

func (WorkspaceNavigate) isIntent() {}

//jjui:bind scope=ui action=open_workspace
type WorkspaceOpen struct{}

func (WorkspaceOpen) isIntent() {}

//jjui:bind scope=workspace action=close
type WorkspaceClose struct{}

func (WorkspaceClose) isIntent() {}

//jjui:bind scope=workspace action=add
type WorkspaceAdd struct{}

func (WorkspaceAdd) isIntent() {}

//jjui:bind scope=workspace action=forget
type WorkspaceForget struct {
	WorkspaceName string
}

func (WorkspaceForget) isIntent() {}

//jjui:bind scope=workspace action=update_stale
type WorkspaceUpdateStale struct{}

func (WorkspaceUpdateStale) isIntent() {}

//jjui:bind scope=workspace action=switch_workspace
type WorkspaceSwitch struct {
	WorkspaceName string
}

func (WorkspaceSwitch) isIntent() {}
