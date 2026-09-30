package schema

import (
	"github.com/golang-jwt/jwt/v5"
)

type GetDiscordAuthURLResponse struct {
	URL string `json:"url"`
}

type GetDiscordTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type DiscordUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
}

type RegenerateTokenInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RegenerateTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type JWTCustomClaims struct {
	UserID string `json:"user_id"`
	// RefreshToken is optional and used when embedding refresh token in JWTs.
	RefreshToken string `json:"refresh_token,omitempty"`
	jwt.RegisteredClaims
}

func ToGetDiscordAuthURLResponse(authURL string) GetDiscordAuthURLResponse {
	return GetDiscordAuthURLResponse{
		URL: authURL,
	}
}

func ToGetDiscordTokenResponse(appToken string) GetDiscordTokenResponse {
	return GetDiscordTokenResponse{
		AccessToken: appToken,
	}
}

func ToRegenerateTokenResponse(appToken string) RegenerateTokenResponse {
	return RegenerateTokenResponse{
		AccessToken: appToken,
	}
}
