package tui

import (
	"strings"

	"birdactyl-panel-backend/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

func handleCommand(cmd string) tea.Cmd {
	logger.Command("%s", cmd)

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch strings.ToLower(parts[0]) {
	case "help":
		body := helpTitleStyle.Render("Available Commands:") + "\n" +
			"  help            - Shows this help message\n" +
			"  user            - View members in an interactive list\n" +
			"  user <query>    - Get details for a specific user\n" +
			"  server          - View servers in an interactive list\n" +
			"  server <query>  - Get details for a specific server\n" +
			"  node            - View and manage Daemon nodes\n" +
			"  dbhost          - View MySQL database hosts\n" +
			"  mount           - View global mounts\n" +
			"  logs            - View all global activity logs\n" +
			"  exit            - Shuts down the panel"
		logger.TUIOut(helpBoxStyle.Render(body))
	case "exit":
		return tea.Quit
	case "user":
		if len(parts) == 1 {
			return fetchUsersCmd()
		}
		query := strings.Join(parts[1:], " ")
		return fetchUserInfoCmd(query)
	case "server":
		if len(parts) == 1 {
			return fetchServersCmd()
		}
		if len(parts) > 1 && strings.ToLower(parts[1]) == "create" {
			return enterServerCreateCmd()
		}
		query := strings.Join(parts[1:], " ")
		return fetchServerInfoCmd(query)
	case "node":
		if len(parts) == 1 {
			return fetchNodesCmd()
		}
		if len(parts) > 1 && strings.ToLower(parts[1]) == "create" {
			return enterNodeCreateCmd()
		}
		if len(parts) > 1 && strings.ToLower(parts[1]) == "pair" {
			return enterNodePairCmd()
		}
		query := strings.Join(parts[1:], " ")
		return fetchNodeInfoCmd(query)
	case "logs", "log":
		return fetchGlobalActivityCmd()
	case "dbhost", "dbhosts":
		if len(parts) == 1 {
			return fetchDatabaseHostsCmd()
		}
		if len(parts) > 1 && strings.ToLower(parts[1]) == "create" {
			return enterDatabaseHostCreateCmd()
		}
		query := strings.Join(parts[1:], " ")
		return fetchDatabaseHostInfoCmd(query)
	case "mount", "mounts":
		if len(parts) == 1 {
			return fetchMountsCmd()
		}
		if len(parts) > 1 && strings.ToLower(parts[1]) == "create" {
			return enterMountCreateCmd()
		}
		query := strings.Join(parts[1:], " ")
		return fetchMountInfoCmd(query)
	default:
		logger.TUIOut(cmdErrorStyle.Render("Unknown command. Type help for a list of commands."))
	}

	return nil
}
