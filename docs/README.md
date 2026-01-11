# Birdactyl Documentation

Birdactyl is a modern game server management panel built in Go and React. This documentation covers both the panel installation/setup and the plugin system.

## Panel Documentation

- [Installation](panel/installation.md) - System requirements and installation guide
- [Panel Setup](panel/panel-setup.md) - Backend configuration and deployment
- [Axis Setup](panel/axis-setup.md) - Node daemon installation and pairing
- [Frontend Setup](panel/frontend-setup.md) - Building and deploying the client
- [Database Setup](panel/database-setup.md) - Database configuration options
- [Configuration Reference](panel/configuration.md) - Complete configuration options

## Plugin System Documentation

- [Getting Started](plugins/getting-started.md) - Set up your first plugin
- [Go SDK](plugins/go-sdk.md) - Build plugins with Go
- [Java SDK](plugins/java-sdk.md) - Build plugins with Java
- [UI](plugins/ui.md) - Add custom pages, tabs, and sidebar items
- [Events](plugins/events.md) - React to panel events
- [Routes](plugins/routes.md) - Add custom HTTP endpoints
- [Mixins](plugins/mixins.md) - Intercept and modify panel operations
- [Schedules](plugins/schedules.md) - Run tasks on a cron schedule
- [Panel API](plugins/panel-api.md) - Interact with servers, users, files, and more
- [Addon Types](plugins/addon-types.md) - Define custom addon installation handlers
- [Configuration](plugins/configuration.md) - Hot-reloadable config files

## Architecture Overview

Birdactyl consists of three main components:

- **Panel** (server/) - Go backend using Fiber framework, handles authentication, database operations, API endpoints, and plugin management
- **Axis** (axis/) - Go node daemon that manages Docker containers on host machines
- **Client** (client/) - React + TypeScript + Tailwind frontend

Plugins run as separate processes and communicate with the panel via gRPC, allowing for isolated and extensible functionality.
