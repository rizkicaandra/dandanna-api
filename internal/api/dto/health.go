package dto

// HealthData is the data payload returned by GET /api/health.
// It is embedded inside the standard Response envelope.
type HealthData struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ReadyzData is the data payload returned by GET /api/readyz.
// Each field represents the status of a dependency: "ok" or "unavailable".
type ReadyzData struct {
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}
