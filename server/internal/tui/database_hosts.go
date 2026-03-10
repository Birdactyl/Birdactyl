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

type dbHostItem struct {
	models.DatabaseHost
	Count int64
}

func (i dbHostItem) Title() string {
	return fmt.Sprintf("%s (%s:%d)", i.Name, i.Host, i.Port)
}
func (i dbHostItem) Description() string {
	return fmt.Sprintf("ID: %s | Databases: %d/%d", i.ID.String(), i.Count, i.MaxDatabases)
}
func (i dbHostItem) FilterValue() string {
	return fmt.Sprintf("%s %s %d %s", i.Name, i.Host, i.Port, i.ID.String())
}

type dbHostCreateItem struct{}

func (i dbHostCreateItem) Title() string       { return "Create New Database Host" }
func (i dbHostCreateItem) Description() string { return "Link a MySQL database server for clients" }
func (i dbHostCreateItem) FilterValue() string { return "create new database host dbhost" }

func fetchDatabaseHostsCmd() tea.Cmd {
	return func() tea.Msg {
		var hosts []models.DatabaseHost
		database.DB.Find(&hosts)
		
		items := make([]list.Item, len(hosts)+1)
		items[0] = dbHostCreateItem{}
		
		for i, h := range hosts {
			var count int64
			database.DB.Model(&models.ServerDatabase{}).Where("host_id = ?", h.ID).Count(&count)
			items[i+1] = dbHostItem{h, count}
		}
		
		return showDatabaseHostListMsg(items)
	}
}

func fetchDatabaseHostInfoCmd(query string) tea.Cmd {
	return func() tea.Msg {
		var host models.DatabaseHost
		qc := "id = ?"
		if _, err := uuid.Parse(query); err != nil {
			qc = "name = ? OR host = ?"
		}

		if err := database.DB.Where(qc, query, query).First(&host).Error; err != nil {
			return errorMessageMsg("Database Host not found: " + query)
		}

		var count int64
		database.DB.Model(&models.ServerDatabase{}).Where("host_id = ?", host.ID).Count(&count)

		items := []list.Item{
			infoItem{
				"Admin Actions",
				"Edit settings or delete this host.",
				fetchDatabaseHostAdminActionsCmd(host),
			},
			infoItem{"ID", host.ID.String(), nil},
			infoItem{"Name", host.Name, nil},
			infoItem{"Host", host.Host, nil},
			infoItem{"Port", strconv.Itoa(host.Port), nil},
			infoItem{"Username", host.Username, nil},
			infoItem{"Max Databases", strconv.Itoa(host.MaxDatabases), nil},
			infoItem{"Active Databases", fmt.Sprintf("%d", count), nil},
			infoItem{"Created At", host.CreatedAt.Format("2006-01-02 15:04:05"), nil},
		}

		return showDatabaseHostInfoMsg{
			title: "Database Host Info: " + host.Name,
			items: items,
		}
	}
}

func fetchDatabaseHostAdminActionsCmd(host models.DatabaseHost) tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			infoItem{"Delete DB Host", "Permanently remove this database host", confirmDbHostAdminExecCmd("delete_dbhost", "WARNING: Permanently delete database host "+host.Name+"?", host)},
			infoItem{"Edit Name", "Change host display name", promptDbHostPromptCmd("Enter new name", "dbhost_name", host.Name)},
			infoItem{"Edit Host IP/FQDN", "Change the connection address", promptDbHostPromptCmd("Enter new hostname", "dbhost_host", host.Host)},
			infoItem{"Edit Port", "Change the MySQL port", promptDbHostPromptCmd("Enter new port (e.g. 3306)", "dbhost_port", strconv.Itoa(host.Port))},
			infoItem{"Edit Username", "Change root MySQL user", promptDbHostPromptCmd("Enter new username", "dbhost_username", host.Username)},
			infoItem{"Edit Password", "Re-enter root MySQL password", promptDbHostPromptCmd("Enter new password", "dbhost_password", "")},
			infoItem{"Edit Max Databases", "Maximum number of databases (0 = unlimited)", promptDbHostPromptCmd("Enter max limit (0 = unlimited)", "dbhost_max", strconv.Itoa(host.MaxDatabases))},
		}

		return showDatabaseHostAdminActionsMsg{
			dbHost: host,
			items:  items,
		}
	}
}

func confirmDbHostAdminExecCmd(action string, desc string, host models.DatabaseHost) tea.Cmd {
	return func() tea.Msg {
		return askConfirmMsg{
			desc: desc,
			cmd: func() tea.Msg {
				return executeDirectDBHostAction(host)
			},
		}
	}
}

func promptDbHostPromptCmd(desc string, field string, placeholder string) tea.Cmd {
	return func() tea.Msg {
		return askDbHostPromptMsg{desc: desc, field: field, placeholder: placeholder}
	}
}

func executeDirectDBHostAction(host models.DatabaseHost) tea.Msg {
	var count int64
	database.DB.Model(&models.ServerDatabase{}).Where("host_id = ?", host.ID).Count(&count)

	if count > 0 {
		return errorMessageMsg("Cannot delete database host because it has active databases linked.")
	}

	database.DB.Where("id = ?", host.ID).Delete(&models.DatabaseHost{})
	return dbHostActionDoneMsg("Deleted database host " + host.Name + ". Returning to logs.")
}

func executeDBHostPromptConfig(host models.DatabaseHost, field string, val string) tea.Cmd {
	return func() tea.Msg {
		val = strings.TrimSpace(val)
		if val == "" {
			return dbHostActionDoneMsg("Edit cancelled: No value provided.")
		}

		updates := map[string]interface{}{}
		
		switch field {
		case "dbhost_name":
			updates["name"] = val
		case "dbhost_host":
			updates["host"] = val
		case "dbhost_port":
			p, err := strconv.Atoi(val)
			if err != nil { return dbHostActionDoneMsg("Edit failed: Port must be a number") }
			updates["port"] = p
		case "dbhost_username":
			updates["username"] = val
		case "dbhost_password":
			updates["password"] = val
		case "dbhost_max":
			max, err := strconv.Atoi(val)
			if err != nil { return dbHostActionDoneMsg("Edit failed: Limit must be a number") }
			updates["max_databases"] = max
		}

		if err := database.DB.Model(&host).Updates(updates).Error; err != nil {
			return dbHostActionDoneMsg("Edit failed: DB error " + err.Error())
		}
		
		return dbHostActionDoneMsg("Updated " + field + " for " + host.Name)
	}
}

func refreshDbHostStateCmd(id string) tea.Cmd {
	return func() tea.Msg {
		infoMsg := fetchDatabaseHostInfoCmd(id)()
		if errInfo, ok := infoMsg.(errorMessageMsg); ok {
			return errInfo
		}

		var host models.DatabaseHost
		database.DB.Where("id = ?", id).First(&host)

		adminMsg := fetchDatabaseHostAdminActionsCmd(host)()

		return refreshDbBothMsg{
			info:  infoMsg.(showDatabaseHostInfoMsg),
			admin: adminMsg.(showDatabaseHostAdminActionsMsg),
		}
	}
}

func enterDatabaseHostCreateCmd() tea.Cmd {
	return func() tea.Msg {
		return showDatabaseHostCreateMsg{}
	}
}

func executeCreateDatabaseHostDirectly(vals []string) tea.Cmd {
	return func() tea.Msg {
		name := strings.TrimSpace(vals[0])
		host := strings.TrimSpace(vals[1])
		portStr := strings.TrimSpace(vals[2])
		user := strings.TrimSpace(vals[3])
		pass := strings.TrimSpace(vals[4])
		maxStr := strings.TrimSpace(vals[5])

		if name == "" || host == "" || user == "" || pass == "" {
			return errorMessageMsg("Name, Host, Username, and Password are required")
		}

		port := 3306
		if portStr != "" {
			var err error
			port, err = strconv.Atoi(portStr)
			if err != nil { return errorMessageMsg("Port must be a valid number") }
		}

		max := 0
		if maxStr != "" {
			var err error
			max, err = strconv.Atoi(maxStr)
			if err != nil { return errorMessageMsg("Max Databases must be a valid number") }
		}

		_, err := services.CreateDatabaseHost(name, host, port, user, pass, max)
		if err != nil {
			return errorMessageMsg("Failed to create database host: " + err.Error())
		}

		return fetchDatabaseHostsCmd()()
	}
}

func updateDatabaseHosts(msg tea.Msg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case viewDatabaseHostList:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewLogs
				return m, nil
			case "enter":
				if i, ok := m.dbHostList.SelectedItem().(dbHostItem); ok {
					return m, fetchDatabaseHostInfoCmd(i.ID.String())
				} else if _, ok := m.dbHostList.SelectedItem().(dbHostCreateItem); ok {
					return m, enterDatabaseHostCreateCmd()
				}
			}
			m.dbHostList, cmd = m.dbHostList.Update(msg)
			return m, cmd
		case viewDatabaseHostInfo:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewDatabaseHostList
				return m, nil
			case "enter":
				if i, ok := m.dbHostInfoList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.dbHostInfoList, cmd = m.dbHostInfoList.Update(msg)
			return m, cmd
		case viewDatabaseHostAdminActions:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewDatabaseHostInfo
				return m, nil
			case "enter":
				if i, ok := m.dbHostAdminList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.dbHostAdminList, cmd = m.dbHostAdminList.Update(msg)
			return m, cmd
		case viewDatabaseHostCreate:
			switch msg.String() {
			case "esc", "ctrl+c":
				m.state = viewDatabaseHostList
				return m, nil
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" { m.createDbHostFocus-- } else { m.createDbHostFocus++ }
				if m.createDbHostFocus > len(m.createDbHostInputs) { m.createDbHostFocus = 0 } else if m.createDbHostFocus < 0 { m.createDbHostFocus = len(m.createDbHostInputs) }
				cmds := make([]tea.Cmd, len(m.createDbHostInputs))
				for i := 0; i <= len(m.createDbHostInputs)-1; i++ {
					if i == m.createDbHostFocus {
						cmds[i] = m.createDbHostInputs[i].Focus()
						m.createDbHostInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						m.createDbHostInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					} else {
						m.createDbHostInputs[i].Blur()
						m.createDbHostInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						m.createDbHostInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					}
				}
				return m, tea.Batch(cmds...)
			case "enter":
				if m.createDbHostFocus == len(m.createDbHostInputs) {
					var vals []string
					for _, t := range m.createDbHostInputs { vals = append(vals, t.Value()) }
					m.state = viewDatabaseHostList
					return m, executeCreateDatabaseHostDirectly(vals)
				} else {
					m.createDbHostFocus++
					cmds := make([]tea.Cmd, len(m.createDbHostInputs))
					for i := 0; i < len(m.createDbHostInputs); i++ {
						if i == m.createDbHostFocus {
							cmds[i] = m.createDbHostInputs[i].Focus()
							m.createDbHostInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
							m.createDbHostInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						} else {
							m.createDbHostInputs[i].Blur()
							m.createDbHostInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
							m.createDbHostInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						}
					}
					return m, tea.Batch(cmds...)
				}
			}
			cmds := make([]tea.Cmd, len(m.createDbHostInputs))
			for i := range m.createDbHostInputs { m.createDbHostInputs[i], cmds[i] = m.createDbHostInputs[i].Update(msg) }
			return m, tea.Batch(cmds...)
		case viewDbHostPrompt:
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
				return m, executeDBHostPromptConfig(m.targetDbHost, m.promptField, val)
			}
		}

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
	case dbHostActionDoneMsg:
		logger.TUIOut(string(msg))
		return m, refreshDbHostStateCmd(m.targetDbHost.ID.String())
	case refreshDbBothMsg:
		m.dbHostInfoList.Title = msg.info.title
		m.dbHostInfoList.SetItems(msg.info.items)
		m.targetDbHost = msg.admin.dbHost
		m.dbHostAdminList.Title = "Manage " + msg.admin.dbHost.Name
		m.dbHostAdminList.SetItems(msg.admin.items)
		return m, nil
	}
	return m, nil
}

func viewDatabaseHosts(m model) string {
	switch m.state {
	case viewDatabaseHostList:
		return m.dbHostList.View()
	case viewDatabaseHostInfo:
		return m.dbHostInfoList.View()
	case viewDatabaseHostAdminActions:
		return m.dbHostAdminList.View()
	case viewDatabaseHostCreate:
		var b strings.Builder
		for i := range m.createDbHostInputs {
			b.WriteString(m.createDbHostInputs[i].View())
			if i < len(m.createDbHostInputs)-1 { b.WriteRune('\n') }
		}
		btn := "[ Submit ]"
		if m.createDbHostFocus == len(m.createDbHostInputs) { btn = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(btn) } else { btn = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(btn) }
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Create New Database Host") + "\n\n" + b.String() + "\n\n" + btn + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Use Tab/Shift+Tab to navigate, Enter to submit, Esc to cancel.")
		return "\n" + box.Render(body)
	case viewDbHostPrompt:
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Input Required") + "\n\n" + m.promptDesc + "\n\n" + m.textInput.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to submit, Esc to cancel.")
		return "\n" + box.Render(body)
	}
	return ""
}
