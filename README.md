# Birdactyl

A modern game server management panel built in Go and React. Fast, extensible, and plugin-friendly.

## Architecture

- **Panel** - Go backend (Fiber) handling auth, database, API, and plugins
- **Axis** - Go node daemon managing Docker containers
- **Client** - React + TypeScript + Tailwind frontend

## Supported Games

Minecraft (Paper, Purpur, Fabric, Forge, NeoForge, Velocity, BungeeCord, Bedrock), Node.js, Bun, Python

## Quick Start

```bash
cd server && go build -o panel && ./panel
cd axis && go build -o axis && sudo ./axis pair
cd client && npm install && npm run build
```

Configure `config.yaml` in each component after first run.

## Requirements

- Go 1.24+
- PostgreSQL/MySQL/SQLite
- Docker (on nodes, if not present axis will install and set it up.)
- Node.js 18+ or Bun (for client)

## Configuration

### Panel Configuration (`server/config.yaml`)

#### Server

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `server.host` | string | `0.0.0.0` | IP address the panel listens on |
| `server.port` | int | `3000` | Port the panel listens on |

#### Logging

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `logging.file` | string | `logs/panel.log` | Path to the log file |

#### Database

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `database.driver` | string | `postgres` | Database driver: `postgres`, `mysql`, or `sqlite` |
| `database.host` | string | `localhost` | Database host address |
| `database.port` | int | `5432` | Database port |
| `database.user` | string | - | Database username |
| `database.password` | string | - | Database password |
| `database.name` | string | `birdactyl` | Database name (for SQLite, must end with `.db`) |
| `database.sslmode` | string | `disable` | SSL mode: `disable`, `require`, `verify-ca`, `verify-full` |
| `database.max_open_conns` | int | `25` | Maximum open database connections |
| `database.max_idle_conns` | int | `5` | Maximum idle database connections |
| `database.conn_max_lifetime` | int | `300` | Connection max lifetime in seconds |

#### Authentication

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `auth.accounts_per_ip` | int | `3` | Maximum accounts allowed per IP address |
| `auth.access_token_expiry` | int | `15` | Access token expiry time in minutes |
| `auth.refresh_token_expiry` | int | `43200` | Refresh token expiry time in minutes (30 days) |
| `auth.token_refresh_threshold` | int | `1` | Minutes before expiry when token refresh is allowed |
| `auth.max_sessions_per_user` | int | `25` | Maximum concurrent sessions per user |
| `auth.jwt_secret` | string | - | Secret key for JWT signing (generate a secure random string) |
| `auth.bcrypt_cost` | int | `12` | Bcrypt hashing cost (10-14 recommended) |

#### Root Admins

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `root_admins` | []string | `[]` | List of user UUIDs with root admin privileges |

#### Addon Sources
#### API Keys

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `api_keys.<name>.key` | string | - | The API key value |
| `api_keys.<name>.headers` | map | - | HTTP headers to send with requests. Use `{{key}}` to interpolate the key value |

Example:
```yaml
api_keys:
  curseforge:
    key: "your-curseforge-api-key"
    headers:
      x-api-key: "{{key}}"
      Accept: "application/json"
  custom-api:
    key: "your-api-key"
    headers:
      Authorization: "Bearer {{key}}"
```

#### Resources

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `resources.enabled` | bool | `true` | Enable resource limits for users |
| `resources.default_ram` | int | `4096` | Default RAM allocation in MB |
| `resources.default_cpu` | int | `200` | Default CPU allocation (100 = 1 core) |
| `resources.default_disk` | int | `10240` | Default disk allocation in MB |
| `resources.max_servers` | int | `3` | Maximum servers per user |

#### Plugins

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `plugins.address` | string | `localhost:50050` | gRPC address for plugin communication |
| `plugins.directory` | string | `plugins` | Directory containing plugin binaries |
| `plugins.allow_dynamic` | bool | `true` | Allow dynamic plugin loading at runtime |

---

### Axis (Node) Configuration (`axis/config.yaml`)

#### Panel Connection

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `panel.url` | string | - | URL of the Birdactyl panel |
| `panel.token` | string | - | Authentication token (generated during pairing) |

#### Node Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `node.listen` | string | `0.0.0.0:8443` | Address and port the node daemon listens on |
| `node.data_dir` | string | `/var/lib/birdactyl` | Directory for server data storage |
| `node.backup_dir` | string | `/var/lib/birdactyl/backups` | Directory for server backups |
| `node.display_ip` | string | - | Public IP/hostname shown to users |

#### Logging

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `logging.file` | string | `logs/axis.log` | Path to the log file |
