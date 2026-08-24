package controller

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/internal/usecase"
)

type AuthController struct {
	authUsecase usecase.IAuthUsecase
}

func NewAuthController(authUsecase usecase.IAuthUsecase) *AuthController {
	return &AuthController{
		authUsecase: authUsecase,
	}
}

// GetDiscordAuthURL godoc
// @Summary Get Discord authentication URL
// @Description Get Discord authentication URL
// @Tags auth
// @Produce json
// @Success 200 {object} schema.GetDiscordAuthURLResponse
// @Failure 500 {object} echo.HTTPError
// @Router /auth/discord [get]
func (ac *AuthController) GetDiscordAuthURL(c echo.Context) error {
	authURL, err := ac.authUsecase.GetDiscordAuthURL(c.Request().Context())
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.JSON(http.StatusOK, schema.ToGetDiscordAuthURLResponse(authURL))
}

// GetDiscordToken godoc
// @Summary Get Discord token by code
// @Description Get Discord token by code
// @Tags auth
// @Produce json
// @Success 200 {object} schema.GetDiscordTokenResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/discord/callback [get]
// @Param code query string true "Discord code"
func (ac *AuthController) AuthenticateUser(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "code is required")
	}
	appToken, refreshToken, err := ac.authUsecase.AuthenticateUser(c.Request().Context(), code)
	if err != nil {
		c.Logger().Error("Failed to authenticate user: %w", err)
		return handleAuthError(c, err)
	}
	switch config.ENV {
	case "prod":
		cookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Path:     "/auth",
		}
		c.SetCookie(cookie)
	case "dev":
		cookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Path:     "/auth",
		}
		c.SetCookie(cookie)
	}

	return c.JSON(http.StatusOK, schema.ToGetDiscordTokenResponse(appToken))
}

// RegenerateToken godoc
// @Summary Regenerate token
// @Description Regenerate token(need refresh token in cookie)
// @Tags auth
// @Produce json
// @Success 200 {object} schema.GetDiscordTokenResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/refresh [post]
func (ac *AuthController) RegenerateToken(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.Logger().Error("Refresh token is required")
		return echo.NewHTTPError(http.StatusBadRequest, "Refresh token is required")
	}
	refreshToken, err := uuid.Parse(cookie.Value)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid refresh token")
	}
	appToken, newRefreshToken, err := ac.authUsecase.RegenerateToken(c.Request().Context(), refreshToken)
	if err != nil {
		return handleAuthError(c, err)
	}
	switch config.ENV {
	case "prod":
		cookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Path:     "/auth",
		}
		c.SetCookie(cookie)
	case "dev":
		cookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Path:     "/auth",
		}
		c.SetCookie(cookie)
	}
	return c.JSON(http.StatusOK, schema.ToRegenerateTokenResponse(appToken))
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and invalidate refresh token
// @Tags auth
// @Success 200
// @Failure 400 {object} echo.HTTPError
// @Router /auth/logout [post]
func (ac *AuthController) Logout(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		// Cookieがない場合でもログアウト成功として扱う
		return c.NoContent(http.StatusOK)
	}

	refreshToken, err := uuid.Parse(cookie.Value)
	if err == nil {
		// DBからリフレッシュトークンを削除（エラーは無視）
		_ = ac.authUsecase.Logout(c.Request().Context(), refreshToken)
	}

	// Cookieを無効化
	switch config.ENV {
	case "prod":
		invalidCookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Path:     "/auth",
			MaxAge:   -1,
		}
		c.SetCookie(invalidCookie)
	case "dev":
		invalidCookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Path:     "/auth",
			MaxAge:   -1,
		}
		c.SetCookie(invalidCookie)
	}

	return c.NoContent(http.StatusOK)
}

func handleAuthError(c echo.Context, err error) error {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	switch {
	case errors.Is(err, domainerrors.ErrUserNotAllowedGuild):
		return echo.NewHTTPError(http.StatusForbidden, "ユーザーは許可されたDiscordギルドに所属していません")
	case errors.Is(err, domainerrors.ErrRefreshTokenExpired):
		return echo.NewHTTPError(http.StatusBadRequest, "リフレッシュトークンが期限切れです")
	case errors.Is(err, domainerrors.ErrRefreshTokenInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, "リフレッシュトークンが無効です")
	case errors.Is(err, domainerrors.ErrFaileRequestToDiscord):
		return echo.NewHTTPError(http.StatusInternalServerError, "Discordへのリクエストに失敗しました")
	case errors.Is(err, domainerrors.ErrClientIDNotSet):
		return echo.NewHTTPError(http.StatusInternalServerError, "DiscordクライアントIDが設定されていません")
	case errors.Is(err, domainerrors.ErrRedirectURLNotSet):
		return echo.NewHTTPError(http.StatusInternalServerError, "リダイレクトURLが設定されていません")
	case errors.Is(err, domainerrors.ErrFailedToCreateUser):
		return echo.NewHTTPError(http.StatusInternalServerError, "ユーザーの作成に失敗しました")
	default:
		c.Logger().Error("Auth error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}
}
