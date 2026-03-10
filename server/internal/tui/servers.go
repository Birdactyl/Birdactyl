package tui

import (
	"fmt"
	"strconv"
	"strings"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/logger"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type serverCreateItem struct{}

func (i serverCreateItem) Title() string       { return "Create New Server" }
func (i serverCreateItem) Description() string { return "Deploy a new server on a node" }
func (i serverCreateItem) FilterValue() string { return "create new server" }

func fetchServersCmd() tea.Cmd {
	return func() tea.Msg {
		var servers []models.Server
		database.DB.Preload("User").Preload("Node").Preload("Package").Find(&servers)
		items := make([]list.Item, len(servers)+1)
		items[0] = serverCreateItem{}
		for i, s := range servers {
			items[i+1] = serverItem{s}
		}
		return showServerListMsg(items)
	}
}

func fetchUserServersCmd(userID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		var servers []models.Server
		database.DB.Preload("User").Preload("Node").Preload("Package").Where("user_id = ?", userID).Find(&servers)
		items := make([]list.Item, len(servers)+1)
		items[0] = serverCreateItem{}
		for i, s := range servers {
			items[i+1] = serverItem{s}
		}
		return showServerListMsg(items)
	}
}

func fetchServerInfoCmd(query string) tea.Cmd {
	return func() tea.Msg {
		var server models.Server
		qc := "name = ?"
		args := []interface{}{query}
		if _, err := uuid.Parse(query); err == nil {
			qc += " OR id = ?"
			args = append(args, query)
		}

		if err := database.DB.Preload("Node").Preload("User").Preload("Package").Where(qc, args...).First(&server).Error; err != nil {
			return errorMessageMsg(cmdErrorStyle.Render("Server not found: " + query))
		}

		ownerDesc := "Unknown"
		if server.User != nil {
			ownerDesc = server.User.Username + " (" + server.User.Email + ")"
		}

		nodeName := "Unknown"
		if server.Node != nil {
			nodeName = server.Node.Name
		}

		packageName := "Unknown"
		if server.Package != nil {
			packageName = server.Package.Name
		}

		items := []list.Item{
			infoItem{
				"Admin Actions",
				"Suspend, unsuspend, or delete this server.",
				fetchServerAdminActionsCmd(server),
			},
			infoItem{"ID", server.ID.String(), nil},
			infoItem{"Name", server.Name, nil},
			infoItem{"Status", string(server.Status), nil},
			infoItem{"Owner", ownerDesc, nil},
			infoItem{"Node", nodeName, nil},
			infoItem{"Package", packageName, nil},
			infoItem{"Suspended", fmt.Sprintf("%v", server.IsSuspended), nil},
			infoItem{"Memory Limit", fmt.Sprintf("%d MB", server.Memory), nil},
			infoItem{"CPU Limit", fmt.Sprintf("%d %%", server.CPU), nil},
			infoItem{"Disk Limit", fmt.Sprintf("%d MB", server.Disk), nil},
			infoItem{"Created At", server.CreatedAt.Format("2006-01-02 15:04:05"), nil},
			infoItem{"Updated At", server.UpdatedAt.Format("2006-01-02 15:04:05"), nil},
		}

		return showServerInfoMsg{
			title: "Server Info: " + server.Name,
			items: items,
		}
	}
}

func fetchServerAdminActionsCmd(server models.Server) tea.Cmd {
	return func() tea.Msg {
		suspendTopic, suspendDesc, suspendType := "Suspend Server", "Stop and lock this server", "suspend"
		if server.IsSuspended {
			suspendTopic, suspendDesc, suspendType = "Unsuspend Server", "Unlock and allow this server to start", "unsuspend"
		}

		ownerName := "Unknown"
		if server.User != nil {
			ownerName = server.User.Username
		}

		items := []list.Item{
			infoItem{"Start Server", "Send start command to the daemon", confirmServerAdminExecCmd("start", "Start server "+server.Name+"?", server)},
			infoItem{"Stop Server", "Send stop command to the daemon", confirmServerAdminExecCmd("stop", "Stop server "+server.Name+"?", server)},
			infoItem{"Restart Server", "Send restart command to the daemon", confirmServerAdminExecCmd("restart", "Restart server "+server.Name+"?", server)},
			infoItem{"Kill Server", "Forcefully terminate daemon process", confirmServerAdminExecCmd("kill", "Force kill server "+server.Name+"?", server)},
			infoItem{suspendTopic, suspendDesc, confirmServerAdminExecCmd(suspendType, "Are you sure you want to "+strings.ToLower(suspendTopic)+" "+server.Name+"?", server)},
			infoItem{"Delete Server", "Permanently delete server and its data", confirmServerAdminExecCmd("delete", "WARNING: Permanently delete server "+server.Name+" and all its data?", server)},
			infoItem{"Edit Boundaries / Limits", "Modify maximum RAM, CPU, and Disk", func() tea.Msg { return fetchServerLimitsCmd(server)() }},
			infoItem{"Edit Server Name", "Change internal display name", promptServerAdminEditCmd("Enter new server name", "name", server.Name)},
			infoItem{"Edit Owner", "Transfer server to new user (UUID or Username)", promptServerAdminEditCmd("Enter new owner UUID or Username", "user_id", ownerName)},
		}

		return showServerAdminActionsMsg{
			server: server,
			items:  items,
		}
	}
}

func confirmServerAdminExecCmd(action string, desc string, server models.Server) tea.Cmd {
	return func() tea.Msg {
		return askConfirmMsg{
			desc: desc,
			cmd: func() tea.Msg {
				return executeDirectServerAdminAction(action, server)
			},
		}
	}
}

func enterServerCreateCmd() tea.Cmd {
	return func() tea.Msg {
		return showServerCreateMsg{}
	}
}

func executeCreateServerDirectly(vals []string) tea.Cmd {
	return func() tea.Msg {
		if len(vals) < 7 {
			return errorMessageMsg("Form incomplete")
		}

		name := strings.TrimSpace(vals[0])
		if name == "" {
			return errorMessageMsg("Server name is required")
		}

		var user models.User
		if uid, err := uuid.Parse(vals[1]); err == nil {
			if database.DB.Where("id = ?", uid).First(&user).Error != nil {
				return errorMessageMsg("Owner UUID not found")
			}
		} else {
			if database.DB.Where("username = ?", vals[1]).First(&user).Error != nil {
				return errorMessageMsg("Owner username not found")
			}
		}

		var node models.Node
		if nid, err := uuid.Parse(vals[2]); err == nil {
			if database.DB.Where("id = ?", nid).First(&node).Error != nil {
				return errorMessageMsg("Node UUID not found")
			}
		} else {
			if database.DB.Where("name = ?", vals[2]).First(&node).Error != nil {
				return errorMessageMsg("Node name not found")
			}
		}

		var pkg models.Package
		if pid, err := uuid.Parse(vals[3]); err == nil {
			if database.DB.Where("id = ?", pid).First(&pkg).Error != nil {
				return errorMessageMsg("Package UUID not found")
			}
		} else {
			if database.DB.Where("name = ?", vals[3]).First(&pkg).Error != nil {
				return errorMessageMsg("Package name not found")
			}
		}

		mem, err := strconv.Atoi(vals[4])
		if err != nil {
			return errorMessageMsg("Memory must be a number")
		}
		cpu, err := strconv.Atoi(vals[5])
		if err != nil {
			return errorMessageMsg("CPU must be a number")
		}
		disk, err := strconv.Atoi(vals[6])
		if err != nil {
			return errorMessageMsg("Disk must be a number")
		}

		req := services.CreateServerRequest{
			Name:      name,
			NodeID:    node.ID,
			PackageID: pkg.ID,
			Memory:    mem,
			CPU:       cpu,
			Disk:      disk,
			Ports:     []models.ServerPort{{Port: 0, Primary: true}},
		}

		serv, err := services.CreateServer(user.ID, req)
		if err != nil {
			return errorMessageMsg("Failed to create server: " + err.Error())
		}

		go services.SendCreateServer(serv)

		return fetchServerInfoCmd(serv.ID.String())()
	}
}

func promptServerAdminEditCmd(desc string, field string, placeholder string) tea.Cmd {
	return func() tea.Msg {
		return askServerPromptMsg{desc: desc, field: field, placeholder: placeholder}
	}
}

func fetchServerLimitsCmd(server models.Server) tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			infoItem{"Edit Limits [RAM]", "Format: 1024 (MB) or 0 (unlimited)", promptServerAdminEditCmd("Edit RAM Limit (0 = Unlimited, MB)", "memory", strconv.Itoa(server.Memory))},
			infoItem{"Edit Limits [CPU]", "Format: 200 (%) or 0 (unlimited)", promptServerAdminEditCmd("Edit CPU Limit (0 = Unlimited, %)", "cpu", strconv.Itoa(server.CPU))},
			infoItem{"Edit Limits [Disk]", "Format: 5000 (MB) or 0 (unlimited)", promptServerAdminEditCmd("Edit Disk Limit (0 = Unlimited, MB)", "disk", strconv.Itoa(server.Disk))},
		}

		return showServerLimitsMsg{
			server: server,
			items:  items,
		}
	}
}

func executeDirectServerAdminAction(action string, server models.Server) tea.Msg {
	currentUser := models.User{Username: "SYSTEM (TUI)", ID: uuid.Nil}
	switch action {
	case "start":
		if err := services.SendStartServer(server.ID); err != nil {
			return serverActionDoneMsg("Failed to start: " + err.Error())
		}
		return serverActionDoneMsg("Started " + server.Name)
	case "stop":
		if err := services.SendStopServer(server.ID); err != nil {
			return serverActionDoneMsg("Failed to stop: " + err.Error())
		}
		return serverActionDoneMsg("Stopped " + server.Name)
	case "restart":
		if err := services.SendRestartServer(server.ID); err != nil {
			return serverActionDoneMsg("Failed to restart: " + err.Error())
		}
		return serverActionDoneMsg("Restarted " + server.Name)
	case "kill":
		if err := services.SendKillServer(server.ID); err != nil {
			return serverActionDoneMsg("Failed to kill: " + err.Error())
		}
		return serverActionDoneMsg("Killed " + server.Name)
	case "suspend":
		services.SendKillServer(server.ID)
		services.UpdateServerStatus(server.ID, models.ServerStatusStopped, "")
		services.SuspendServer(server.ID)
		return serverActionDoneMsg("Suspended " + server.Name)
	case "unsuspend":
		services.UnsuspendServer(server.ID)
		return serverActionDoneMsg("Unsuspended " + server.Name)
	case "delete":
		services.SendDeleteServer(server.ID)
		services.DeleteServerAdmin(server.ID)
		return errorMessageMsg("Deleted server " + server.Name + ". Returning to logs.")
	}
	_ = currentUser
	return nil
}

func executeServerPromptConfig(server models.Server, field string, val string) tea.Cmd {
	return func() tea.Msg {
		val = strings.TrimSpace(val)
		if val == "" {
			return serverActionDoneMsg("Edit cancelled: No value provided.")
		}

		updates := map[string]interface{}{}

		switch field {
		case "name":
			updates["name"] = val
		case "user_id":
			var user models.User
			var found bool
			if uid, err := uuid.Parse(val); err == nil {
				if database.DB.Where("id = ?", uid).First(&user).Error == nil {
					found = true
				}
			}
			if !found {
				if database.DB.Where("username = ?", val).First(&user).Error == nil {
					found = true
				}
			}
			if !found {
				return serverActionDoneMsg("Edit failed: User not found.")
			}
			updates["user_id"] = user.ID
		case "memory", "cpu", "disk":
			i, err := strconv.Atoi(val)
			if err != nil {
				return serverActionDoneMsg("Edit failed: Value must be an integer.")
			}
			if i < 0 {
				updates[field] = 0
			} else {
				updates[field] = i
			}
		}

		if err := database.DB.Model(&server).Updates(updates).Error; err != nil {
			return serverActionDoneMsg("Edit failed: DB error " + err.Error())
		}

		return serverActionDoneMsg("Updated " + field + " for " + server.Name)
	}
}

func refreshServerAdminStateCmd(sid string) tea.Cmd {
	return func() tea.Msg {
		infoMsg := fetchServerInfoCmd(sid)()
		if errInfo, ok := infoMsg.(errorMessageMsg); ok {
			return errInfo
		}

		var server models.Server
		if err := database.DB.Preload("User").Preload("Node").Preload("Package").Where("id = ?", sid).First(&server).Error; err != nil {
			return errorMessageMsg("Server not found: " + sid)
		}

		adminMsg := fetchServerAdminActionsCmd(server)()
		limitsMsg := fetchServerLimitsCmd(server)()

		var allServers []models.Server
		database.DB.Find(&allServers)
		serverItems := make([]list.Item, len(allServers))
		for i, s := range allServers {
			serverItems[i] = serverItem{s}
		}

		return refreshServerBothMsg{
			info:    infoMsg.(showServerInfoMsg),
			admin:   adminMsg.(showServerAdminActionsMsg),
			limits:  limitsMsg.(showServerLimitsMsg),
			servers: serverItems,
		}
	}
}

func updateServers(msg tea.Msg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case viewServerList:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewLogs
				return m, nil
			case "enter":
				if i, ok := m.serverList.SelectedItem().(serverItem); ok {
					return m, fetchServerInfoCmd(i.ID.String())
				} else if _, ok := m.serverList.SelectedItem().(serverCreateItem); ok {
					return m, enterServerCreateCmd()
				}
			}
			m.serverList, cmd = m.serverList.Update(msg)
			return m, cmd
		case viewServerInfo:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewServerList
				return m, nil
			case "enter":
				if i, ok := m.serverInfoList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.serverInfoList, cmd = m.serverInfoList.Update(msg)
			return m, cmd
		case viewServerAdminActions:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewServerInfo
				return m, nil
			case "enter":
				if i, ok := m.serverAdminList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.serverAdminList, cmd = m.serverAdminList.Update(msg)
			return m, cmd
		case viewServerLimits:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewServerAdminActions
				return m, nil
			case "enter":
				if i, ok := m.serverLimitsList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.serverLimitsList, cmd = m.serverLimitsList.Update(msg)
			return m, cmd
		case viewServerPrompt:
			switch msg.String() {
			case "esc", "ctrl+c":
				m.state = m.previousAdminState
				m.textInput.SetValue("")
				m.textInput.Placeholder = ""
				return m, nil
			case "enter":
				val := m.textInput.Value()
				m.textInput.SetValue("")
				m.textInput.Placeholder = ""
				m.state = m.previousAdminState
				return m, executeServerPromptConfig(m.targetServer, m.promptField, val)
			}
		case viewServerCreate:
			switch msg.String() {
			case "esc", "ctrl+c":
				m.state = viewServerList
				return m, nil
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" {
					m.createServerFocus--
				} else {
					m.createServerFocus++
				}
				if m.createServerFocus > len(m.createServerInputs) {
					m.createServerFocus = 0
				} else if m.createServerFocus < 0 {
					m.createServerFocus = len(m.createServerInputs)
				}
				
				cmds := make([]tea.Cmd, len(m.createServerInputs))
				for i := 0; i <= len(m.createServerInputs)-1; i++ {
					if i == m.createServerFocus {
						cmds[i] = m.createServerInputs[i].Focus()
						m.createServerInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						m.createServerInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					} else {
						m.createServerInputs[i].Blur()
						m.createServerInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						m.createServerInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					}
				}
				return m, tea.Batch(cmds...)
			case "enter":
				if m.createServerFocus == len(m.createServerInputs) {
					var vals []string
					for _, t := range m.createServerInputs {
						vals = append(vals, t.Value())
					}
					m.state = viewServerList
					return m, executeCreateServerDirectly(vals)
				} else {
					m.createServerFocus++
					cmds := make([]tea.Cmd, len(m.createServerInputs))
					for i := 0; i < len(m.createServerInputs); i++ {
						if i == m.createServerFocus {
							cmds[i] = m.createServerInputs[i].Focus()
							m.createServerInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
							m.createServerInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						} else {
							m.createServerInputs[i].Blur()
							m.createServerInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
							m.createServerInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						}
					}
					return m, tea.Batch(cmds...)
				}
			}
			cmds := make([]tea.Cmd, len(m.createServerInputs))
			for i := range m.createServerInputs {
				m.createServerInputs[i], cmds[i] = m.createServerInputs[i].Update(msg)
			}
			return m, tea.Batch(cmds...)
		}
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
	case serverActionDoneMsg:
		logger.TUIOut(string(msg))
		return m, refreshServerAdminStateCmd(m.targetServer.ID.String())

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
	}
	return m, nil
}

func viewServers(m model) string {
	switch m.state {
	case viewServerList:
		return m.serverList.View()
	case viewServerInfo:
		return m.serverInfoList.View()
	case viewServerAdminActions:
		return m.serverAdminList.View()
	case viewServerLimits:
		return m.serverLimitsList.View()
	case viewServerCreate:
		var b strings.Builder
		for i := range m.createServerInputs {
			b.WriteString(m.createServerInputs[i].View())
			if i < len(m.createServerInputs)-1 {
				b.WriteRune('\n')
			}
		}
		
		btn := "[ Submit ]"
		if m.createServerFocus == len(m.createServerInputs) {
			btn = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(btn)
		} else {
			btn = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(btn)
		}
		
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Create New Server") + "\n\n" + b.String() + "\n\n" + btn + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Use Tab/Shift+Tab to navigate, Enter to submit, Esc to cancel.")
		return "\n" + box.Render(body)
	case viewServerPrompt:
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Input Required") + "\n\n" + m.promptDesc + "\n\n" + m.textInput.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to submit, Esc to cancel.")
		return "\n" + box.Render(body)
	}
	return ""
}
