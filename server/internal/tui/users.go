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

func fetchUsersCmd() tea.Cmd {
	return func() tea.Msg {
		var users []models.User
		database.DB.Find(&users)
		items := make([]list.Item, len(users))
		for i, u := range users {
			items[i] = userItem{u}
		}
		return showUserListMsg(items)
	}
}

func fetchUserInfoCmd(query string) tea.Cmd {
	return func() tea.Msg {
		var user models.User
		qc := "username = ? OR email = ?"
		args := []interface{}{query, query}
		if _, err := uuid.Parse(query); err == nil {
			qc += " OR id = ?"
			args = append(args, query)
		}

		if err := database.DB.Where(qc, args...).First(&user).Error; err != nil {
			return errorMessageMsg(cmdErrorStyle.Render("User not found: " + query))
		}

		var sessionCount, serverCount, subuserCount, apiKeyCount, ipRegCount, logCount int64
		database.DB.Model(&models.Session{}).Where("user_id = ?", user.ID).Count(&sessionCount)
		database.DB.Model(&models.Server{}).Where("user_id = ?", user.ID).Count(&serverCount)
		database.DB.Model(&models.Subuser{}).Where("user_id = ?", user.ID).Count(&subuserCount)
		database.DB.Model(&models.APIKey{}).Where("user_id = ?", user.ID).Count(&apiKeyCount)
		database.DB.Model(&models.IPRegistration{}).Where("user_id = ?", user.ID).Count(&ipRegCount)
		database.DB.Model(&models.ActivityLog{}).Where("user_id = ?", user.ID).Count(&logCount)

		formatLimit := func(l *int, suffix string) string {
			if l == nil {
				return "Unlimited"
			}
			return fmt.Sprintf("%d%s", *l, suffix)
		}

		role := "Standard User"
		if user.IsAdmin {
			role = "Administrator"
		}
		status := "Active"
		if user.IsBanned {
			status = "BANNED"
		}

		items := []list.Item{
			infoItem{
				"Admin Actions",
				"Modify limits, ban, delete, or forcibly reset this user.",
				fetchAdminActionsCmd(user),
			},
			infoItem{"ID", user.ID.String(), nil},
			infoItem{"Username", user.Username, nil},
			infoItem{"Email", user.Email, nil},
			infoItem{"Register IP", user.RegisterIP, nil},
			infoItem{"Role", role, nil},
			infoItem{"Account Status", status, nil},
			infoItem{"Email Verified", fmt.Sprintf("%v", user.EmailVerified), nil},
			infoItem{"TOTP Enabled", fmt.Sprintf("%v", user.TOTPEnabled), nil},
			infoItem{"Force Pass Reset", fmt.Sprintf("%v", user.ForcePasswordReset), nil},
			infoItem{"Server Limit", formatLimit(user.ServerLimit, ""), nil},
			infoItem{"RAM Limit", formatLimit(user.RAMLimit, " MB"), nil},
			infoItem{"CPU Limit", formatLimit(user.CPULimit, " %"), nil},
			infoItem{"Disk Limit", formatLimit(user.DiskLimit, " MB"), nil},
			infoItem{"Servers Owned", fmt.Sprintf("%d", serverCount), nil},
			infoItem{"Subuser Accesses", fmt.Sprintf("%d", subuserCount), nil},
			infoItem{"Active API Keys", fmt.Sprintf("%d", apiKeyCount), nil},
			infoItem{"Known Login IPs", fmt.Sprintf("%d", ipRegCount), nil},
			infoItem{"Active Sessions", fmt.Sprintf("%d", sessionCount), nil},
			infoItem{"Created At", user.CreatedAt.Format("2006-01-02 15:04:05"), nil},
			infoItem{"Updated At", user.UpdatedAt.Format("2006-01-02 15:04:05"), nil},
		}

		if logCount > 0 {
			items = append(items, infoItem{
				"Activity Logs",
				fmt.Sprintf("View %d records (Press Enter)", logCount),
				fetchUserActivityCmd(user.ID),
			})
		}

		if serverCount > 0 {
			items = append(items, infoItem{
				"User's Servers",
				fmt.Sprintf("View %d servers owned by this user (Press Enter)", serverCount),
				fetchUserServersCmd(user.ID),
			})
		}

		return showUserInfoMsg{
			title: "User Info: " + user.Username,
			items: items,
		}
	}
}

func fetchAdminActionsCmd(user models.User) tea.Cmd {
	return func() tea.Msg {
		banTopic, banDesc, banType := "Ban User", "Suspend this user's account", "ban"
		if user.IsBanned {
			banTopic, banDesc, banType = "Unban User", "Un-suspend this user's account", "unban"
		}

		adminTopic, adminDesc, adminType := "Grant Admin", "Make this user a root administrator", "giveadmin"
		if user.IsAdmin {
			adminTopic, adminDesc, adminType = "Revoke Admin", "Remove this user's root privileges", "takeadmin"
		}

		items := []list.Item{
			infoItem{banTopic, banDesc, confirmAdminExecCmd(banType, "Are you sure you want to "+strings.ToLower(banTopic)+" "+user.Username+"?", user)},
			infoItem{adminTopic, adminDesc, confirmAdminExecCmd(adminType, "Are you sure you want to "+strings.ToLower(adminTopic)+" for "+user.Username+"?", user)},
			infoItem{"Reset Password", "Force a password reset (clears sessions)", confirmAdminExecCmd("resetpw", "Forcibly reset password and clear sessions for "+user.Username+"?", user)},
		}

		if user.TOTPEnabled {
			items = append(items, infoItem{"Disable 2FA", "Disable TOTP on this account", confirmAdminExecCmd("disable2fa", "Administratively remove 2FA from "+user.Username+"'s account?", user)})
		} else {
			items = append(items, infoItem{"Disable 2FA", "User does not have 2FA enabled", nil})
		}

		items = append(items,
			infoItem{"Delete User", "Permanently delete account and activity", confirmAdminExecCmd("delete", "WARNING: Permanently delete user "+user.Username+" and all metadata?", user)},
			infoItem{"Ban User's IPs", "Ban all known IP addresses used by this user", confirmAdminExecCmd("banips", "Are you sure you want to ban ALL known IPs for "+user.Username+"?", user)},
			infoItem{"Edit Boundaries / Limits", "Modify maximum RAM, CPU, Disk, and Servers", func() tea.Msg { return fetchAdminLimitsCmd(user)() }},
			infoItem{"Edit Username", "Change panel handle", promptAdminEditCmd("Enter new username", "username", user.Username)},
			infoItem{"Edit Email", "Change authentication email", promptAdminEditCmd("Enter new email", "email", user.Email)},
			infoItem{"Edit Password", "Directly set the user's password", promptAdminEditCmd("Enter new password (8+ chars)", "password", "")},
		)

		return showAdminActionsMsg{
			user:  user,
			items: items,
		}
	}
}

func fetchAdminLimitsCmd(user models.User) tea.Cmd {
	return func() tea.Msg {
		fmtIntPtr := func(v *int) string {
			if v == nil {
				return "0"
			}
			return strconv.Itoa(*v)
		}

		items := []list.Item{
			infoItem{"Edit Limits [RAM]", "Format: 1024 (MB) or 0 (unlimited)", promptAdminEditCmd("Edit RAM Limit (0 = Unlimited, MB)", "ram_limit", fmtIntPtr(user.RAMLimit))},
			infoItem{"Edit Limits [CPU]", "Format: 200 (%) or 0 (unlimited)", promptAdminEditCmd("Edit CPU Limit (0 = Unlimited, %)", "cpu_limit", fmtIntPtr(user.CPULimit))},
			infoItem{"Edit Limits [Disk]", "Format: 5000 (MB) or 0 (unlimited)", promptAdminEditCmd("Edit Disk Limit (0 = Unlimited, MB)", "disk_limit", fmtIntPtr(user.DiskLimit))},
			infoItem{"Edit Limits [Servers]", "Format: 3 or 0 (unlimited)", promptAdminEditCmd("Edit Server Limit (0 = Unlimited)", "server_limit", fmtIntPtr(user.ServerLimit))},
		}

		return showAdminLimitsMsg{
			user:  user,
			items: items,
		}
	}
}

func confirmAdminExecCmd(action string, desc string, user models.User) tea.Cmd {
	return func() tea.Msg {
		return askConfirmMsg{
			desc: desc,
			cmd: func() tea.Msg {
				return executeDirectAdminAction(action, user)
			},
		}
	}
}

func promptAdminEditCmd(desc string, field string, placeholder string) tea.Cmd {
	return func() tea.Msg {
		return askPromptMsg{desc: desc, field: field, placeholder: placeholder}
	}
}

func executeDirectAdminAction(action string, user models.User) tea.Msg {
	currentUser := models.User{Username: "SYSTEM (TUI)", ID: uuid.Nil}
	switch action {
	case "ban":
		database.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("is_banned", true)
		database.DB.Where("user_id = ?", user.ID).Delete(&models.Session{})
		return actionDoneMsg("Banned " + user.Username)
	case "unban":
		database.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("is_banned", false)
		return actionDoneMsg("Unbanned " + user.Username)
	case "giveadmin":
		database.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("is_admin", true)
		return actionDoneMsg("Granted Admin to " + user.Username)
	case "takeadmin":
		database.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("is_admin", false)
		return actionDoneMsg("Revoked Admin from " + user.Username)
	case "resetpw":
		database.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("force_password_reset", true)
		database.DB.Where("user_id = ?", user.ID).Delete(&models.Session{})
		return actionDoneMsg("Forced password reset for " + user.Username)
	case "disable2fa":
		database.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{"totp_enabled": false, "totp_secret": "", "backup_codes": ""})
		return actionDoneMsg("Disabled 2FA for " + user.Username)
	case "delete":
		database.DB.Where("user_id = ?", user.ID).Delete(&models.Session{})
		database.DB.Where("user_id = ?", user.ID).Delete(&models.ActivityLog{})
		database.DB.Where("id = ?", user.ID).Delete(&models.User{})
		return errorMessageMsg("Deleted user " + user.Username + ". Returning to logs.")
	case "banips":
		var ips []string
		database.DB.Model(&models.IPRegistration{}).Where("user_id = ?", user.ID).Pluck("ip", &ips)
		
		if user.RegisterIP != "" {
			ips = append(ips, user.RegisterIP)
		}

		bannedCount := 0
		for _, ip := range ips {
			var existing models.IPBan
			if database.DB.Where("ip = ?", ip).First(&existing).Error != nil {
				ban := models.IPBan{
					IP:       ip,
					Reason:   "Banned by admin via TUI for user " + user.Username,
					BannedBy: uuid.Nil,
				}
				database.DB.Create(&ban)
				bannedCount++
			}
		}
		
		return actionDoneMsg("Banned " + strconv.Itoa(bannedCount) + " unique IP(s) associated with " + user.Username)
	}
	_ = currentUser
	return nil
}

func executeAdminPromptConfig(user models.User, field string, val string) tea.Cmd {
	return func() tea.Msg {
		val = strings.TrimSpace(val)
		if val == "" {
			return actionDoneMsg("Edit cancelled: No value provided.")
		}

		updates := map[string]interface{}{}
		
		switch field {
		case "username":
			var existing models.User
			if database.DB.Where("username = ? AND id != ?", val, user.ID).First(&existing).Error == nil {
				return actionDoneMsg("Edit failed: Username already taken.")
			}
			updates["username"] = val
		case "email":
			var existing models.User
			if database.DB.Where("email = ? AND id != ?", val, user.ID).First(&existing).Error == nil {
				return actionDoneMsg("Edit failed: Email already in use.")
			}
			updates["email"] = val
		case "password":
			if len(val) < 8 {
				return actionDoneMsg("Edit failed: Password must be 8+ characters.")
			}
			hash, err := services.HashPassword(val)
			if err != nil {
				return actionDoneMsg("Edit failed: Hash error.")
			}
			updates["password_hash"] = hash
		case "ram_limit", "cpu_limit", "disk_limit", "server_limit":
			i, err := strconv.Atoi(val)
			if err != nil {
				return actionDoneMsg("Edit failed: Value must be an integer.")
			}
			if i <= 0 {
				updates[field] = nil
			} else {
				updates[field] = i
			}
		}

		if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
			return actionDoneMsg("Edit failed: DB error " + err.Error())
		}
		
		return actionDoneMsg("Updated " + field + " for " + user.Username)
	}
}

func refreshAdminStateCmd(uid string) tea.Cmd {
	return func() tea.Msg {
		infoMsg := fetchUserInfoCmd(uid)()
		if errInfo, ok := infoMsg.(errorMessageMsg); ok {
			return errInfo
		}

		var user models.User
		if err := database.DB.Where("id = ?", uid).First(&user).Error; err != nil {
			return errorMessageMsg("User not found: " + uid)
		}
		
		adminMsg := fetchAdminActionsCmd(user)()
		limitsMsg := fetchAdminLimitsCmd(user)()

		var allUsers []models.User
		database.DB.Find(&allUsers)
		userItems := make([]list.Item, len(allUsers))
		for i, u := range allUsers {
			userItems[i] = userItem{u}
		}

		return refreshBothMsg{
			info:   infoMsg.(showUserInfoMsg),
			admin:  adminMsg.(showAdminActionsMsg),
			limits: limitsMsg.(showAdminLimitsMsg),
			users:  userItems,
		}
	}
}

func updateUsers(msg tea.Msg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case viewUserList:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewLogs
				return m, nil
			case "enter":
				if i, ok := m.userList.SelectedItem().(userItem); ok {
					return m, fetchUserInfoCmd(i.ID.String())
				}
			}
			m.userList, cmd = m.userList.Update(msg)
			return m, cmd
		case viewUserInfo:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewUserList
				return m, nil
			case "enter":
				if i, ok := m.userInfoList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.userInfoList, cmd = m.userInfoList.Update(msg)
			return m, cmd
		case viewAdminActions:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewUserInfo
				return m, nil
			case "enter":
				if i, ok := m.adminActionsList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.adminActionsList, cmd = m.adminActionsList.Update(msg)
			return m, cmd
		case viewAdminLimits:
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				m.state = viewAdminActions
				return m, nil
			case "enter":
				if i, ok := m.adminLimitsList.SelectedItem().(infoItem); ok && i.actionCmd != nil {
					return m, i.actionCmd
				}
			}
			m.adminLimitsList, cmd = m.adminLimitsList.Update(msg)
			return m, cmd
		case viewAdminConfirm:
			switch strings.ToLower(msg.String()) {
			case "y":
				resCmd := m.pendingConfirmCmd
				m.state = m.previousAdminState
				return m, resCmd
			case "n", "esc", "q", "ctrl+c":
				m.state = m.previousAdminState
				return m, nil
			}
			return m, nil
		case viewAdminPrompt:
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
				return m, executeAdminPromptConfig(m.targetUser, m.promptField, val)
			}
		}
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
	case actionDoneMsg:
		logger.TUIOut(string(msg))
		return m, refreshAdminStateCmd(m.targetUser.ID.String())
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
	}
	return m, nil
}

func viewUsers(m model) string {
	switch m.state {
	case viewUserList:
		return m.userList.View()
	case viewUserInfo:
		return m.userInfoList.View()
	case viewAdminActions:
		return m.adminActionsList.View()
	case viewAdminLimits:
		return m.adminLimitsList.View()
	case viewAdminConfirm:
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("196")).Padding(1, 2)
		body := helpTitleStyle.Render("Confirm Action") + "\n\n" + m.pendingConfirmDesc + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press (y) to Confirm, or (n) / (Esc) to Cancel")
		return "\n" + box.Render(body)
	case viewAdminPrompt:
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2)
		body := helpTitleStyle.Render("Input Required") + "\n\n" + m.promptDesc + "\n\n" + m.textInput.View() + "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to submit, Esc to cancel.")
		return "\n" + box.Render(body)
	}
	return ""
}
