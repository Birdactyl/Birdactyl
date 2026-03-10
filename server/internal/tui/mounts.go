package tui

import (
	"fmt"
	"strings"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/logger"
	"birdactyl-panel-backend/internal/models"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type mountItem struct {
	models.Mount
	NodeCount    int
	PackageCount int
}

func (i mountItem) Title() string {
	return fmt.Sprintf("%s", i.Name)
}
func (i mountItem) Description() string {
	return fmt.Sprintf("ID: %s | Source: %s | Target: %s", i.ID.String(), i.Source, i.Target)
}
func (i mountItem) FilterValue() string {
	return fmt.Sprintf("%s %s %s %s", i.Name, i.Source, i.Target, i.ID.String())
}

type mountCreateItem struct{}

func (i mountCreateItem) Title() string       { return "Create New Mount" }
func (i mountCreateItem) Description() string { return "Create a global mount point for servers" }
func (i mountCreateItem) FilterValue() string { return "create new mount global folder" }

func fetchMountsCmd() tea.Cmd {
	return func() tea.Msg {
		var mounts []models.Mount
		database.DB.Preload("Nodes").Preload("Packages").Find(&mounts)
		
		items := make([]list.Item, len(mounts)+1)
		items[0] = mountCreateItem{}
		
		for i, m := range mounts {
			items[i+1] = mountItem{m, len(m.Nodes), len(m.Packages)}
		}
		
		return showMountListMsg(items)
	}
}

func fetchMountInfoCmd(query string) tea.Cmd {
	return func() tea.Msg {
		var mount models.Mount
		qc := "id = ?"
		if _, err := uuid.Parse(query); err != nil {
			qc = "name = ?"
		}

		if err := database.DB.Preload("Nodes").Preload("Packages").Where(qc, query).First(&mount).Error; err != nil {
			return errorMessageMsg("Mount not found: " + query)
		}

		items := []list.Item{
			infoItem{
				"Admin Actions",
				"Edit settings or delete this mount.",
				fetchMountAdminActionsCmd(mount),
			},
			infoItem{"ID", mount.ID.String(), nil},
			infoItem{"Name", mount.Name, nil},
			infoItem{"Description", mount.Description, nil},
			infoItem{"Source Path", mount.Source, nil},
			infoItem{"Target Path", mount.Target, nil},
			infoItem{"Read Only", fmt.Sprintf("%v", mount.ReadOnly), nil},
			infoItem{"User Mountable", fmt.Sprintf("%v", mount.UserMountable), nil},
			infoItem{"Navigable", fmt.Sprintf("%v", mount.Navigable), nil},
			infoItem{"Nodes Bound", fmt.Sprintf("%d", len(mount.Nodes)), nil},
			infoItem{"Packages Bound", fmt.Sprintf("%d", len(mount.Packages)), nil},
			infoItem{"Created At", mount.CreatedAt.Format("2006-01-02 15:04:05"), nil},
		}

		return showMountInfoMsg{
			title: "Mount Info: " + mount.Name,
			items: items,
		}
	}
}

func fetchMountAdminActionsCmd(mount models.Mount) tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			infoItem{"Delete Mount", "Permanently remove this mount", confirmMountAdminExecCmd("delete_mount", "WARNING: Permanently delete mount "+mount.Name+"?", mount)},
			infoItem{"Edit Name", "Change mount display name", promptMountPromptCmd("Enter new name", "mount_name", mount.Name)},
			infoItem{"Edit Description", "Change mount description", promptMountPromptCmd("Enter new description", "mount_desc", mount.Description)},
			infoItem{"Edit Source Path", "Change the host source directory", promptMountPromptCmd("Enter new source path", "mount_source", mount.Source)},
			infoItem{"Edit Target Path", "Change the container target directory", promptMountPromptCmd("Enter new target path", "mount_target", mount.Target)},
			infoItem{"Toggle Read Only", "Toggle read-only state", confirmMountAdminExecCmd("toggle_read", "Toggle read-only state?", mount)},
			infoItem{"Toggle User Mountable", "Toggle user mountability", confirmMountAdminExecCmd("toggle_mountable", "Toggle user mountability?", mount)},
			infoItem{"Toggle Navigable", "Toggle file manager navigation", confirmMountAdminExecCmd("toggle_navigable", "Toggle file manager navigation?", mount)},
			infoItem{"Add Node", "Bind this mount to a node", promptMountPromptCmd("Enter node name or UUID to add", "add_node", "")},
			infoItem{"Remove Node", "Unbind this mount from a node", promptMountPromptCmd("Enter node name or UUID to remove", "remove_node", "")},
			infoItem{"Add Package", "Bind this mount to a package", promptMountPromptCmd("Enter package name or UUID to add", "add_package", "")},
			infoItem{"Remove Package", "Unbind this mount from a package", promptMountPromptCmd("Enter package name or UUID to remove", "remove_package", "")},
		}

		return showMountAdminActionsMsg{
			mount: mount,
			items: items,
		}
	}
}

func confirmMountAdminExecCmd(action string, desc string, mount models.Mount) tea.Cmd {
	return func() tea.Msg {
		return askConfirmMsg{
			desc: desc,
			cmd: func() tea.Msg {
				return executeDirectMountAction(action, mount)
			},
		}
	}
}

func promptMountPromptCmd(desc string, field string, placeholder string) tea.Cmd {
	return func() tea.Msg {
		return askMountPromptMsg{desc: desc, field: field, placeholder: placeholder}
	}
}

func executeDirectMountAction(action string, mount models.Mount) tea.Msg {
	switch action {
	case "delete_mount":
		database.DB.Where("id = ?", mount.ID).Delete(&models.Mount{})
		return errorMessageMsg("Deleted mount " + mount.Name + ". Returning to logs.")
	case "toggle_read":
		mount.ReadOnly = !mount.ReadOnly
		database.DB.Save(&mount)
		return mountActionDoneMsg("Toggled read only for " + mount.Name)
	case "toggle_mountable":
		mount.UserMountable = !mount.UserMountable
		database.DB.Save(&mount)
		return mountActionDoneMsg("Toggled user mountable for " + mount.Name)
	case "toggle_navigable":
		mount.Navigable = !mount.Navigable
		database.DB.Save(&mount)
		return mountActionDoneMsg("Toggled navigable for " + mount.Name)
	}
	return nil
}

func executeMountPromptConfig(mount models.Mount, field string, val string) tea.Cmd {
	return func() tea.Msg {
		val = strings.TrimSpace(val)
		if val == "" {
			return mountActionDoneMsg("Edit cancelled: No value provided.")
		}

		updates := map[string]interface{}{}
		
		switch field {
		case "mount_name":
			updates["name"] = val
		case "mount_desc":
			updates["description"] = val
		case "mount_source":
			updates["source"] = val
		case "mount_target":
			updates["target"] = val
		case "add_node":
			var node models.Node
			qc := "name = ?"
			if _, err := uuid.Parse(val); err == nil { qc += " OR id = ?" }
			if database.DB.Where(qc, val, val).First(&node).Error == nil {
				database.DB.Model(&mount).Association("Nodes").Append(&node)
				return mountActionDoneMsg("Added node " + val + " to mount")
			}
			return mountActionDoneMsg("Node not found")
		case "remove_node":
			var node models.Node
			qc := "name = ?"
			if _, err := uuid.Parse(val); err == nil { qc += " OR id = ?" }
			if database.DB.Where(qc, val, val).First(&node).Error == nil {
				database.DB.Model(&mount).Association("Nodes").Delete(&node)
				return mountActionDoneMsg("Removed node " + val + " from mount")
			}
			return mountActionDoneMsg("Node not found")
		case "add_package":
			var pkg models.Package
			qc := "name = ?"
			if _, err := uuid.Parse(val); err == nil { qc += " OR id = ?" }
			if database.DB.Where(qc, val, val).First(&pkg).Error == nil {
				database.DB.Model(&mount).Association("Packages").Append(&pkg)
				return mountActionDoneMsg("Added package " + val + " to mount")
			}
			return mountActionDoneMsg("Package not found")
		case "remove_package":
			var pkg models.Package
			qc := "name = ?"
			if _, err := uuid.Parse(val); err == nil { qc += " OR id = ?" }
			if database.DB.Where(qc, val, val).First(&pkg).Error == nil {
				database.DB.Model(&mount).Association("Packages").Delete(&pkg)
				return mountActionDoneMsg("Removed package " + val + " from mount")
			}
			return mountActionDoneMsg("Package not found")
		}

		if len(updates) > 0 {
			if err := database.DB.Model(&mount).Updates(updates).Error; err != nil {
				return mountActionDoneMsg("Edit failed: DB error " + err.Error())
			}
		}
		
		return mountActionDoneMsg("Updated " + field + " for " + mount.Name)
	}
}

func refreshMountStateCmd(id string) tea.Cmd {
	return func() tea.Msg {
		infoMsg := fetchMountInfoCmd(id)()
		if errInfo, ok := infoMsg.(errorMessageMsg); ok {
			return errInfo
		}

		var mount models.Mount
		database.DB.Where("id = ?", id).First(&mount)

		adminMsg := fetchMountAdminActionsCmd(mount)()

		return refreshMountBothMsg{
			info:  infoMsg.(showMountInfoMsg),
			admin: adminMsg.(showMountAdminActionsMsg),
		}
	}
}

func enterMountCreateCmd() tea.Cmd {
	return func() tea.Msg {
		return showMountCreateMsg{}
	}
}

func executeCreateMountDirectly(vals []string) tea.Cmd {
	return func() tea.Msg {
		name := strings.TrimSpace(vals[0])
		desc := strings.TrimSpace(vals[1])
		source := strings.TrimSpace(vals[2])
		target := strings.TrimSpace(vals[3])

		if name == "" || source == "" || target == "" {
			return errorMessageMsg("Name, Source Path, and Target Path are required")
		}

		mount := models.Mount{
			Name:          name,
			Description:   desc,
			Source:        source,
			Target:        target,
			ReadOnly:      false,
			UserMountable: false,
			Navigable:     false,
		}

		if err := database.DB.Create(&mount).Error; err != nil {
			return errorMessageMsg("Failed to create mount: " + err.Error())
		}

		return fetchMountsCmd()()
	}
}

func updateMounts(msg tea.Msg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case viewMountList:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewLogs
				return m, nil
			case "enter":
				if i, ok := m.mountList.SelectedItem().(mountItem); ok {
					return m, fetchMountInfoCmd(i.ID.String())
				} else if _, ok := m.mountList.SelectedItem().(mountCreateItem); ok {
					return m, enterMountCreateCmd()
				}
			}
			m.mountList, cmd = m.mountList.Update(msg)
			return m, cmd
		case viewMountInfo:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewMountList
				return m, nil
			case "enter":
				if i, ok := m.mountInfoList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.mountInfoList, cmd = m.mountInfoList.Update(msg)
			return m, cmd
		case viewMountAdminActions:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewMountInfo
				return m, nil
			case "enter":
				if i, ok := m.mountAdminList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.mountAdminList, cmd = m.mountAdminList.Update(msg)
			return m, cmd
		case viewMountCreate:
			switch msg.String() {
			case "esc", "ctrl+c":
				m.state = viewMountList
				return m, nil
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" { m.createMountFocus-- } else { m.createMountFocus++ }
				if m.createMountFocus > len(m.createMountInputs) { m.createMountFocus = 0 } else if m.createMountFocus < 0 { m.createMountFocus = len(m.createMountInputs) }
				cmds := make([]tea.Cmd, len(m.createMountInputs))
				for i := 0; i <= len(m.createMountInputs)-1; i++ {
					if i == m.createMountFocus {
						cmds[i] = m.createMountInputs[i].Focus()
						m.createMountInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						m.createMountInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					} else {
						m.createMountInputs[i].Blur()
						m.createMountInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						m.createMountInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					}
				}
				return m, tea.Batch(cmds...)
			case "enter":
				if m.createMountFocus == len(m.createMountInputs) {
					var vals []string
					for _, t := range m.createMountInputs { vals = append(vals, t.Value()) }
					m.state = viewMountList
					return m, executeCreateMountDirectly(vals)
				} else {
					m.createMountFocus++
					cmds := make([]tea.Cmd, len(m.createMountInputs))
					for i := 0; i < len(m.createMountInputs); i++ {
						if i == m.createMountFocus {
							cmds[i] = m.createMountInputs[i].Focus()
							m.createMountInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
							m.createMountInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						} else {
							m.createMountInputs[i].Blur()
							m.createMountInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
							m.createMountInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						}
					}
					return m, tea.Batch(cmds...)
				}
			}
			cmds := make([]tea.Cmd, len(m.createMountInputs))
			for i := range m.createMountInputs { m.createMountInputs[i], cmds[i] = m.createMountInputs[i].Update(msg) }
			return m, tea.Batch(cmds...)
		case viewMountPrompt:
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
				return m, executeMountPromptConfig(m.targetMount, m.promptField, val)
			}
		}
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
	case mountActionDoneMsg:
		logger.TUIOut(string(msg))
		return m, refreshMountStateCmd(m.targetMount.ID.String())

	case refreshMountBothMsg:
		m.mountInfoList.Title = msg.info.title
		m.mountInfoList.SetItems(msg.info.items)
		m.targetMount = msg.admin.mount
		m.mountAdminList.Title = "Manage " + msg.admin.mount.Name
		m.mountAdminList.SetItems(msg.admin.items)
		return m, nil
	}
	return m, nil
}

func viewMounts(m model) string {
	switch m.state {
	case viewMountList:
		return m.mountList.View()
	case viewMountInfo:
		return m.mountInfoList.View()
	case viewMountAdminActions:
		return m.mountAdminList.View()
	case viewMountCreate:
		var b strings.Builder
		for i := range m.createMountInputs {
			b.WriteString(m.createMountInputs[i].View())
			if i < len(m.createMountInputs)-1 { b.WriteRune('\n') }
		}
		btn := "[ Submit ]"
		if m.createMountFocus == len(m.createMountInputs) { btn = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(btn) } else { btn = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(btn) }
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Create New Mount") + "\n\n" + b.String() + "\n\n" + btn + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Use Tab/Shift+Tab to navigate, Enter to submit, Esc to cancel.")
		return "\n" + box.Render(body)
	case viewMountPrompt:
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Input Required") + "\n\n" + m.promptDesc + "\n\n" + m.textInput.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to submit, Esc to cancel.")
		return "\n" + box.Render(body)
	}
	return ""
}
