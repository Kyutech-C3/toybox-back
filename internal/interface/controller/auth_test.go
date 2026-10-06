package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/interface/controller/mock"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/pkg/echovalidator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthController_GetDiscordAuthURL(t *testing.T) {
	ctx := context.Background()

	successAuthURL := "https://discord.com/oauth2/authorize"
	successResponseBytes, _ := json.Marshal(schema.ToGetDiscordAuthURLResponse(successAuthURL))
	clientIDNotSetResponseBytes, _ := json.Marshal(map[string]string{"message": "DiscordクライアントIDが設定されていません"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "Internal server error"})

	tests := []struct {
		name       string
		setupMock  func(mockAuthUsecase *mock.MockIAuthUsecase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "正常系: 認証URL取得成功",
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					GetDiscordAuthURL(gomock.Any()).
					DoAndReturn(func(context.Context) (string, error) {
						return successAuthURL, nil
					})
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "異常系: DiscordクライアントID未設定",
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					GetDiscordAuthURL(gomock.Any()).
					Return("", domainerrors.ErrClientIDNotSet)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   clientIDNotSetResponseBytes,
		},
		{
			name: "異常系: 予期しないエラー",
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					GetDiscordAuthURL(gomock.Any()).
					Return("", errors.New("unexpected error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   internalErrorResponseBytes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUsecase := mock.NewMockIAuthUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockAuthUsecase)
			}

			authController := controller.NewAuthController(mockAuthUsecase)
			e.GET("/auth/discord/", authController.GetDiscordAuthURL)

			req := httptest.NewRequest(http.MethodGet, "/auth/discord/", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}

func TestAuthController_AuthenticateUser(t *testing.T) {
	originalEnv := config.ENV
	t.Cleanup(func() { config.ENV = originalEnv })

	refreshToken := &entity.Token{RefreshToken: uuid.New(), ExpiredAt: time.Now().Add(30 * 24 * time.Hour)}

	successResponseBytes, _ := json.Marshal(schema.ToGetDiscordTokenResponse("app-token"))
	codeRequiredResponseBytes, _ := json.Marshal(map[string]string{"message": "code is required"})
	userNotAllowedResponseBytes, _ := json.Marshal(map[string]string{"message": "ユーザーは許可されたDiscordギルドに所属していません"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "Internal server error"})
	failedRequestResponseBytes, _ := json.Marshal(map[string]string{"message": "Discordへのリクエストに失敗しました"})

	tests := []struct {
		name       string
		env        string
		query      string
		setupMock  func(mockAuthUsecase *mock.MockIAuthUsecase)
		wantStatus int
		wantBody   []byte
	}{
		{
			env:   "dev",
			name:  "正常系: ユーザー認証成功（dev）",
			query: "?code=test-code",
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					AuthenticateUser(gomock.Any(), "test-code").
					Return("app-token", refreshToken, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			env:   "prod",
			name:  "正常系: ユーザー認証成功（prod）",
			query: "?code=test-code",
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					AuthenticateUser(gomock.Any(), "test-code").
					Return("app-token", refreshToken, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name:       "異常系: codeが空",
			query:      "",
			setupMock:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   codeRequiredResponseBytes,
		},
		{
			name:  "異常系: 許可されていないギルドのユーザー",
			query: "?code=another-code",
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					AuthenticateUser(gomock.Any(), "another-code").
					Return("", nil, domainerrors.ErrUserNotAllowedGuild)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   userNotAllowedResponseBytes,
		},
		{
			name:  "異常系: Discordリクエスト失敗",
			query: "?code=discord-error",
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					AuthenticateUser(gomock.Any(), "discord-error").
					Return("", nil, domainerrors.ErrFaileRequestToDiscord)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   failedRequestResponseBytes,
		},
		{
			name:  "異常系: 予期しないエラー",
			query: "?code=unexpected",
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					AuthenticateUser(gomock.Any(), "unexpected").
					Return("", nil, errors.New("unexpected error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   internalErrorResponseBytes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.ENV = tt.env
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUsecase := mock.NewMockIAuthUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockAuthUsecase)
			}

			authController := controller.NewAuthController(mockAuthUsecase)
			e.GET("/auth/discord/callback", authController.AuthenticateUser)

			req := httptest.NewRequest(http.MethodGet, "/auth/discord/callback"+tt.query, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
			if tt.wantStatus == http.StatusOK {
				assertRefreshTokenCookie(t, rec, refreshToken)
			}
		})
	}
}

func TestAuthController_RegenerateToken(t *testing.T) {
	originalEnv := config.ENV
	t.Cleanup(func() { config.ENV = originalEnv })

	regeneratedToken := &entity.Token{RefreshToken: uuid.New(), ExpiredAt: time.Now().Add(30 * 24 * time.Hour)}

	successResponseBytes, _ := json.Marshal(schema.ToRegenerateTokenResponse("new-app-token"))
	refreshRequiredResponseBytes, _ := json.Marshal(map[string]string{"message": "Refresh token is required"})
	expiredResponseBytes, _ := json.Marshal(map[string]string{"message": "リフレッシュトークンが期限切れです"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "Internal server error"})

	validRefreshToken := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	expiredRefreshToken := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	unexpectedRefreshToken := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	tests := []struct {
		name       string
		env        string
		cookie     *http.Cookie
		setupMock  func(mockAuthUsecase *mock.MockIAuthUsecase)
		wantStatus int
		wantBody   []byte
	}{
		{
			env:    "dev",
			name:   "正常系: トークン再発行成功（dev）",
			cookie: &http.Cookie{Name: "refresh_token", Value: validRefreshToken.String()},
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					RegenerateToken(gomock.Any(), validRefreshToken).
					Return("new-app-token", regeneratedToken, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			env:    "prod",
			name:   "正常系: トークン再発行成功（prod）",
			cookie: &http.Cookie{Name: "refresh_token", Value: validRefreshToken.String()},
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					RegenerateToken(gomock.Any(), validRefreshToken).
					Return("new-app-token", regeneratedToken, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name:       "異常系: リフレッシュトークン未送信",
			cookie:     nil,
			setupMock:  nil,
			wantStatus: http.StatusBadRequest,
			wantBody:   refreshRequiredResponseBytes,
		},
		{
			name:   "異常系: リフレッシュトークン期限切れ",
			cookie: &http.Cookie{Name: "refresh_token", Value: expiredRefreshToken.String()},
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					RegenerateToken(gomock.Any(), expiredRefreshToken).
					Return("", nil, domainerrors.ErrRefreshTokenExpired)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   expiredResponseBytes,
		},
		{
			name:   "異常系: 予期しないエラー",
			cookie: &http.Cookie{Name: "refresh_token", Value: unexpectedRefreshToken.String()},
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					RegenerateToken(gomock.Any(), unexpectedRefreshToken).
					Return("", nil, errors.New("unexpected error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   internalErrorResponseBytes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.ENV = tt.env
			e := echo.New()
			e.Validator = echovalidator.NewValidator()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUsecase := mock.NewMockIAuthUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockAuthUsecase)
			}

			authController := controller.NewAuthController(mockAuthUsecase)
			e.POST("/auth/refresh", authController.RegenerateToken)

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
			if tt.wantStatus == http.StatusOK {
				assertRefreshTokenCookie(t, rec, regeneratedToken)
			}
		})
	}
}

func assertRefreshTokenCookie(t *testing.T, rec *httptest.ResponseRecorder, want *entity.Token) {
	t.Helper()
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "refresh_token", cookies[0].Name)
	assert.Equal(t, want.RefreshToken.String(), cookies[0].Value)
	assert.WithinDuration(t, want.ExpiredAt, cookies[0].Expires, time.Second)
	assert.WithinDuration(t, want.ExpiredAt, time.Now().Add(time.Duration(cookies[0].MaxAge)*time.Second), 2*time.Second)
}

func TestAuthController_Logout(t *testing.T) {
	validRefreshToken := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name           string
		cookie         *http.Cookie
		setupMock      func(mockAuthUsecase *mock.MockIAuthUsecase)
		wantStatus     int
		wantCookieGone bool
	}{
		{
			name:   "正常系: ログアウト成功",
			cookie: &http.Cookie{Name: "refresh_token", Value: validRefreshToken.String()},
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					Logout(gomock.Any(), validRefreshToken).
					Return(nil)
			},
			wantStatus:     http.StatusOK,
			wantCookieGone: true,
		},
		{
			name:           "正常系: Cookieなしでもログアウト成功",
			cookie:         nil,
			setupMock:      nil,
			wantStatus:     http.StatusOK,
			wantCookieGone: false,
		},
		{
			name:   "正常系: 無効なトークンでもログアウト成功（Cookieは削除）",
			cookie: &http.Cookie{Name: "refresh_token", Value: "invalid-uuid"},
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				// UUIDパースエラーの場合、Logoutは呼ばれない
			},
			wantStatus:     http.StatusOK,
			wantCookieGone: true,
		},
		{
			name:   "正常系: DBエラーでもログアウト成功（Cookieは削除）",
			cookie: &http.Cookie{Name: "refresh_token", Value: validRefreshToken.String()},
			setupMock: func(mockAuthUsecase *mock.MockIAuthUsecase) {
				mockAuthUsecase.EXPECT().
					Logout(gomock.Any(), validRefreshToken).
					Return(errors.New("db error"))
			},
			wantStatus:     http.StatusOK,
			wantCookieGone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUsecase := mock.NewMockIAuthUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockAuthUsecase)
			}

			authController := controller.NewAuthController(mockAuthUsecase)
			e.POST("/auth/logout", authController.Logout)

			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			// Cookieが削除されているか確認
			if tt.wantCookieGone {
				cookies := rec.Result().Cookies()
				for _, c := range cookies {
					if c.Name == "refresh_token" {
						assert.Equal(t, "", c.Value)
						assert.Equal(t, -1, c.MaxAge)
					}
				}
			}
		})
	}
}
