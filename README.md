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
