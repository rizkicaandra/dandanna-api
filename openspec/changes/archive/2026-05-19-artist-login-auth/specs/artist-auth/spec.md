## ADDED Requirements

### Requirement: Artist can log in with email and password
The system SHALL verify the artist's credentials against the stored argon2id hash and issue an opaque access token and refresh token on success. Both tokens are stored in Redis under app-namespaced keys (`artist:*`). If an existing session is found for the artist, it SHALL be revoked before the new tokens are issued (single-session policy).

#### Scenario: Successful login
- **WHEN** an artist sends valid email and password to `POST /api/artists/login`
- **THEN** the system returns HTTP 200 with `access_token`, `refresh_token`, and artist profile data

#### Scenario: Wrong password
- **WHEN** an artist sends a valid email but incorrect password
- **THEN** the system returns HTTP 401 with error code `INVALID_CREDENTIALS` and no token is issued

#### Scenario: Email not found
- **WHEN** an artist sends an email that does not exist in the system
- **THEN** the system returns HTTP 401 with error code `INVALID_CREDENTIALS` (same response as wrong password — no user enumeration)

#### Scenario: Account suspended
- **WHEN** an artist's `role_status` is not `active`
- **THEN** the system returns HTTP 403 with error code `ACCOUNT_SUSPENDED` and no token is issued

#### Scenario: Existing session is kicked on new login
- **WHEN** an artist logs in while a previous session exists
- **THEN** the old access and refresh tokens are deleted from Redis before new tokens are issued

#### Scenario: Missing required fields
- **WHEN** an artist sends a login request without email or password
- **THEN** the system returns HTTP 422 with validation errors identifying the missing fields

---

### Requirement: Login is protected against brute-force attacks
The system SHALL track failed login attempts per IP address using Redis. After 10 failed attempts within a 15-minute sliding window, the system SHALL reject further login requests from that IP regardless of credentials. The counter SHALL reset on successful login.

#### Scenario: IP exceeds attempt limit
- **WHEN** an IP address makes 10 or more failed login attempts within 15 minutes
- **THEN** the system returns HTTP 422 with error code `TOO_MANY_LOGIN_ATTEMPTS` without checking credentials

#### Scenario: Successful login resets attempt counter
- **WHEN** an artist successfully logs in after previous failed attempts
- **THEN** the IP's failed attempt counter is reset to zero in Redis

---

### Requirement: Artist can refresh their access token
The system SHALL validate the provided refresh token and issue a new access token and a new refresh token (rotation). The old refresh token and old access token SHALL be deleted from Redis atomically with the new token issuance.

#### Scenario: Successful token refresh
- **WHEN** an artist sends a valid refresh token to `POST /api/artists/auth/refresh`
- **THEN** the system returns HTTP 200 with a new `access_token` and a new `refresh_token`, and the old tokens are revoked

#### Scenario: Expired or invalid refresh token
- **WHEN** an artist sends an expired, revoked, or invalid refresh token
- **THEN** the system returns HTTP 401 with error code `INVALID_REFRESH_TOKEN`

#### Scenario: Old refresh token cannot be reused after rotation
- **WHEN** an artist uses a refresh token that was already rotated
- **THEN** the system returns HTTP 401 with error code `INVALID_REFRESH_TOKEN`

---

### Requirement: Artist can log out
The system SHALL immediately revoke both the access token and the refresh token from Redis upon logout. The session pointer SHALL also be deleted so no dangling references remain.

#### Scenario: Successful logout
- **WHEN** an artist sends a valid refresh token to `POST /api/artists/logout`
- **THEN** the system returns HTTP 200, and both the access token and refresh token are deleted from Redis immediately

#### Scenario: Access token is immediately invalid after logout
- **WHEN** an artist uses their access token after a successful logout
- **THEN** the system SHALL reject the token (access token no longer exists in Redis)

#### Scenario: Invalid refresh token on logout
- **WHEN** an artist sends an invalid or already-revoked refresh token to logout
- **THEN** the system returns HTTP 401 with error code `INVALID_REFRESH_TOKEN`
