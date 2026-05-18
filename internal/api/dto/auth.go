package dto

type LoginArtistRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginArtistResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	Artist       ArtistResponse `json:"artist"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
