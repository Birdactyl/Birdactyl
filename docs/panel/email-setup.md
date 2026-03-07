# Email Setup

Birdactyl uses SMTP to send transactional emails, including password resets and account verification codes.

## SMTP Configuration

Email settings are located in `server/config.yaml` under the `smtp` section.

```yaml
smtp:
  enabled: true
  host: "smtp.gmail.com"
  port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  from_email: "noreply@example.com"
  from_name: "Birdactyl"
```

| Option | Description |
|--------|-------------|
| `enabled` | Must be `true` to send emails |
| `host` | SMTP server address |
| `port` | SMTP port (typically 587 for TLS or 465 for SSL) |
| `username` | SMTP username (usually your email address) |
| `password` | SMTP password or app-specific password |
| `from_email` | The "From" address shown to recipients |
| `from_name` | The name shown as the sender |

## Email Verification

You can require users to verify their email addresses before they can perform certain actions.

### Enabling Verification

Verification settings can be managed through the admin panel under **Settings > Security**.

Alternatively, you can use the Panel API (for plugins):

```go
// Set email verification enabled
api.UpdateSettings(/* currently handled via specialized routes */)
```

### Restrictions

You can configure what actions are blocked for unverified users:
- `auth.login` - Block login entirely (requires verification after registration)
- `server.create` - Block server creation
- `server.control` - Block power actions
- `file.modify` - Block file edits

## Testing

To test your email configuration, you can use the "Send Test Email" button in the admin settings or use the Panel API from a plugin:

```go
api.SendEmail("test@example.com", "Hello", "<h1>Test Email</h1><p>It works!</p>")
```
