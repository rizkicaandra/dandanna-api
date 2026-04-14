package dto

// HealthData is the data payload returned by GET /api/health.
// It is embedded inside the standard Response envelope.
type HealthData struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
