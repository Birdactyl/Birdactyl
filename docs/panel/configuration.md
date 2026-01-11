# Configuration Reference

Complete configuration options for the Birdactyl panel and Axis node daemon.

## Panel Configuration

Located at `server/config.yaml`.

### Server

```yaml
server:
  host: "0.0.0.0"
  port: 3000
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `host` | string | `0.0.0.0` | IP address to bind |
| `port` | int | `3000` | Port to listen on |

### Logging

```yaml
logging:
  file: "logs/panel.log"
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `file` | string | `logs/panel.log` | Log file path |

### Database

```yaml
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: ""
  name: "birdactyl"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `driver` | string | `postgres` | `postgres`, `mysql`, or `sqlite` |
| `host` | string | `localhost` | Database host |
| `port` | int | `5432` | Database port |
| `user` | string | - | Database username |
| `password` | string | - | Database password |
| `name` | string | `birdactyl` | Database name |
| `sslmode` | string | `disable` | SSL mode (postgres only) |
| `max_open_conns` | int | `25` | Max open connections |
| `max_idle_conns` | int | `5` | Max idle connections |
| `conn_max_lifetime` | int | `300` | Connection lifetime (seconds) |

### Authentication

```yaml
auth:
  accounts_per_ip: 3
  access_token_expiry: 15
  refresh_token_expiry: 43200
  token_refresh_threshold: 1
  max_sessions_per_user: 5
  bcrypt_cost: 12
  jwt_secret: ""
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `accounts_per_ip` | int | `3` | Max accounts per IP |
| `access_token_expiry` | int | `15` | Access token expiry (minutes) |
| `refresh_token_expiry` | int | `43200` | Refresh token expiry (minutes) |
| `token_refresh_threshold` | int | `1` | Minutes before expiry to allow refresh |
| `max_sessions_per_user` | int | `5` | Max concurrent sessions |
| `bcrypt_cost` | int | `12` | Bcrypt hashing cost |
| `jwt_secret` | string | auto | JWT signing secret (auto-generated if empty) |

### Resources

```yaml
resources:
  enabled: true
  default_ram: 4096
  default_cpu: 200
  default_disk: 10240
  max_servers: 3
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable resource limits |
| `default_ram` | int | `4096` | Default RAM (MB) |
| `default_cpu` | int | `200` | Default CPU (100 = 1 core) |
| `default_disk` | int | `10240` | Default disk (MB) |
| `max_servers` | int | `3` | Max servers per user |

### Plugins

```yaml
plugins:
  address: "localhost:50050"
  directory: "plugins"
  allow_dynamic: true
  container:
    enabled: false
    image: "birdactyl/plugin-runtime:latest"
    network_mode: "host"
    memory_limit: "512m"
    cpu_limit: "1.0"
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `address` | string | `localhost:50050` | gRPC server address |
| `directory` | string | `plugins` | Plugin directory |
| `allow_dynamic` | bool | `true` | Allow runtime plugin loading |
| `container.enabled` | bool | `false` | Run plugins in containers |
| `container.image` | string | - | Container image |
| `container.network_mode` | string | `host` | Docker network mode |
| `container.memory_limit` | string | `512m` | Memory limit |
| `container.cpu_limit` | string | `1.0` | CPU limit |

### Root Admins

```yaml
root_admins:
  - "user-uuid-1"
  - "user-uuid-2"
```

User UUIDs listed here have permanent root admin access.

### API Keys

```yaml
api_keys:
  curseforge:
    key: "your-api-key"
    headers:
      x-api-key: "{{key}}"
```

External API keys for addon sources. Use `{{key}}` to interpolate the key value in headers.

## Axis Configuration

Located at `axis/config.yaml`.

### Panel Connection

```yaml
panel:
  url: "http://localhost:3000"
  token: ""
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `url` | string | - | Panel URL |
| `token` | string | - | Authentication token |

### Node Settings

```yaml
node:
  listen: "0.0.0.0:8443"
  data_dir: "/var/lib/birdactyl/servers"
  backup_dir: "/var/lib/birdactyl/backups"
  display_ip: ""
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `listen` | string | `0.0.0.0:8443` | Listen address |
| `data_dir` | string | `/var/lib/birdactyl/servers` | Server data directory |
| `backup_dir` | string | `/var/lib/birdactyl/backups` | Backup directory |
| `display_ip` | string | - | Public IP for users |

### Logging

```yaml
logging:
  file: "logs/axis.log"
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `file` | string | `logs/axis.log` | Log file path |

## Environment Variables

The panel supports environment variable overrides:

| Variable | Overrides |
|----------|-----------|
| `DB_HOST` | `database.host` |
| `DB_USER` | `database.user` |
| `DB_PASSWORD` | `database.password` |
| `DB_NAME` | `database.name` |
