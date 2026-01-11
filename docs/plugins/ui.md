# Plugin UI System

Plugins can add custom user interfaces to the panel, including pages, tabs, and sidebar items. The UI is built with React and TypeScript using the `@birdactyl/plugin-ui` SDK.

## Overview

The plugin UI system consists of:

1. **Backend (Go/Java)** - Declares what UI elements your plugin provides
2. **Frontend (React)** - The actual UI components bundled as JavaScript
3. **SDK** - Provides React components, hooks, and utilities

## Project Structure

A plugin with UI typically has this structure:

```
my-plugin/
  main.go              # Plugin backend
  go.mod
  dist/
    bundle.js          # Built UI bundle (embedded in binary)
  ui/
    src/
      index.tsx        # Entry point - exports all components
      pages/
        DashboardPage.tsx
        SettingsPage.tsx
      tabs/
        ServerTab.tsx
    package.json
    vite.config.ts
    tsconfig.json
```

## Backend Setup (Go)

### Declaring UI Elements

Use the `UI()` builder to declare your plugin's UI:

```go
package main

import (
    "embed"
    birdactyl "github.com/Birdactyl/Birdactyl-Go-SDK"
)

//go:embed dist/bundle.js
var uiBundle embed.FS

func main() {
    plugin := birdactyl.New("my-plugin", "1.0.0").
        SetName("My Plugin")

    plugin.UI().
        HasBundle().
        EmbedBundle(uiBundle, "dist/bundle.js").
        Page("/", "DashboardPage").Title("Dashboard").Icon("home").Done().
        Page("/settings", "SettingsPage").Title("Settings").Icon("settings").Done().
        Tab("server-tab", "ServerTab", birdactyl.TabTargetServer, "My Plugin").
            Icon("puzzle").Order(100).Done().
        Tab("settings-tab", "SettingsTab", birdactyl.TabTargetUserSettings, "My Plugin").
            Icon("puzzle").Order(10).Done().
        SidebarItem("my-plugin", "My Plugin", "/console/plugins/my-plugin", birdactyl.SidebarSectionPlatform).
            Icon("puzzle").Order(50).
            Child("Dashboard", "/console/plugins/my-plugin").
            Child("Settings", "/console/plugins/my-plugin/settings").
            Done()

    plugin.Start("localhost:50050")
}
```

### Pages

Pages are full-page views accessible via URL:

```go
plugin.UI().
    Page("/", "DashboardPage").Title("Dashboard").Icon("home").Done().
    Page("/settings", "SettingsPage").Title("Settings").Icon("settings").Done().
    Page("/admin", "AdminPage").Title("Admin").Icon("shield").AdminOnly().Done().
    Page("/vip", "VIPPage").Title("VIP").Icon("star").Guard("vip").Done()
```

Pages are served at `/console/plugins/{plugin-id}/{path}`.

| Method | Description |
|--------|-------------|
| `Page(path, component)` | Create a page at path using the named component |
| `Title(title)` | Set the page title |
| `Icon(icon)` | Set the page icon |
| `AdminOnly()` | Restrict to admin users |
| `Guard(guard)` | Use a custom guard (evaluated by your `evaluateGuard` function) |
| `Done()` | Return to the UI builder |

### Tabs

Tabs inject into existing panel pages:

```go
plugin.UI().
    Tab("server-tab", "ServerTab", birdactyl.TabTargetServer, "My Plugin").
        Icon("puzzle").Order(100).Done().
    Tab("settings-tab", "SettingsTab", birdactyl.TabTargetUserSettings, "Plugin Settings").
        Icon("settings").Order(10).Done()
```

| Target | Description |
|--------|-------------|
| `TabTargetServer` | Adds a tab to the server detail page |
| `TabTargetUserSettings` | Adds a tab to the user settings page |

| Method | Description |
|--------|-------------|
| `Tab(id, component, target, label)` | Create a tab |
| `Icon(icon)` | Set the tab icon |
| `Order(order)` | Set sort order (lower = earlier) |
| `Done()` | Return to the UI builder |

### Sidebar Items

Add items to the panel sidebar:

```go
plugin.UI().
    SidebarItem("my-plugin", "My Plugin", "/console/plugins/my-plugin", birdactyl.SidebarSectionPlatform).
        Icon("puzzle").Order(50).
        Child("Dashboard", "/console/plugins/my-plugin").
        Child("Settings", "/console/plugins/my-plugin/settings").
        Done().
    SidebarItem("my-plugin-admin", "Admin", "/console/plugins/my-plugin/admin", birdactyl.SidebarSectionAdmin).
        Icon("shield").AdminOnly().Done()
```

| Section | Description |
|---------|-------------|
| `SidebarSectionNav` | Main navigation section |
| `SidebarSectionPlatform` | Platform section (below servers) |
| `SidebarSectionAdmin` | Admin section (admin users only) |

| Method | Description |
|--------|-------------|
| `SidebarItem(id, label, href, section)` | Create a sidebar item |
| `Icon(icon)` | Set the icon |
| `Order(order)` | Set sort order |
| `AdminOnly()` | Restrict to admin users |
| `Child(label, href)` | Add a child link |
| `Done()` | Return to the UI builder |

## Frontend Setup

### Package.json

```json
{
  "name": "my-plugin-ui",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "vite build"
  },
  "dependencies": {
    "@birdactyl/plugin-ui": "^1.0.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

### Vite Config

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [
    react({
      jsxRuntime: 'classic',
    }),
  ],
  build: {
    lib: {
      entry: 'src/index.tsx',
      formats: ['iife'],
      name: 'PluginBundle',
      fileName: () => 'bundle.js',
    },
    outDir: '../dist',
    emptyOutDir: true,
    rollupOptions: {
      external: ['react', 'react-dom', '@birdactyl/plugin-ui'],
      output: {
        globals: {
          'react': 'window.BIRDACTYL.React',
          'react-dom': 'window.BIRDACTYL.ReactDOM',
          '@birdactyl/plugin-ui': 'window.BIRDACTYL_SDK',
        },
      },
    },
  },
});
```

### Entry Point (index.tsx)

Export all your components and optionally a guard evaluator:

```tsx
export { default as DashboardPage } from './pages/DashboardPage';
export { default as SettingsPage } from './pages/SettingsPage';
export { default as ServerTab } from './tabs/ServerTab';

export function evaluateGuard(guard: string, user: { id: string; username: string; is_admin: boolean } | null): boolean {
  if (!user) return false;
  if (guard === 'vip') {
    return ['admin', 'vip_user'].includes(user.username) || user.is_admin;
  }
  return false;
}
```

## SDK Reference

### Hooks

```tsx
import { usePluginAPI, useState, useEffect, useEvent } from '@birdactyl/plugin-ui';

function MyComponent() {
  const api = usePluginAPI();
  const [data, setData] = useState(null);

  useEvent('server:start', (event) => {
    console.log('Server started:', event.serverId);
  });

  useEffect(() => {
    api.get('/my-data').then(setData);
  }, []);
}
```

| Hook | Description |
|------|-------------|
| `usePluginAPI()` | Access the plugin API |
| `useState(initial)` | React useState |
| `useEffect(fn, deps)` | React useEffect |
| `useCallback(fn, deps)` | React useCallback |
| `useMemo(fn, deps)` | React useMemo |
| `useRef(initial)` | React useRef |
| `useEvent(event, callback)` | Subscribe to panel events |

### Plugin API

```tsx
const api = usePluginAPI();

const data = await api.get<MyType>('/endpoint');
await api.post('/endpoint', { key: 'value' });
await api.put('/endpoint', { key: 'value' });
await api.delete('/endpoint');

api.notify('Title', 'Message', 'success');
api.notify('Error', 'Something went wrong', 'error');
api.notify('Info', 'FYI', 'info');

const user = api.getUser();
const isAdmin = api.isAdmin();

api.navigate('/console/servers');
```

API requests are automatically prefixed with `/api/v1/plugins/{plugin-id}`.

### Components

```tsx
import {
  Button,
  Card,
  StatCard,
  Input,
  Modal,
  Table,
  Checkbox,
  DatePicker,
  Pagination,
  BulkActionBar,
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  Icons,
} from '@birdactyl/plugin-ui';
```

#### Button

```tsx
<Button onClick={handleClick}>Click Me</Button>
<Button variant="secondary">Secondary</Button>
<Button variant="danger">Delete</Button>
<Button variant="ghost">Ghost</Button>
<Button loading={isLoading}>Submit</Button>
```

#### Card

```tsx
<Card title="My Card" description="Optional description">
  <p>Card content</p>
</Card>
```

#### StatCard

```tsx
<StatCard label="Servers" value={42} icon="server" />
<StatCard label="Memory" value="2.5 GB" max="of 4 GB" percent={62.5} icon="pieChart" />
```

#### Input

```tsx
<Input label="Username" value={value} onChange={e => setValue(e.target.value)} />
<Input label="Password" hideable />
<Input placeholder="Search..." />
```

#### Modal

```tsx
<Modal open={isOpen} onClose={() => setIsOpen(false)} title="Confirm" description="Are you sure?">
  <Button onClick={handleConfirm}>Yes</Button>
  <Button variant="ghost" onClick={() => setIsOpen(false)}>Cancel</Button>
</Modal>
```

#### Table

```tsx
const columns = [
  { key: 'name', header: 'Name', render: (item) => item.name },
  { key: 'status', header: 'Status', render: (item) => item.status },
];

<Table columns={columns} data={items} keyField="id" />
```

#### Checkbox

```tsx
<Checkbox checked={enabled} onChange={setEnabled} label="Enable feature" />
```

#### DropdownMenu

```tsx
<DropdownMenu>
  <DropdownMenuTrigger>
    <Button>Options</Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent>
    <DropdownMenuItem onSelect={() => handleEdit()}>Edit</DropdownMenuItem>
    <DropdownMenuItem onSelect={() => handleDelete()} destructive>Delete</DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
```

#### Icons

```tsx
<Icons.server className="w-4 h-4" />
<Icons.users className="w-4 h-4" />
<Icons.settings className="w-4 h-4" />
```

### Events

Subscribe to panel events:

```tsx
import { useEvent, events } from '@birdactyl/plugin-ui';

useEvent('server:start', (data) => {
  console.log('Server started:', data.serverId);
});

useEvent('file:saved', (data) => {
  console.log('File saved:', data.path);
});

events.emit('plugin:my-plugin:custom-event', { foo: 'bar' });
```

| Event | Data |
|-------|------|
| `server:status` | `{ serverId, status, previousStatus }` |
| `server:stats` | `{ serverId, memory, memoryLimit, cpu, disk }` |
| `server:log` | `{ serverId, line }` |
| `server:start` | `{ serverId }` |
| `server:stop` | `{ serverId }` |
| `server:restart` | `{ serverId }` |
| `server:kill` | `{ serverId }` |
| `file:created` | `{ serverId, path }` |
| `file:deleted` | `{ serverId, path }` |
| `file:moved` | `{ serverId, from, to }` |
| `file:uploaded` | `{ serverId, path }` |
| `file:saved` | `{ serverId, path }` |
| `navigation` | `{ path, previousPath }` |
| `user:login` | `{ userId, username }` |
| `user:logout` | `{}` |
| `plugin:*` | Custom plugin events |

## Building

1. Build the UI:
```bash
cd ui
npm install
npm run build
```

2. Build the plugin (embeds the bundle):
```bash
go build -o my-plugin
```

3. Copy to plugins directory:
```bash
cp my-plugin /path/to/panel/plugins/
```

## Example Components

### Page Component

```tsx
import { usePluginAPI, useState, useEffect, Card, Button, StatCard } from '@birdactyl/plugin-ui';

interface Stats {
  servers: number;
  users: number;
}

export default function DashboardPage() {
  const api = usePluginAPI();
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.get<Stats>('/stats').then(setStats).finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-neutral-100">Dashboard</h1>
      
      <div className="grid grid-cols-2 gap-4">
        <StatCard label="Servers" value={stats?.servers ?? '-'} icon="server" />
        <StatCard label="Users" value={stats?.users ?? '-'} icon="users" />
      </div>
      
      <Card title="About">
        <p className="text-neutral-300">Welcome to my plugin!</p>
      </Card>
    </div>
  );
}
```

### Tab Component

Server tabs receive `serverId` and `server` as props:

```tsx
import { usePluginAPI, Card, Button } from '@birdactyl/plugin-ui';

interface Props {
  serverId: string;
  server: { id: string; name: string; status: string };
}

export default function ServerTab({ serverId, server }: Props) {
  const api = usePluginAPI();

  const handleAction = () => {
    api.post('/action', { serverId }).then(() => {
      api.notify('Success', 'Action completed', 'success');
    });
  };

  return (
    <Card title="My Plugin" description={`Server: ${server.name}`}>
      <Button onClick={handleAction}>Run Action</Button>
    </Card>
  );
}
```

User settings tabs receive `user` as a prop:

```tsx
import { usePluginAPI, Card, Input, Button, useState } from '@birdactyl/plugin-ui';

interface Props {
  user: { id: string; username: string; email: string };
}

export default function UserSettingsTab({ user }: Props) {
  const api = usePluginAPI();
  const [setting, setSetting] = useState('');

  const handleSave = async () => {
    await api.post('/user-settings', { setting });
    api.notify('Saved', 'Settings updated', 'success');
  };

  return (
    <Card title="Plugin Settings">
      <Input label="Setting" value={setting} onChange={e => setSetting(e.target.value)} />
      <Button onClick={handleSave}>Save</Button>
    </Card>
  );
}
```
