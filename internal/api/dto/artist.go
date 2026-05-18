package dto

type RegisterArtistRequest struct {
	Fullname       string `json:"fullname"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Password       string `json:"password"`
	BusinessName   string `json:"business_name"`
	PrimaryService string `json:"primary_service"`
	City           string `json:"city"`
	Instagram      string `json:"instagram"`
}

type ArtistResponse struct {
	ID             string `json:"id"`
	Fullname       string `json:"fullname"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	BusinessName   string `json:"business_name"`
	PrimaryService string `json:"primary_service"`
	City           string `json:"city"`
	Instagram      string `json:"instagram"`
}
