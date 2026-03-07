# Security & Two-Factor Authentication

Birdactyl provides several security layers to protect user accounts, including Two-Factor Authentication (2FA) via TOTP.

## Two-Factor Authentication (2FA)

Users can enable 2FA on their profile settings page. Birdactyl uses the standard TOTP protocol, compatible with apps like Google Authenticator, Authy, and 1Password.

### Enablement Flow
1. User requests 2FA setup.
2. Panel generates a TOTP secret and QR code.
3. User scans the QR code and enters the 6-digit verification code.
4. Panel verifies the code and provides a set of **Backup Codes**.

### Backup Codes
When 2FA is enabled, the system generates 8 one-time-use backup codes. These allow users to regain access if they lose their TOTP device. 

> [!IMPORTANT]
> Users should be encouraged to save their backup codes in a safe place.

## Administrative Controls

Admins can manage 2FA for any user through the User Management section.

### Disabling 2FA
If a user loses both their device and backup codes, an admin can manually disable 2FA for their account.

**Via Panel API (Go):**
```go
err := api.AdminDisable2FA("user-uuid")
```

**Via Panel API (Java):**
```java
api.adminDisable2FA("user-uuid");
```

## Session Management

Birdactyl tracks all active user sessions. Users can view and revoke sessions at any time from their profile.

- **Access Tokens**: Short-lived JWTs used for API requests.
- **Refresh Tokens**: Long-lived tokens used to rotate access tokens.
- **Session Revocation**: Logging out of "all sessions" invalidates all refresh tokens for that user.

## Root Admins

Root admins are defined in `config.yaml` and have permanent access that cannot be revoked through the UI.

```yaml
root_admins:
  - "user-uuid-1"
```
