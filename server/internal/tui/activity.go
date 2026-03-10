package tui

import (
	"fmt"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

func fetchUserActivityCmd(uid uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		var logs []models.ActivityLog
		database.DB.Where("user_id = ?", uid).Order("created_at desc").Limit(100).Find(&logs)
		items := make([]list.Item, len(logs))
		for i, l := range logs {
			items[i] = activityItem{l}
		}
		return showUserActivityMsg{items: items, isGlobal: false}
	}
}

func fetchGlobalActivityCmd() tea.Cmd {
	return func() tea.Msg {
		var logs []models.ActivityLog
		database.DB.Order("created_at desc").Limit(200).Find(&logs)
		items := make([]list.Item, len(logs))
		for i, l := range logs {
			items[i] = activityItem{l}
		}
		return showUserActivityMsg{items: items, isGlobal: true}
	}
}

func fetchActivityDetailsCmd(log models.ActivityLog) tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			infoItem{"ID", log.ID.String(), nil},
			infoItem{"Action Record", log.Action, nil},
			infoItem{"Origin IP", log.IP, nil},
			infoItem{"User Agent", log.UserAgent, nil},
			infoItem{"System Flag", fmt.Sprintf("Admin Interaction: %v", log.IsAdmin), nil},
		}

		if log.Metadata != "" {
			items = append(items, infoItem{"Metadata Content", log.Metadata, nil})
		}
		
		items = append(items, infoItem{"Timestamp", log.CreatedAt.Format("2006-01-02 15:04:05"), nil})

		return showActivityDetailsMsg{
			title: "Log Details: " + log.Description,
			items: items,
		}
	}
}

func updateActivity(msg tea.Msg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case viewUserActivity:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				if m.isGlobalActivity {
					m.state = viewLogs
				} else {
					m.state = viewUserInfo
				}
				return m, nil
			case "enter":
				if i, ok := m.userActivityList.SelectedItem().(activityItem); ok {
					return m, fetchActivityDetailsCmd(i.ActivityLog)
				}
			}
			m.userActivityList, cmd = m.userActivityList.Update(msg)
			return m, cmd
		case viewActivityDetails:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewUserActivity
				return m, nil
			}
			m.activityDetailsList, cmd = m.activityDetailsList.Update(msg)
			return m, cmd
		}
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
	}
	return m, nil
}

func viewActivity(m model) string {
	switch m.state {
	case viewUserActivity:
		return m.userActivityList.View()
	case viewActivityDetails:
		return m.activityDetailsList.View()
	}
	return ""
}
