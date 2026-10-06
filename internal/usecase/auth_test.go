package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/usecase/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthUsecase_GetDiscordAuthURL(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mock.MockDiscordRepository, *mock.MockUserRepository, *mock.MockTokenProvider, *mock.MockTokenRepository, *mock.MockAssetRepository)
		wantAuthURL string
		wantErr     bool
	}{
		{
			name: "正常系: ディスコード認証URL取得成功",
			setupMock: func(m *mock.MockDiscordRepository, _ *mock.MockUserRepository, _ *mock.MockTokenProvider, _ *mock.MockTokenRepository, _ *mock.MockAssetRepository) {
				m.EXPECT().
					GetDiscordClientID(gomock.Any()).
					Return("1234567890", nil).
					Times(1)
				m.EXPECT().
					GetRedirectURL(gomock.Any()).
					Return("https://localhost:8080", nil).
					Times(1)
				m.EXPECT().
					GetDiscordAuthURL(gomock.Any()).
					Return("https://localhost:8080/auth/discord/callback", nil).
					Times(1)
			},
			wantAuthURL: "https://localhost:8080/auth/discord/callback",
			wantErr:     false,
		},
		{
			name: "異常系: ディスコードクライアントID取得失敗",
			setupMock: func(m *mock.MockDiscordRepository, _ *mock.MockUserRepository, _ *mock.MockTokenProvider, _ *mock.MockTokenRepository, _ *mock.MockAssetRepository) {
				m.EXPECT().
					GetDiscordClientID(gomock.Any()).
					Return("", domainerrors.ErrClientIDNotSet).
					Times(1)
			},
			wantErr: true,
		},
		{
			name: "異常系: リダイレクトURL取得失敗",
			setupMock: func(m *mock.MockDiscordRepository, _ *mock.MockUserRepository, _ *mock.MockTokenProvider, _ *mock.MockTokenRepository, _ *mock.MockAssetRepository) {
				m.EXPECT().
					GetDiscordClientID(gomock.Any()).
					Return("1234567890", nil).
					Times(1)
				m.EXPECT().
					GetRedirectURL(gomock.Any()).
					Return("", domainerrors.ErrRedirectURLNotSet).
					Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDiscordRepository := mock.NewMockDiscordRepository(ctrl)
			mockUserRepository := mock.NewMockUserRepository(ctrl)
			mockTokenProvider := mock.NewMockTokenProvider(ctrl)
			mockTokenRepository := mock.NewMockTokenRepository(ctrl)
			mockAssetRepository := mock.NewMockAssetRepository(ctrl)
			tt.setupMock(mockDiscordRepository, mockUserRepository, mockTokenProvider, mockTokenRepository, mockAssetRepository)
			uc := NewAuthUsecase(mockDiscordRepository, mockUserRepository, mockTokenProvider, mockTokenRepository, mockAssetRepository)
			got, err := uc.GetDiscordAuthURL(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestAuthUsecase_AuthenticateUser(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mock.MockDiscordRepository, *mock.MockUserRepository, *mock.MockTokenProvider, *mock.MockTokenRepository, *mock.MockAssetRepository)
		wantErr   bool
	}{
		{
			name: "正常系: 既存ユーザーのディスコード認証成功",
			setupMock: func(m *mock.MockDiscordRepository, u *mock.MockUserRepository, tp *mock.MockTokenProvider, tr *mock.MockTokenRepository, _ *mock.MockAssetRepository) {
				m.EXPECT().
					GetDiscordToken(gomock.Any(), gomock.Any()).
					Return(entity.DiscordToken{
						AccessToken:  "test",
						RefreshToken: "test",
						Expiry:       time.Now().Add(1 * time.Hour),
						ExpiresIn:    3600,
						TokenType:    "Bearer",
					}, nil).
					Times(1)
				m.EXPECT().
					FetchDiscordUser(gomock.Any(), gomock.Any()).
					Return(entity.DiscordUser{
						ID: "test",
					}, nil).
					Times(1)
				m.EXPECT().
					GetDiscordGuilds(gomock.Any(), gomock.Any()).
					Return([]string{"test"}, nil).
					Times(1)
				m.EXPECT().
					GetAllowedDiscordGuilds(gomock.Any()).
					Return([]string{"test"}, nil).
					Times(1)
				u.EXPECT().
					GetUserByDiscordUserID(gomock.Any(), gomock.Any()).
					Return(&entity.User{
						ID:            uuid.New(),
						Name:          "test",
						Email:         "test@example.com",
						DisplayName:   "test",
						DiscordUserID: "test",
						AvatarURL:     "https://example.com/avatar.png",
					}, nil).
					Times(1)
				tp.EXPECT().
					GenerateToken(gomock.Any()).
					Return("test", nil).
					Times(1)
				tr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Token{
						RefreshToken: uuid.New(),
					}, nil).
					Times(1)
			},
			wantErr: false,
		},
		{
			name: "正常系: 新規ユーザーのディスコード認証成功",
			setupMock: func(m *mock.MockDiscordRepository, u *mock.MockUserRepository, tp *mock.MockTokenProvider, tr *mock.MockTokenRepository, a *mock.MockAssetRepository) {
				m.EXPECT().
					GetDiscordToken(gomock.Any(), gomock.Any()).
					Return(entity.DiscordToken{
						AccessToken: "test",
					}, nil).
					Times(1)
				m.EXPECT().
					FetchDiscordUser(gomock.Any(), gomock.Any()).
					Return(entity.DiscordUser{
						ID:         "test",
						AvatarHash: "avatarhash",
					}, nil).
					Times(1)
				m.EXPECT().
					GetDiscordGuilds(gomock.Any(), gomock.Any()).
					Return([]string{"test"}, nil).
					Times(1)
				m.EXPECT().
					GetAllowedDiscordGuilds(gomock.Any()).
					Return([]string{"test"}, nil).
					Times(1)
				u.EXPECT().
					GetUserByDiscordUserID(gomock.Any(), gomock.Any()).
					Return(nil, domainerrors.ErrUserNotFound).
					Times(1)
				avatarURL := "https://example.com/avatar.webp"
				a.EXPECT().
					UploadAvatar(gomock.Any(), "test", "avatarhash").
					Return(&avatarURL, nil).
					Times(1)
				u.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.User{
						ID: uuid.New(),
					}, nil).
					Times(1)
				tp.EXPECT().
					GenerateToken(gomock.Any()).
					Return("test", nil).
					Times(1)
				tr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Token{
						RefreshToken: uuid.New(),
					}, nil).
					Times(1)
			},
			wantErr: false,
		},
		{
			name: "異常系: ディスコード認証失敗",
			setupMock: func(m *mock.MockDiscordRepository, u *mock.MockUserRepository, tp *mock.MockTokenProvider, tr *mock.MockTokenRepository, _ *mock.MockAssetRepository) {
				m.EXPECT().
					GetDiscordToken(gomock.Any(), gomock.Any()).
					Return(entity.DiscordToken{}, domainerrors.ErrFaileRequestToDiscord).
					Times(1)
			},
			wantErr: true,
		},
		{
			name: "異常系: ユーザーが許可されたDiscordギルドに所属していない",
			setupMock: func(m *mock.MockDiscordRepository, u *mock.MockUserRepository, tp *mock.MockTokenProvider, tr *mock.MockTokenRepository, _ *mock.MockAssetRepository) {
				m.EXPECT().
					GetDiscordToken(gomock.Any(), gomock.Any()).
					Return(entity.DiscordToken{
						AccessToken: "test",
					}, nil).
					Times(1)
				m.EXPECT().
					FetchDiscordUser(gomock.Any(), gomock.Any()).
					Return(entity.DiscordUser{
						ID: "test",
					}, nil).
					Times(1)
				m.EXPECT().
					GetDiscordGuilds(gomock.Any(), gomock.Any()).
					Return([]string{"guild"}, nil).
					Times(1)
				m.EXPECT().
					GetAllowedDiscordGuilds(gomock.Any()).
					Return([]string{"allowed_guild"}, nil).
					Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDiscordRepository := mock.NewMockDiscordRepository(ctrl)
			mockUserRepository := mock.NewMockUserRepository(ctrl)
			mockTokenProvider := mock.NewMockTokenProvider(ctrl)
			mockTokenRepository := mock.NewMockTokenRepository(ctrl)
			mockAssetRepository := mock.NewMockAssetRepository(ctrl)
			tt.setupMock(mockDiscordRepository, mockUserRepository, mockTokenProvider, mockTokenRepository, mockAssetRepository)
			uc := NewAuthUsecase(mockDiscordRepository, mockUserRepository, mockTokenProvider, mockTokenRepository, mockAssetRepository)
			appToken, refreshToken, err := uc.AuthenticateUser(context.Background(), "test")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, appToken)
				assert.Nil(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, appToken)
				assert.NotNil(t, refreshToken)
			}
		})
	}
}

func TestAuthUsecase_Logout(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken uuid.UUID
		setupMock    func(*mock.MockDiscordRepository, *mock.MockUserRepository, *mock.MockTokenProvider, *mock.MockTokenRepository, *mock.MockAssetRepository)
		wantErr      bool
	}{
		{
			name:         "正常系: ログアウト成功",
			refreshToken: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			setupMock: func(_ *mock.MockDiscordRepository, _ *mock.MockUserRepository, _ *mock.MockTokenProvider, tr *mock.MockTokenRepository, _ *mock.MockAssetRepository) {
				tr.EXPECT().
					DeleteRefreshToken(gomock.Any(), uuid.MustParse("11111111-1111-1111-1111-111111111111")).
					Return(nil).
					Times(1)
			},
			wantErr: false,
		},
		{
			name:         "異常系: トークン削除失敗",
			refreshToken: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			setupMock: func(_ *mock.MockDiscordRepository, _ *mock.MockUserRepository, _ *mock.MockTokenProvider, tr *mock.MockTokenRepository, _ *mock.MockAssetRepository) {
				tr.EXPECT().
					DeleteRefreshToken(gomock.Any(), uuid.MustParse("22222222-2222-2222-2222-222222222222")).
					Return(domainerrors.ErrRefreshTokenInvalid).
					Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDiscordRepository := mock.NewMockDiscordRepository(ctrl)
			mockUserRepository := mock.NewMockUserRepository(ctrl)
			mockTokenProvider := mock.NewMockTokenProvider(ctrl)
			mockTokenRepository := mock.NewMockTokenRepository(ctrl)
			mockAssetRepository := mock.NewMockAssetRepository(ctrl)
			tt.setupMock(mockDiscordRepository, mockUserRepository, mockTokenProvider, mockTokenRepository, mockAssetRepository)
			uc := NewAuthUsecase(mockDiscordRepository, mockUserRepository, mockTokenProvider, mockTokenRepository, mockAssetRepository)
			err := uc.Logout(context.Background(), tt.refreshToken)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
