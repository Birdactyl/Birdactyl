package tui

import (
	"fmt"
	"strconv"
	"strings"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"
	"birdactyl-panel-backend/internal/services"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type showNodeListMsg []list.Item
type showNodeInfoMsg struct {
	title string
	items []list.Item
}
type showNodeAdminActionsMsg struct {
	node  models.Node
	items []list.Item
}

type nodeItem struct {
	models.Node
}

func (i nodeItem) Title() string {
	status := "Offline"
	if i.IsOnline {
		status = "Online"
	}
	if i.AuthError {
		status = "Auth Error"
	}
	return fmt.Sprintf("%s (%s) %s", i.Name, i.FQDN, status)
}
func (i nodeItem) Description() string { return fmt.Sprintf("ID: %s | Port: %d", i.ID.String(), i.Port) }
func (i nodeItem) FilterValue() string { return i.Name + " " + i.FQDN + " " + i.ID.String() }

type nodeCreateItem struct{}

func (i nodeCreateItem) Title() string       { return "Create New Node (Manual)" }
func (i nodeCreateItem) Description() string { return "Create an unlinked node reference safely" }
func (i nodeCreateItem) FilterValue() string { return "create new node" }

type nodePairItem struct{}

func (i nodePairItem) Title() string       { return "Pair New Node (Automatic)" }
func (i nodePairItem) Description() string { return "Securely link Axis to Birdactyl automatically" }
func (i nodePairItem) FilterValue() string { return "pair new node" }

func fetchNodesCmd() tea.Cmd {
	return func() tea.Msg {
		nodes, _ := services.GetNodes()
		items := make([]list.Item, len(nodes)+2)
		items[0] = nodePairItem{}
		items[1] = nodeCreateItem{}
		for i, n := range nodes {
			items[i+2] = nodeItem{n}
		}
		return showNodeListMsg(items)
	}
}

func fetchNodeInfoCmd(query string) tea.Cmd {
	return func() tea.Msg {
		var node models.Node
		qc := "name = ?"
		args := []interface{}{query}
		if _, err := uuid.Parse(query); err == nil {
			qc += " OR id = ?"
			args = append(args, query)
		}

		if err := database.DB.Where(qc, args...).First(&node).Error; err != nil {
			return errorMessageMsg(cmdErrorStyle.Render("Node not found: " + query))
		}

		items := []list.Item{
			infoItem{
				"Admin Actions",
				"Reset token or delete this node.",
				fetchNodeAdminActionsCmd(node),
			},
			infoItem{"ID", node.ID.String(), nil},
			infoItem{"Name", node.Name, nil},
			infoItem{"FQDN", node.FQDN, nil},
			infoItem{"Port", fmt.Sprintf("%d", node.Port), nil},
			infoItem{"Online Status", fmt.Sprintf("%v", node.IsOnline), nil},
		}

		if node.IsOnline && node.SystemInfo.Hostname != "" {
			items = append(items, infoItem{"Hostname", node.SystemInfo.Hostname, nil})
			items = append(items, infoItem{"OS", fmt.Sprintf("%s %s (%s)", node.SystemInfo.OS.Name, node.SystemInfo.OS.Version, node.SystemInfo.OS.Arch), nil})
			items = append(items, infoItem{"CPU Cores", fmt.Sprintf("%d", node.SystemInfo.CPU.Cores), nil})
			items = append(items, infoItem{"CPU Usage", fmt.Sprintf("%.1f%%", node.SystemInfo.CPU.Usage), nil})
			items = append(items, infoItem{"Mem Usage", fmt.Sprintf("%.1f%%", node.SystemInfo.Memory.Usage), nil})
			items = append(items, infoItem{"Disk Usage", fmt.Sprintf("%.1f%%", node.SystemInfo.Disk.Usage), nil})
		}

		return showNodeInfoMsg{
			title: "Node Info: " + node.Name,
			items: items,
		}
	}
}

func fetchNodeAdminActionsCmd(node models.Node) tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			infoItem{"Reset Daemon Token", "Force a new secret key for this node", promptNodeAdminActionCmd("reset", node)},
			infoItem{"Delete Node", "Permanently delete node from the database", promptNodeAdminActionCmd("delete", node)},
		}

		return showNodeAdminActionsMsg{
			node:  node,
			items: items,
		}
	}
}

func promptNodeAdminActionCmd(action string, node models.Node) tea.Cmd {
	return func() tea.Msg {
		desc := ""
		if action == "reset" {
			desc = "Are you sure you want to reset the token for " + node.Name + "? This will break your Axis connection until updated."
		} else if action == "delete" {
			desc = "WARNING: Permanently delete node " + node.Name + "?"
		}
		return askConfirmMsg{
			desc: desc,
			cmd: func() tea.Msg {
				return executeDirectNodeAdminAction(action, node)
			},
		}
	}
}

func executeDirectNodeAdminAction(action string, node models.Node) tea.Msg {
	switch action {
	case "reset":
		token, err := services.ResetNodeToken(node.ID)
		if err != nil {
			return errorMessageMsg("Failed to reset: " + err.Error())
		}
		msg := fmt.Sprintf("\n\n======== NODE TOKEN RESET FOR %s ========\nURL: http://localhost:8080\nTOKEN: %s.%s\n==================================================", node.Name, token.TokenID, token.Token)
		return errorMessageMsg(msg)
	case "delete":
		services.DeleteNode(node.ID)
		return errorMessageMsg("Deleted node " + node.Name + ". Returning to logs.")
	}
	return nil
}

type showNodeCreateMsg struct{}

func enterNodeCreateCmd() tea.Cmd {
	return func() tea.Msg {
		return showNodeCreateMsg{}
	}
}

func executeCreateNodeDirectly(vals []string) tea.Cmd {
	return func() tea.Msg {
		if len(vals) < 3 {
			return errorMessageMsg("Form incomplete")
		}

		name := strings.TrimSpace(vals[0])
		if name == "" {
			return errorMessageMsg("Node name is required")
		}
		fqdn := strings.TrimSpace(vals[1])
		if fqdn == "" {
			return errorMessageMsg("Node FQDN is required")
		}

		port, err := strconv.Atoi(vals[2])
		if err != nil {
			port = 8443
		}

		node, token, err := services.CreateNode(name, fqdn, port)
		if err != nil {
			return errorMessageMsg("Failed to create node: " + err.Error())
		}

		msg := fmt.Sprintf("\n\n======== NEW NODE CREATED: %s ========\nURL: http://localhost:8080\nTOKEN: %s\n=====================================================", node.Name, token.DaemonToken)
		return errorMessageMsg(msg)
	}
}

type showNodePairMsg struct{}

func enterNodePairCmd() tea.Cmd {
	return func() tea.Msg {
		return showNodePairMsg{}
	}
}

type askNodePairCodeMsg struct{
	vals []string
	code string
}

func executePairNodeInit(vals []string) tea.Cmd {
	return func() tea.Msg {
		if len(vals) < 3 {
			return errorMessageMsg("Form incomplete")
		}

		name := strings.TrimSpace(vals[0])
		if name == "" {
			return errorMessageMsg("Node name is required")
		}
		fqdn := strings.TrimSpace(vals[1])
		if fqdn == "" {
			return errorMessageMsg("Node FQDN is required")
		}
		port, err := strconv.Atoi(vals[2])
		if err != nil {
			port = 8443
		}

		code := services.GeneratePairingCode()

		return askNodePairCodeMsg{
			vals: []string{name, fqdn, strconv.Itoa(port)},
			code: code,
		}
	}
}

func executePairNodeFinalize(vals []string, code string) tea.Cmd {
	return func() tea.Msg {
		name := vals[0]
		fqdn := vals[1]
		port, _ := strconv.Atoi(vals[2])
		
		node, _, err := services.PairWithNode(name, fqdn, port, "http://localhost:8080", code)
		if err != nil {
			return errorMessageMsg("Pairing failed: " + err.Error())
		}
		return fetchNodeInfoCmd(node.ID.String())()
	}
}

func updateNodes(msg tea.Msg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case viewNodeList:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewLogs
				return m, nil
			case "enter":
				if i, ok := m.nodeList.SelectedItem().(nodeItem); ok {
					return m, fetchNodeInfoCmd(i.ID.String())
				} else if _, ok := m.nodeList.SelectedItem().(nodeCreateItem); ok {
					return m, enterNodeCreateCmd()
				} else if _, ok := m.nodeList.SelectedItem().(nodePairItem); ok {
					return m, enterNodePairCmd()
				}
			}
			m.nodeList, cmd = m.nodeList.Update(msg)
			return m, cmd
		case viewNodeInfo:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewNodeList
				return m, nil
			case "enter":
				if i, ok := m.nodeInfoList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.nodeInfoList, cmd = m.nodeInfoList.Update(msg)
			return m, cmd
		case viewNodeAdminActions:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewNodeInfo
				return m, nil
			case "enter":
				if i, ok := m.nodeAdminList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.nodeAdminList, cmd = m.nodeAdminList.Update(msg)
			return m, cmd
		case viewNodeCreate:
			switch msg.String() {
			case "esc", "ctrl+c":
				m.state = viewNodeList
				return m, nil
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" { m.createNodeFocus-- } else { m.createNodeFocus++ }
				if m.createNodeFocus > len(m.createNodeInputs) { m.createNodeFocus = 0 } else if m.createNodeFocus < 0 { m.createNodeFocus = len(m.createNodeInputs) }
				cmds := make([]tea.Cmd, len(m.createNodeInputs))
				for i := 0; i <= len(m.createNodeInputs)-1; i++ {
					if i == m.createNodeFocus {
						cmds[i] = m.createNodeInputs[i].Focus()
						m.createNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						m.createNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					} else {
						m.createNodeInputs[i].Blur()
						m.createNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						m.createNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					}
				}
				return m, tea.Batch(cmds...)
			case "enter":
				if m.createNodeFocus == len(m.createNodeInputs) {
					var vals []string
					for _, t := range m.createNodeInputs { vals = append(vals, t.Value()) }
					m.state = viewNodeList
					return m, executeCreateNodeDirectly(vals)
				} else {
					m.createNodeFocus++
					cmds := make([]tea.Cmd, len(m.createNodeInputs))
					for i := 0; i < len(m.createNodeInputs); i++ {
						if i == m.createNodeFocus {
							cmds[i] = m.createNodeInputs[i].Focus()
							m.createNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
							m.createNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						} else {
							m.createNodeInputs[i].Blur()
							m.createNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
							m.createNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						}
					}
					return m, tea.Batch(cmds...)
				}
			}
			cmds := make([]tea.Cmd, len(m.createNodeInputs))
			for i := range m.createNodeInputs { m.createNodeInputs[i], cmds[i] = m.createNodeInputs[i].Update(msg) }
			return m, tea.Batch(cmds...)
		case viewNodePair:
			switch msg.String() {
			case "esc", "ctrl+c":
				m.state = viewNodeList
				return m, nil
			case "tab", "shift+tab", "up", "down":
				s := msg.String()
				if s == "up" || s == "shift+tab" { m.pairNodeFocus-- } else { m.pairNodeFocus++ }
				if m.pairNodeFocus > len(m.pairNodeInputs) { m.pairNodeFocus = 0 } else if m.pairNodeFocus < 0 { m.pairNodeFocus = len(m.pairNodeInputs) }
				cmds := make([]tea.Cmd, len(m.pairNodeInputs))
				for i := 0; i <= len(m.pairNodeInputs)-1; i++ {
					if i == m.pairNodeFocus {
						cmds[i] = m.pairNodeInputs[i].Focus()
						m.pairNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						m.pairNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					} else {
						m.pairNodeInputs[i].Blur()
						m.pairNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						m.pairNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					}
				}
				return m, tea.Batch(cmds...)
			case "enter":
				if m.pairNodeFocus == len(m.pairNodeInputs) {
					var vals []string
					for _, t := range m.pairNodeInputs { vals = append(vals, t.Value()) }
					return m, executePairNodeInit(vals)
				} else {
					m.pairNodeFocus++
					cmds := make([]tea.Cmd, len(m.pairNodeInputs))
					for i := 0; i < len(m.pairNodeInputs); i++ {
						if i == m.pairNodeFocus {
							cmds[i] = m.pairNodeInputs[i].Focus()
							m.pairNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
							m.pairNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
						} else {
							m.pairNodeInputs[i].Blur()
							m.pairNodeInputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
							m.pairNodeInputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
						}
					}
					return m, tea.Batch(cmds...)
				}
			}
			cmds := make([]tea.Cmd, len(m.pairNodeInputs))
			for i := range m.pairNodeInputs { m.pairNodeInputs[i], cmds[i] = m.pairNodeInputs[i].Update(msg) }
			return m, tea.Batch(cmds...)
		case viewNodePairWait:
			switch msg.String() {
			case "esc", "ctrl+c":
				m.state = viewNodeList
				return m, nil
			}
		}
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
	}
	return m, nil
}

func viewNodes(m model) string {
	switch m.state {
	case viewNodeList:
		return m.nodeList.View()
	case viewNodeInfo:
		return m.nodeInfoList.View()
	case viewNodeAdminActions:
		return m.nodeAdminList.View()
	case viewNodeCreate:
		var b strings.Builder
		for i := range m.createNodeInputs {
			b.WriteString(m.createNodeInputs[i].View())
			if i < len(m.createNodeInputs)-1 { b.WriteRune('\n') }
		}
		btn := "[ Submit ]"
		if m.createNodeFocus == len(m.createNodeInputs) { btn = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(btn) } else { btn = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(btn) }
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Create New Node") + "\n\n" + b.String() + "\n\n" + btn + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Use Tab/Shift+Tab to navigate, Enter to create manually, Esc to cancel.")
		return "\n" + box.Render(body)
	case viewNodePair:
		var b strings.Builder
		for i := range m.pairNodeInputs {
			b.WriteString(m.pairNodeInputs[i].View())
			if i < len(m.pairNodeInputs)-1 { b.WriteRune('\n') }
		}
		btn := "[ Pair Request ]"
		if m.pairNodeFocus == len(m.pairNodeInputs) { btn = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(btn) } else { btn = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(btn) }
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Pair New Node") + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("74")).Render("Make sure to run 'axis pair' on the target machine BEFORE doing this!") + "\n\n" + b.String() + "\n\n" + btn + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Use Tab/Shift+Tab to navigate, Enter to send request, Esc to cancel.")
		return "\n" + box.Render(body)
	case viewNodePairWait:
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 4)
		codeView := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Padding(1, 2).Render(m.pairNodeCode)
		body := helpTitleStyle.Render("Waiting For Machine...") + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Check your Node's remote terminal. Does this code match?") + "\n" + codeView + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Waiting for axis to allow pairing... Press Esc to abort.")
		return "\n" + box.Render(body)
	}
	return ""
}
