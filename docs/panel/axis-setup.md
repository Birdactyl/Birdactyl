# Axis Setup

Axis is the node daemon that runs on each server host and manages Docker containers for game servers.

## Requirements

- Linux operating system
- Root access for initial setup
- Docker (auto-installed if not present)
- Network connectivity to the panel

## Building

```bash
cd axis
go build -o axis
```

## First Run

```bash
sudo ./axis
```

On first run, Axis:
1. Generates a default `config.yaml`
2. Creates the `birdactyl` system user
3. Sets up data directories
4. Initializes Docker (installs if needed)

## Configuration

Edit `config.yaml`:

```yaml
panel:
  url: "http://panel.example.com:3000"
  token: ""

node:
  listen: "0.0.0.0:8443"
  data_dir: "/var/lib/birdactyl/servers"
  backup_dir: "/var/lib/birdactyl/backups"
  display_ip: "your.public.ip"

logging:
  file: "logs/axis.log"
```

| Option | Description |
|--------|-------------|
| `panel.url` | URL of your Birdactyl panel |
| `panel.token` | Authentication token (set during pairing) |
| `node.listen` | Address and port Axis listens on |
| `node.data_dir` | Directory for server data |
| `node.backup_dir` | Directory for backups |
| `node.display_ip` | Public IP shown to users |

## Pairing with Panel

### Method 1: Pairing Mode (Recommended)

1. Start Axis in pairing mode:
```bash
sudo ./axis pair
```

2. In the panel admin area, create a new node and initiate pairing

3. Axis displays the pairing request with a verification code:
```
========================================
  PAIRING REQUEST RECEIVED
========================================
  Panel URL: http://panel.example.com:3000
  Code: abc123
========================================

Does this match what you see in the panel?
Accept pairing? [y/N]:
```

4. Verify the code matches and type `y` to accept

5. The token is automatically saved to `config.yaml`

### Method 2: Manual Token

1. Create a node in the panel admin area
2. Copy the generated token
3. Add it to `config.yaml`:
```yaml
panel:
  token: "tokenid.tokensecret"
```

## Running

```bash
sudo ./axis
```

Expected output:
```
[INFO] Starting Cauthon Axis...
[INFO] Loaded token: abc123def4...
[SUCCESS] Docker ready
[SUCCESS] Connected to panel
[INFO] API server listening on 0.0.0.0:8443
```

## Running as a Service

Create `/etc/systemd/system/birdactyl-axis.service` (adjust paths to where you placed the binary):

```ini
[Unit]
Description=Birdactyl Axis Node Daemon
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
WorkingDirectory=/path/to/axis
ExecStart=/path/to/axis/axis
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable birdactyl-axis
sudo systemctl start birdactyl-axis
```

## Data Directories

Axis creates and manages:

- `/var/lib/birdactyl/servers/` - Server data volumes
- `/var/lib/birdactyl/backups/` - Backup storage

These directories need write permissions. Running with `sudo` on first start sets up proper permissions.

## Docker

Axis uses Docker to run game servers in isolated containers. If Docker is not installed, Axis attempts to install it automatically on supported distributions:

- Ubuntu/Debian (apt)
- Fedora (dnf)
- CentOS/RHEL (yum)

For other distributions, install Docker manually before running Axis.

## Heartbeat

Axis sends heartbeats to the panel every 30 seconds to report node status. If heartbeats fail, check:

- Network connectivity to panel
- Panel URL in config
- Token validity

## Firewall

Open the following ports:

- `8443` (or your configured listen port) - API communication with panel
- Ports for game servers as allocated
