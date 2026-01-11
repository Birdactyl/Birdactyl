# Frontend Setup

The Birdactyl frontend is a React application built with TypeScript and Tailwind CSS.

## Requirements

- Node.js 18+ or Bun
- npm, yarn, or bun package manager

## Building

### With npm

```bash
cd client
npm install
npm run build
```

### With Bun

```bash
cd client
bun install
bun run build
```

## Output

Built files are placed in `client/dist/`. This directory contains:

- `index.html` - Entry point
- `assets/` - JavaScript, CSS, and other assets

## Development Server

For development with hot reload:

```bash
npm run dev
```

This starts a development server at `http://localhost:5173` by default.

## Deployment

### Option 1: Serve with Nginx

```nginx
server {
    listen 80;
    server_name panel.example.com;
    root /var/www/birdactyl;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

Copy built files:

```bash
sudo cp -r client/dist/* /var/www/birdactyl/
```

### Option 2: Serve with Caddy

```
panel.example.com {
    root * /var/www/birdactyl
    file_server
    try_files {path} /index.html

    handle /api/* {
        reverse_proxy localhost:3000
    }
}
```

### Option 3: Single Origin Setup

Configure Vite to proxy API requests during development. Edit `vite.config.ts`:

```typescript
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
        ws: true,
      },
    },
  },
})
```

## Tech Stack

- React 18
- React Router for navigation
- Tailwind CSS for styling
- Monaco Editor for file editing
- Vite for building
