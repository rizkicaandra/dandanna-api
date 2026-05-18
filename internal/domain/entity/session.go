package entity

// Session is the payload stored under an access token key in Redis.
type Session struct {
	UserID string
	Role   string
	Name   string
}

// SessionPointer is stored under the per-user session key in Redis.
// It links a user to their current active access and refresh tokens,
// enabling single-session enforcement and immediate access token revocation on logout.
type SessionPointer struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
