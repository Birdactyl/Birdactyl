# Panel Setup

The panel is the central backend that handles authentication, database operations, API endpoints, and plugin management.

## Building

```bash
cd server
go build -o panel
```

## First Run

```bash
./panel
```

On first run, the panel generates a default `config.yaml` and exits. Configure your database settings before restarting.

## Configuration

Edit `config.yaml` with your settings:

```yaml
server:
  host: "0.0.0.0"
  port: 3000

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your-password"
  name: "birdactyl"
  sslmode: "disable"

auth:
  accounts_per_ip: 3
  bcrypt_cost: 12

plugins:
  address: "localhost:50050"
  directory: "plugins"
```

See [Configuration Reference](configuration.md) for all options.

## Running

```bash
./panel
```

## Running as a Service

Create `/etc/systemd/system/birdactyl-panel.service` (adjust paths to where you placed the binary):

```ini
[Unit]
Description=Birdactyl Panel
After=network.target

[Service]
Type=simple
WorkingDirectory=/path/to/panel
ExecStart=/path/to/panel/panel
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable birdactyl-panel
sudo systemctl start birdactyl-panel
```

## Reverse Proxy

### Nginx

```nginx
server {
    listen 80;
    server_name panel.example.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy

```
panel.example.com {
    reverse_proxy localhost:3000
}
```

## Root Admin Setup

The first user to register becomes a root admin user.

To grant root admin privileges, add user's UUID to `config.yaml`:

```yaml
root_admins:
  - "user-uuid-here"
```

Root admins have full access to all panel functionality and cannot have their admin status revoked through the UI.

## Logging

Logs are written to `logs/panel.log` by default. Configure the path in `config.yaml`:

```yaml
logging:
  file: "logs/panel.log"
```

## Environment Variables

Database settings can be overridden with environment variables:

- `DB_HOST` - Database host
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
