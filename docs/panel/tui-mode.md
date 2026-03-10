# Terminal User Interface (TUI)

Birdactyl Panel features a built-in, fully interactive Terminal User Interface (TUI) designed to allow administrators to manage their entire infrastructure directly from the server console—without ever needing to open a web browser.

The TUI is built using [Bubble Tea](https://github.com/charmbracelet/bubbletea) and provides a polished, keyboard-driven experience.

## Running the Panel

By default, launching the panel executable will start it in **headless** mode (running quietly in the background, logging to standard output).

You can control how the panel starts using positional arguments:

### 1. Headless Mode (Default)

```bash
./panel
# OR
./panel headless
```

Running in headless mode disables the interactive TUI. This is the recommended mode when running Birdactyl Panel as a background service (e.g., using Systemd or Docker).

### 2. TUI Mode

```bash
./panel tui
```

Appending `tui` will launch the panel with the interactive console UI enabled. If you start the panel with the TUI, you can safely exit it at any time by typing `exit` or hitting `Ctrl+C` / `Esc` at the main prompt. **Exiting the TUI shuts down the panel.**

## Features

The TUI provides native console access to almost all administrative capabilities available in the frontend interface:

- **Live Activity Logs**: View a streaming feed of the global activity logs, showing all user actions, logins, and system events in real-time.
- **User Management (`user`)**:
  - List all registered users.
  - View detailed user information (email, 2FA status, creation dates).
  - Modify user resource limits (Server Count, CPU, Memory, Disk, Allocations).
  - Execute administrative actions (Reset Password, Require 2FA, Ban IPs, Delete User, Delete All Servers).
- **Server Management (`server`)**:
  - List all servers and view their current installation status.
  - View server details (Node, Owner, Package, UUID).
  - Modify server resource limits.
  - Execute administrative actions (Change Owner, Reinstall, Suspend/Unsuspend, Delete).
  - **Create Servers**: Interactively create new servers, prompting for Name, Owner UUID, Node UUID, Package UUID, and resource limits.
- **Node Management (`node`)**:
  - List all Daemon nodes and view their online/offline status.
  - Inspect live system statistics retrieved from Axis (CPU cores/usage, Memory usage, Disk usage, OS details).
  - **Node Creation**: Manually register a new node or use the automated **Pairing Flow**.
- **Database Hosts (`dbhost`)**:
  - List existing MySQL database hosts.
  - View attached active databases count and connection details.
  - Register new database hosts directly from the terminal.
- **Mounts (`mount`)**:
  - View and configure global host-to-container directory mounts.
  - Specify source paths, target paths, and configure user-mountable or read-only states.