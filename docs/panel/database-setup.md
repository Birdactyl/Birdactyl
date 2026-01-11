# Database Setup

Birdactyl supports PostgreSQL, MySQL, and SQLite as database backends.

## PostgreSQL (Recommended)

### Installation

Ubuntu/Debian:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

CentOS/RHEL:
```bash
sudo dnf install postgresql-server postgresql-contrib
sudo postgresql-setup --initdb
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

### Create Database

```bash
sudo -u postgres psql
```

```sql
CREATE USER birdactyl WITH PASSWORD 'your-secure-password';
CREATE DATABASE birdactyl OWNER birdactyl;
GRANT ALL PRIVILEGES ON DATABASE birdactyl TO birdactyl;
\q
```

### Configuration

```yaml
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  user: "birdactyl"
  password: "your-secure-password"
  name: "birdactyl"
  sslmode: "disable"
```

SSL modes:
- `disable` - No SSL
- `require` - SSL required, no verification
- `verify-ca` - SSL with CA verification
- `verify-full` - SSL with full verification

## MySQL / MariaDB

### Installation

Ubuntu/Debian:
```bash
sudo apt update
sudo apt install mysql-server
```

### Create Database

```bash
sudo mysql
```

```sql
CREATE DATABASE birdactyl;
CREATE USER 'birdactyl'@'localhost' IDENTIFIED BY 'your-secure-password';
GRANT ALL PRIVILEGES ON birdactyl.* TO 'birdactyl'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### Configuration

```yaml
database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  user: "birdactyl"
  password: "your-secure-password"
  name: "birdactyl"
```

## SQLite

SQLite requires no external database server. Useful for development or small deployments.

### Configuration

```yaml
database:
  driver: "sqlite"
  name: "birdactyl.db"
```

The database file is created in the panel's working directory.

## Connection Pool Settings

Configure connection pooling for production:

```yaml
database:
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300
```

| Option | Default | Description |
|--------|---------|-------------|
| `max_open_conns` | 25 | Maximum open connections |
| `max_idle_conns` | 5 | Maximum idle connections |
| `conn_max_lifetime` | 300 | Connection lifetime in seconds |

## Auto Migration

The panel automatically creates and updates database tables on startup. No manual migration is required.

Tables created:
- `users` - User accounts
- `sessions` - Active sessions
- `ip_registrations` - IP registration tracking
- `nodes` - Node registrations
- `packages` - Server packages
- `servers` - Server instances
- `activity_logs` - Activity history
- `ip_bans` - IP ban list
- `settings` - Panel settings
- `subusers` - Server subusers
- `database_hosts` - External database hosts
- `server_databases` - Server databases
- `schedules` - Scheduled tasks
- `api_keys` - API keys

## Backup

### PostgreSQL

```bash
pg_dump -U birdactyl birdactyl > backup.sql
```

### MySQL

```bash
mysqldump -u birdactyl -p birdactyl > backup.sql
```

### SQLite

```bash
cp birdactyl.db backup.db
```
