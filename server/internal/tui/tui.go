package tui

import (
	"fmt"
	"os"
	"strings"

	"birdactyl-panel-backend/internal/logger"
	"birdactyl-panel-backend/internal/models"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type logMsg string

type viewState int

const (
	viewLogs viewState = iota
	viewUserList
	viewUserInfo
	viewUserActivity
	viewActivityDetails
	viewAdminActions
	viewAdminLimits
	viewAdminConfirm
	viewAdminPrompt
	viewServerList
	viewServerInfo
	viewServerAdminActions
	viewServerLimits
	viewServerPrompt
	viewServerCreate
	viewNodeList
	viewNodeInfo
	viewNodeAdminActions
	viewNodeCreate
	viewNodePair
	viewNodePairWait
	viewDatabaseHostList
	viewDatabaseHostInfo
	viewDatabaseHostAdminActions
	viewDatabaseHostCreate
	viewDbHostPrompt
	viewMountList
	viewMountInfo
	viewMountAdminActions
	viewMountCreate
	viewMountPrompt
)

type model struct {
	viewport            viewport.Model
	textInput           textinput.Model
	userList            list.Model
	userInfoList        list.Model
	userActivityList    list.Model
	activityDetailsList list.Model
	adminActionsList    list.Model
	adminLimitsList     list.Model
	serverList          list.Model
	serverInfoList      list.Model
	serverAdminList     list.Model
	serverLimitsList    list.Model
	logs                []string
	history             []string
	historyIdx          int
	ready               bool
	state               viewState

	createServerInputs []textinput.Model
	createServerFocus  int

	nodeList         list.Model
	nodeInfoList     list.Model
	nodeAdminList    list.Model
	createNodeInputs []textinput.Model
	createNodeFocus  int
	pairNodeInputs   []textinput.Model
	pairNodeFocus    int
	pairNodeCode     string

	dbHostList         list.Model
	dbHostInfoList     list.Model
	dbHostAdminList    list.Model
	createDbHostInputs []textinput.Model
	createDbHostFocus  int
	targetDbHost       models.DatabaseHost

	mountList         list.Model
	mountInfoList     list.Model
	mountAdminList    list.Model
	createMountInputs []textinput.Model
	createMountFocus  int
	targetMount       models.Mount

	targetUser         models.User
	targetServer       models.Server
	pendingConfirmDesc string
	pendingConfirmCmd  tea.Cmd
	promptDesc         string
	promptField        string
	previousAdminState viewState
	isGlobalActivity   bool
}

type userItem struct {
	models.User
}

func (i userItem) Title() string       { return i.Username }
func (i userItem) Description() string { return fmt.Sprintf("%s | %s", i.ID.String(), i.Email) }
func (i userItem) FilterValue() string { return i.Username + " " + i.Email + " " + i.ID.String() }

type serverItem struct {
	models.Server
}

func (s serverItem) Title() string       { return s.Name }
func (s serverItem) Description() string { return fmt.Sprintf("%s | %s", s.ID.String(), s.Status) }
func (s serverItem) FilterValue() string {
	return s.Name + " " + string(s.Status) + " " + s.ID.String()
}

type showUserListMsg []list.Item
type showUserInfoMsg struct {
	title string
	items []list.Item
}
type showUserActivityMsg struct {
	items    []list.Item
	isGlobal bool
}
type showActivityDetailsMsg struct {
	title string
	items []list.Item
}
type showAdminActionsMsg struct {
	user  models.User
	items []list.Item
}
type showAdminLimitsMsg struct {
	user  models.User
	items []list.Item
}
type showServerListMsg []list.Item
type showServerInfoMsg struct {
	title string
	items []list.Item
}
type showServerAdminActionsMsg struct {
	server models.Server
	items  []list.Item
}
type showServerLimitsMsg struct {
	server models.Server
	items  []list.Item
}
type showServerCreateMsg struct{}
type showDatabaseHostListMsg []list.Item
type showDatabaseHostInfoMsg struct {
	title string
	items []list.Item
}
type showDatabaseHostAdminActionsMsg struct {
	dbHost models.DatabaseHost
	items  []list.Item
}
type showDatabaseHostCreateMsg struct{}
type showMountListMsg []list.Item
type showMountInfoMsg struct {
	title string
	items []list.Item
}
type showMountAdminActionsMsg struct {
	mount models.Mount
	items []list.Item
}
type showMountCreateMsg struct{}
type askConfirmMsg struct {
	desc string
	cmd  tea.Cmd
}
type askPromptMsg struct {
	desc        string
	field       string
	placeholder string
}
type askServerPromptMsg struct {
	desc        string
	field       string
	placeholder string
}
type askDbHostPromptMsg struct {
	desc        string
	field       string
	placeholder string
}
type askMountPromptMsg struct {
	desc        string
	field       string
	placeholder string
}
type actionDoneMsg string
type serverActionDoneMsg string
type dbHostActionDoneMsg string
type mountActionDoneMsg string
type errorMessageMsg string

type refreshDbBothMsg struct {
	info  showDatabaseHostInfoMsg
	admin showDatabaseHostAdminActionsMsg
}

type refreshMountBothMsg struct {
	info  showMountInfoMsg
	admin showMountAdminActionsMsg
}

type refreshBothMsg struct {
	info   showUserInfoMsg
	admin  showAdminActionsMsg
	limits showAdminLimitsMsg
	users  []list.Item
}

type refreshServerBothMsg struct {
	info    showServerInfoMsg
	admin   showServerAdminActionsMsg
	limits  showServerLimitsMsg
	servers []list.Item
}

type infoItem struct {
	topic     string
	desc      string
	actionCmd tea.Cmd
}

func (i infoItem) Title() string       { return i.topic }
func (i infoItem) Description() string { return i.desc }
func (i infoItem) FilterValue() string { return i.topic + " " + i.desc }

type activityItem struct {
	models.ActivityLog
}

func (a activityItem) Title() string {
	return fmt.Sprintf("[%s] %s", a.CreatedAt.Format("01-02 15:04"), a.ActivityLog.Action)
}
func (a activityItem) Description() string {
	if a.IP == "" {
		return a.ActivityLog.Description
	}
	return fmt.Sprintf("IP: %s | %s", a.IP, a.ActivityLog.Description)
}
func (a activityItem) FilterValue() string {
	return a.ActivityLog.Action + " " + a.ActivityLog.Description
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter command..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Birdactyl Users"

	iList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	iList.Title = "User Info"

	aList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	aList.Title = "Activity Logs"

	dList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	dList.Title = "Activity Record Details"

	adminList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	adminList.Title = "User Admin Actions"

	adminLimitsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	adminLimitsList.Title = "User Resource Limits"

	sList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sList.Title = "Birdactyl Servers"

	siList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	siList.Title = "Server Info"

	saList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	saList.Title = "Server Admin Actions"

	slList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	slList.Title = "Server Resource Limits"

	cInputs := make([]textinput.Model, 7)
	placeholders := []string{"Server Name", "Owner Username/UUID", "Node Name/UUID", "Package Name/UUID", "RAM Limit (1024)", "CPU Limit (100)", "Disk Limit (5000)"}
	for i := range cInputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 64
		t.Width = 40
		if i == 0 {
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		}
		cInputs[i] = t
	}

	nlList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	nlList.Title = "Nodes"

	niList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	niList.Title = "Node Info"

	naList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	naList.Title = "Node Admin Actions"

	cpInputs := make([]textinput.Model, 3)
	cnInputs := make([]textinput.Model, 3)
	nodePlaceholders := []string{"Node Name", "FQDN (e.g. node1.example.com)", "Port (e.g. 8443)"}
	for i := range cpInputs {
		t := textinput.New()
		t.Placeholder = nodePlaceholders[i]
		t.CharLimit = 64
		t.Width = 40

		t2 := textinput.New()
		t2.Placeholder = nodePlaceholders[i]
		t2.CharLimit = 64
		t2.Width = 40

		if i == 0 {
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

			t2.Focus()
			t2.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t2.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		}
		cpInputs[i] = t
		cnInputs[i] = t2
	}

	dhList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	dhList.Title = "Database Hosts"

	dhiList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	dhiList.Title = "Database Host Info"

	dhaList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	dhaList.Title = "Database Host Admin Actions"

	cdhInputs := make([]textinput.Model, 6)
	dhPlaceholders := []string{"Name (Main DB)", "Host (db.example.com)", "Port (3306)", "Username (root)", "Password", "Max Databases (0 = unlimited)"}
	for i := range cdhInputs {
		t := textinput.New()
		t.Placeholder = dhPlaceholders[i]
		t.CharLimit = 64
		t.Width = 40
		if i == 4 {
			t.EchoMode = textinput.EchoPassword
		}
		if i == 0 {
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		}
		cdhInputs[i] = t
	}

	mList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	mList.Title = "Global Mounts"

	miList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	miList.Title = "Mount Info"

	maList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	maList.Title = "Mount Admin Actions"

	cmInputs := make([]textinput.Model, 4)
	mPlaceholders := []string{"Name", "Description", "Source Path (/var/lib/mount)", "Target Path (/home/container/mnt)"}
	for i := range cmInputs {
		t := textinput.New()
		t.Placeholder = mPlaceholders[i]
		t.CharLimit = 128
		t.Width = 40
		if i == 0 {
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		}
		cmInputs[i] = t
	}

	return model{
		textInput:           ti,
		userList:            l,
		userInfoList:        iList,
		userActivityList:    aList,
		activityDetailsList: dList,
		adminActionsList:    adminList,
		adminLimitsList:     adminLimitsList,
		serverList:          sList,
		serverInfoList:      siList,
		serverAdminList:     saList,
		serverLimitsList:    slList,
		createServerInputs:  cInputs,
		createServerFocus:   0,
		nodeList:            nlList,
		nodeInfoList:        niList,
		nodeAdminList:       naList,
		pairNodeInputs:      cpInputs,
		pairNodeFocus:       0,
		createNodeInputs:    cnInputs,
		pairNodeCode:        "",
		dbHostList:          dhList,
		dbHostInfoList:      dhiList,
		dbHostAdminList:     dhaList,
		createDbHostInputs:  cdhInputs,
		createDbHostFocus:   0,
		mountList:           mList,
		mountInfoList:       miList,
		mountAdminList:      maList,
		createMountInputs:   cmInputs,
		createMountFocus:    0,
		logs:                []string{},
		history:             []string{},
		historyIdx:          0,
		state:               viewLogs,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, waitForLog())
}

func waitForLog() tea.Cmd {
	return func() tea.Msg {
		if logger.LogChannel != nil {
			msg, ok := <-logger.LogChannel
			if ok {
				return logMsg(msg)
			}
		}
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.state == viewLogs || m.state == viewAdminConfirm {
			}
			if m.state == viewLogs {
				return m, tea.Quit
			}
		case tea.KeyUp:
			if m.state == viewLogs && len(m.history) > 0 && m.historyIdx > 0 {
				m.historyIdx--
				m.textInput.SetValue(m.history[m.historyIdx])
				m.textInput.SetCursor(len(m.history[m.historyIdx]))
				return m, nil
			}
		case tea.KeyDown:
			if m.state == viewLogs && len(m.history) > 0 && m.historyIdx < len(m.history) {
				m.historyIdx++
				if m.historyIdx == len(m.history) {
					m.textInput.SetValue("")
				} else {
					m.textInput.SetValue(m.history[m.historyIdx])
					m.textInput.SetCursor(len(m.history[m.historyIdx]))
				}
				return m, nil
			}
		case tea.KeyEnter:
			if m.state == viewLogs {
				cmdStr := strings.TrimSpace(m.textInput.Value())
				if cmdStr != "" {
					if len(m.history) == 0 || m.history[len(m.history)-1] != cmdStr {
						m.history = append(m.history, cmdStr)
					}
					m.historyIdx = len(m.history)
					cmd = handleCommand(cmdStr)
					m.textInput.SetValue("")
				}
				return m, cmd
			}
		}

	case tea.WindowSizeMsg:
		headerHeight := 0
		footerHeight := 3
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.SetContent(strings.Join(m.logs, "\n"))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
		m.userList.SetSize(msg.Width, msg.Height)
		m.userInfoList.SetSize(msg.Width, msg.Height)
		m.userActivityList.SetSize(msg.Width, msg.Height)
		m.activityDetailsList.SetSize(msg.Width, msg.Height)
		m.adminActionsList.SetSize(msg.Width, msg.Height)
		m.adminLimitsList.SetSize(msg.Width, msg.Height)
		m.serverList.SetSize(msg.Width, msg.Height)
		m.serverInfoList.SetSize(msg.Width, msg.Height)
		m.serverAdminList.SetSize(msg.Width, msg.Height)
		m.serverLimitsList.SetSize(msg.Width, msg.Height)
		m.nodeList.SetSize(msg.Width, msg.Height)
		m.nodeInfoList.SetSize(msg.Width, msg.Height)
		m.nodeAdminList.SetSize(msg.Width, msg.Height)
		m.dbHostList.SetSize(msg.Width, msg.Height)
		m.dbHostInfoList.SetSize(msg.Width, msg.Height)
		m.dbHostAdminList.SetSize(msg.Width, msg.Height)
		m.mountList.SetSize(msg.Width, msg.Height)
		m.mountInfoList.SetSize(msg.Width, msg.Height)
		m.mountAdminList.SetSize(msg.Width, msg.Height)

	case errorMessageMsg:
		logger.TUIOut(string(msg))
		m.state = viewLogs
		return m, nil

	case actionDoneMsg:
		logger.TUIOut(string(msg))
		return m, refreshAdminStateCmd(m.targetUser.ID.String())

	case serverActionDoneMsg:
		logger.TUIOut(string(msg))
		return m, refreshServerAdminStateCmd(m.targetServer.ID.String())

	case dbHostActionDoneMsg:
		logger.TUIOut(string(msg))
		return m, refreshDbHostStateCmd(m.targetDbHost.ID.String())

	case mountActionDoneMsg:
		logger.TUIOut(string(msg))
		return m, refreshMountStateCmd(m.targetMount.ID.String())

	case refreshDbBothMsg:
		m.dbHostInfoList.Title = msg.info.title
		m.dbHostInfoList.SetItems(msg.info.items)
		m.targetDbHost = msg.admin.dbHost
		m.dbHostAdminList.Title = "Manage " + msg.admin.dbHost.Name
		m.dbHostAdminList.SetItems(msg.admin.items)
		return m, nil

	case refreshMountBothMsg:
		m.mountInfoList.Title = msg.info.title
		m.mountInfoList.SetItems(msg.info.items)
		m.targetMount = msg.admin.mount
		m.mountAdminList.Title = "Manage " + msg.admin.mount.Name
		m.mountAdminList.SetItems(msg.admin.items)
		return m, nil

	case refreshServerBothMsg:
		m.serverInfoList.Title = msg.info.title
		m.serverInfoList.SetItems(msg.info.items)
		m.targetServer = msg.admin.server
		m.serverAdminList.Title = "Manage " + msg.admin.server.Name
		m.serverAdminList.SetItems(msg.admin.items)
		m.serverLimitsList.Title = "Resource Limits for " + msg.limits.server.Name
		m.serverLimitsList.SetItems(msg.limits.items)
		m.serverList.SetItems(msg.servers)
		return m, nil

	case refreshBothMsg:
		m.userInfoList.Title = msg.info.title
		m.userInfoList.SetItems(msg.info.items)
		m.targetUser = msg.admin.user
		m.adminActionsList.Title = "Manage " + msg.admin.user.Username
		m.adminActionsList.SetItems(msg.admin.items)
		m.adminLimitsList.Title = "Resource Limits for " + msg.limits.user.Username
		m.adminLimitsList.SetItems(msg.limits.items)
		m.userList.SetItems(msg.users)
		return m, nil

	case askConfirmMsg:
		m.pendingConfirmDesc = msg.desc
		m.pendingConfirmCmd = msg.cmd
		m.previousAdminState = m.state
		m.state = viewAdminConfirm

	case askPromptMsg:
		m.promptDesc = msg.desc
		m.promptField = msg.field
		m.textInput.Placeholder = msg.placeholder
		m.previousAdminState = m.state
		m.state = viewAdminPrompt
		return m, nil

	case askServerPromptMsg:
		m.promptDesc = msg.desc
		m.promptField = msg.field
		m.textInput.Placeholder = msg.placeholder
		m.previousAdminState = m.state
		m.state = viewServerPrompt
		return m, nil

	case askDbHostPromptMsg:
		m.promptDesc = msg.desc
		m.promptField = msg.field
		m.textInput.Placeholder = msg.placeholder
		m.previousAdminState = m.state
		m.state = viewDbHostPrompt
		return m, nil

	case askMountPromptMsg:
		m.promptDesc = msg.desc
		m.promptField = msg.field
		m.textInput.Placeholder = msg.placeholder
		m.previousAdminState = m.state
		m.state = viewMountPrompt
		return m, nil

	case showUserListMsg:
		m.userList.SetItems(msg)
		m.state = viewUserList
		return m, nil

	case showUserInfoMsg:
		m.userInfoList.Title = msg.title
		m.userInfoList.SetItems(msg.items)
		m.state = viewUserInfo
		return m, nil

	case showAdminActionsMsg:
		m.targetUser = msg.user
		m.adminActionsList.Title = "Manage " + msg.user.Username
		m.adminActionsList.SetItems(msg.items)
		m.state = viewAdminActions
		return m, nil

	case showAdminLimitsMsg:
		m.targetUser = msg.user
		m.adminLimitsList.Title = "Resource Limits for " + msg.user.Username
		m.adminLimitsList.SetItems(msg.items)
		m.state = viewAdminLimits
		return m, nil

	case showServerListMsg:
		m.serverList.SetItems(msg)
		m.state = viewServerList
		return m, nil

	case showServerInfoMsg:
		m.serverInfoList.Title = msg.title
		m.serverInfoList.SetItems(msg.items)
		m.state = viewServerInfo
		return m, nil

	case showServerAdminActionsMsg:
		m.targetServer = msg.server
		m.serverAdminList.Title = "Manage " + msg.server.Name
		m.serverAdminList.SetItems(msg.items)
		m.state = viewServerAdminActions
		return m, nil

	case showServerLimitsMsg:
		m.targetServer = msg.server
		m.serverLimitsList.Title = "Resource Limits for " + msg.server.Name
		m.serverLimitsList.SetItems(msg.items)
		m.state = viewServerLimits
		return m, nil

	case showServerCreateMsg:
		m.state = viewServerCreate
		m.createServerFocus = 0
		for i := range m.createServerInputs {
			m.createServerInputs[i].SetValue("")
			if i == 0 {
				m.createServerInputs[i].Focus()
				m.createServerInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				m.createServerInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			} else {
				m.createServerInputs[i].Blur()
				m.createServerInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				m.createServerInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}
		}
		return m, nil

	case showNodeListMsg:
		m.nodeList.SetItems(msg)
		m.state = viewNodeList
		return m, nil

	case showNodeInfoMsg:
		m.nodeInfoList.Title = msg.title
		m.nodeInfoList.SetItems(msg.items)
		m.state = viewNodeInfo
		return m, nil

	case showNodeAdminActionsMsg:
		m.nodeAdminList.Title = "Manage " + msg.node.Name
		m.nodeAdminList.SetItems(msg.items)
		m.state = viewNodeAdminActions
		return m, nil

	case showNodeCreateMsg:
		m.state = viewNodeCreate
		m.createNodeFocus = 0
		for i := range m.createNodeInputs {
			m.createNodeInputs[i].SetValue("")
			if i == 0 {
				m.createNodeInputs[i].Focus()
				m.createNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				m.createNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			} else {
				m.createNodeInputs[i].Blur()
				m.createNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				m.createNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}
		}
		return m, nil

	case showNodePairMsg:
		m.state = viewNodePair
		m.pairNodeFocus = 0
		for i := range m.pairNodeInputs {
			m.pairNodeInputs[i].SetValue("")
			if i == 0 {
				m.pairNodeInputs[i].Focus()
				m.pairNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				m.pairNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			} else {
				m.pairNodeInputs[i].Blur()
				m.pairNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				m.pairNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}
		}
		return m, nil

	case askNodePairCodeMsg:
		m.state = viewNodePairWait
		m.pairNodeCode = msg.code
		return m, executePairNodeFinalize(msg.vals, msg.code)

	case showDatabaseHostListMsg:
		m.dbHostList.SetItems(msg)
		m.state = viewDatabaseHostList
		return m, nil

	case showDatabaseHostInfoMsg:
		m.dbHostInfoList.Title = msg.title
		m.dbHostInfoList.SetItems(msg.items)
		m.state = viewDatabaseHostInfo
		return m, nil

	case showDatabaseHostAdminActionsMsg:
		m.targetDbHost = msg.dbHost
		m.dbHostAdminList.Title = "Manage " + msg.dbHost.Name
		m.dbHostAdminList.SetItems(msg.items)
		m.state = viewDatabaseHostAdminActions
		return m, nil

	case showDatabaseHostCreateMsg:
		m.state = viewDatabaseHostCreate
		m.createDbHostFocus = 0
		for i := range m.createDbHostInputs {
			m.createDbHostInputs[i].SetValue("")
			if i == 0 {
				m.createDbHostInputs[i].Focus()
				m.createDbHostInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				m.createDbHostInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			} else {
				m.createDbHostInputs[i].Blur()
				m.createDbHostInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				m.createDbHostInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}
		}
		return m, nil

	case showMountListMsg:
		m.mountList.SetItems(msg)
		m.state = viewMountList
		return m, nil

	case showMountInfoMsg:
		m.mountInfoList.Title = msg.title
		m.mountInfoList.SetItems(msg.items)
		m.state = viewMountInfo
		return m, nil

	case showMountAdminActionsMsg:
		m.targetMount = msg.mount
		m.mountAdminList.Title = "Manage " + msg.mount.Name
		m.mountAdminList.SetItems(msg.items)
		m.state = viewMountAdminActions
		return m, nil

	case showMountCreateMsg:
		m.state = viewMountCreate
		m.createMountFocus = 0
		for i := range m.createMountInputs {
			m.createMountInputs[i].SetValue("")
			if i == 0 {
				m.createMountInputs[i].Focus()
				m.createMountInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				m.createMountInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			} else {
				m.createMountInputs[i].Blur()
				m.createMountInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				m.createMountInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			}
		}
		return m, nil

	case showUserActivityMsg:
		m.isGlobalActivity = msg.isGlobal
		m.userActivityList.SetItems(msg.items)
		m.state = viewUserActivity
		return m, nil

	case showActivityDetailsMsg:
		m.activityDetailsList.Title = msg.title
		m.activityDetailsList.SetItems(msg.items)
		m.state = viewActivityDetails
		return m, nil

	case logMsg:
		m.logs = append(m.logs, string(msg))
		if m.ready {
			m.viewport.SetContent(strings.Join(m.logs, "\n"))
			m.viewport.GotoBottom()
		}
		cmds = append(cmds, waitForLog())
	}

	m, cmd = updateUsers(msg, m)
	cmds = append(cmds, cmd)

	m, cmd = updateServers(msg, m)
	cmds = append(cmds, cmd)

	m, cmd = updateNodes(msg, m)
	cmds = append(cmds, cmd)

	m, cmd = updateDatabaseHosts(msg, m)
	cmds = append(cmds, cmd)

	m, cmd = updateMounts(msg, m)
	cmds = append(cmds, cmd)

	m, cmd = updateActivity(msg, m)
	cmds = append(cmds, cmd)

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

var (
	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	cmdErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			MarginTop(1).
			MarginBottom(1)
)

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	if m.state == viewUserList || m.state == viewUserInfo || m.state == viewAdminActions || m.state == viewAdminLimits || m.state == viewAdminConfirm || m.state == viewAdminPrompt {
		return viewUsers(m)
	} else if m.state == viewServerList || m.state == viewServerInfo || m.state == viewServerAdminActions || m.state == viewServerLimits || m.state == viewServerCreate || m.state == viewServerPrompt {
		return viewServers(m)
	} else if m.state == viewNodeList || m.state == viewNodeInfo || m.state == viewNodeAdminActions || m.state == viewNodeCreate || m.state == viewNodePair || m.state == viewNodePairWait {
		return viewNodes(m)
	} else if m.state == viewDatabaseHostList || m.state == viewDatabaseHostInfo || m.state == viewDatabaseHostAdminActions || m.state == viewDatabaseHostCreate || m.state == viewDbHostPrompt {
		return viewDatabaseHosts(m)
	} else if m.state == viewMountList || m.state == viewMountInfo || m.state == viewMountAdminActions || m.state == viewMountCreate || m.state == viewMountPrompt {
		return viewMounts(m)
	} else if m.state == viewUserActivity || m.state == viewActivityDetails {
		return viewActivity(m)
	}

	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textInput.View(),
	)
}

func (m *model) updateServerCreateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.createServerInputs))
	for i := range m.createServerInputs {
		m.createServerInputs[i], cmds[i] = m.createServerInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func Start() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting TUI: %v", err)
		os.Exit(1)
	}
}
