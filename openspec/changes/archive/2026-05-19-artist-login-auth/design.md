## Context

Artist registration is complete but there is no authentication flow. Artists cannot log in, receive tokens, or access protected endpoints. This design covers token issuance, rotation, revocation, single-session enforcement, and brute-force protection — all using existing infrastructure (Redis, PostgreSQL, `net/http`).

## Goals / Non-Goals

**Goals:**
- Secure credential verification using existing argon2id hashes
- Opaque token pair (access + refresh) backed by Redis — instant revocation at any time
- Single active session per artist — new login kicks the previous session
- Refresh token rotation on every use — stolen refresh tokens are single-use
- Immediate access token revocation on logout — no waiting for TTL
- IP-based brute-force protection on login (Redis counter, 15-min sliding window)
- App-namespaced Redis keys (`artist:*`) — structurally prevents cross-app token misuse

**Non-Goals:**
- JWT — opaque tokens are strictly better given Redis is already wired in and instant revocation is required
- Multi-device sessions — not needed for the current artist use case; can be introduced later by relaxing the session pointer logic
- Customer or admin auth — this change covers only the `artist` app namespace
- OAuth / social login

## Decisions

### 1. Opaque tokens over JWT
**Decision:** 32 random bytes → 64-char hex, stored in Redis.
**Rationale:** JWTs cannot be instantly revoked without a Redis blocklist — which would make Redis load-bearing anyway. Opaque tokens give us instant revocation for free. No extra dependency.
**Alternative:** JWT with short TTL + Redis blocklist → same Redis dependency, more complexity, no benefit.

### 2. App-namespaced Redis keys
**Decision:** All keys prefixed with the app name: `artist:access:<token>`, `artist:refresh:<token>`, `artist:session:<userID>`, `artist:login_attempts:<ip>`.
**Rationale:** When customer auth is added, `customer:*` keys are isolated by design. A token issued for one app cannot be validated by another app's endpoint — the namespace enforces this at the Redis level without any application-level field check.

### 3. Single-session enforcement via session pointer
**Decision:** `artist:session:<userID>` → `{refresh_token, access_token}` (TTL = refresh TTL).
**Rationale:** Prevents unlimited session accumulation from repeated logins without logout. On new login: read pointer → delete old tokens → issue new tokens → update pointer. Atomic enough at this scale; no Redis transaction needed since worst case is a dangling old token that expires naturally.

### 4. Refresh token rotation
**Decision:** Every call to `POST /api/artists/auth/refresh` issues a new access + refresh token pair and deletes the old pair.
**Rationale:** Single-use refresh tokens limit the blast radius of a stolen token. If an attacker uses it first, the legitimate user's next refresh call returns 401 — alerting them to the breach.

### 5. Immediate access token revocation on logout
**Decision:** Session pointer stores the current access token so logout can delete `artist:access:<token>` directly.
**Rationale:** Without this, a logged-out artist's access token remains valid for up to 15 minutes. Since we're already storing per-user session state in Redis, the cost of storing the access token reference is zero.

### 6. IP-based brute-force protection
**Decision:** `artist:login_attempts:<ip>` counter, max 10 failures per 15-min sliding window. Increment on any auth failure (wrong password or email not found). Reset on successful login.
**Rationale:** Email-based lockout enables account DoS (attacker locks out victim). IP-based lockout throttles the attacker. For dandanna's user base (individual artists on personal devices), shared-IP false positives are negligible.

### 7. TokenRepository interface in domain layer
**Decision:** `internal/domain/repository/token.go` defines the interface; `internal/infrastructure/redis/token_repository.go` implements it.
**Rationale:** Consistent with existing architecture (`ArtistRepository` interface in domain, postgres implementation in infrastructure). Keeps the service layer testable without a real Redis instance.

## Risks / Trade-offs

- **Single-session strictness** → An artist using two devices simultaneously will be kicked from the first device when logging in on the second. Acceptable for now; relax later by moving to a per-device session set if needed.
- **Redis as auth dependency** → If Redis goes down, all token validation fails. Mitigated by Redis health check on `/api/readyz` and proper connection pooling.
- **IP extraction behind proxy** → `X-Real-IP` header can be spoofed if the reverse proxy is misconfigured. Mitigation: ensure the proxy is configured to overwrite (not append) `X-Real-IP`.
- **Session pointer race on concurrent logins** → Two simultaneous logins from the same user could both read the same old session pointer, each deleting the other's new tokens. Acceptable at current scale; fix with Redis WATCH/MULTI if needed later.
