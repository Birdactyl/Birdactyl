# Installation

This guide covers the system requirements and installation process for Birdactyl.

## System Requirements

### Panel Server
- Go 1.24 or higher
- PostgreSQL, MySQL, or SQLite
- 1GB RAM minimum (2GB+ recommended)
- Linux, macOS, or Windows

### Node Server (Axis)
- Go 1.24 or higher
- Docker (Axis will auto-install if not present on Linux)
- Linux only (Docker container management)
- Root access for initial setup
- Sufficient disk space for server data

### Frontend Build
- Node.js 18+ or Bun
- npm or yarn

## Quick Installation

### 1. Clone the Repository

```bash
git clone https://github.com/Birdactyl/Birdactyl.git
cd Birdactyl
```

### 2. Build the Panel

```bash
cd server
go build -o panel
./panel
```

On first run, a `config.yaml` file is generated. Configure your database settings and restart.

### 3. Build Axis (Node Daemon)

```bash
cd axis
go build -o axis
sudo ./axis
```

On first run, a `config.yaml` file is generated. See [Axis Setup](axis-setup.md) for pairing instructions.

### 4. Build the Frontend

```bash
cd client
npm install
npm run build
```

The built files will be in `client/dist/`. Serve these with your web server with a proxy to the panel.

## Directory Structure

Both panel and axis use relative paths from wherever you run them. A typical setup:

```
server/
  panel             # Panel binary
  config.yaml       # Panel configuration (generated on first run)
  plugins/          # Plugin directory
  logs/panel.log    # Log file

axis/
  axis              # Axis binary
  config.yaml       # Axis configuration (generated on first run)
  logs/axis.log     # Log file
```

Axis stores server data in `/var/lib/birdactyl/` by default (configurable):
- `servers/` - Server data volumes
- `backups/` - Backup storage

## Next Steps

- [Panel Setup](panel-setup.md) - Configure the backend
- [Database Setup](database-setup.md) - Set up your database
- [Axis Setup](axis-setup.md) - Install and pair nodes
- [Frontend Setup](frontend-setup.md) - Deploy the client
