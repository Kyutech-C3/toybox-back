package usecase_test

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/usecase"
	"github.com/simesaba80/toybox-back/internal/usecase/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserUseCase_GetAllUser(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mock.MockUserRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "正常系: ユーザー取得成功",
			setupMock: func(m *mock.MockUserRepository) {
				expectedUsers := []*entity.User{
					{
						ID:          uuid.New(),
						Name:        "user1",
						Email:       "user1@example.com",
						DisplayName: "User One",
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
					{
						ID:          uuid.New(),
						Name:        "user2",
						Email:       "user2@example.com",
						DisplayName: "User Two",
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				m.EXPECT().
					GetAll(gomock.Any()).
					Return(expectedUsers, nil).
					Times(1)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "正常系: ユーザーが0件",
			setupMock: func(m *mock.MockUserRepository) {
				m.EXPECT().
					GetAll(gomock.Any()).
					Return([]*entity.User{}, nil).
					Times(1)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "異常系: リポジトリエラー",
			setupMock: func(m *mock.MockUserRepository) {
				m.EXPECT().
					GetAll(gomock.Any()).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockUserRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewUserUseCase(mockRepo, mock.NewMockAssetRepository(ctrl))

			got, err := uc.GetAllUser(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Len(t, got, tt.wantCount)
			}
		})
	}
}

func TestUserUseCase_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		setupMock func(*mock.MockUserRepository, uuid.UUID)
		wantErr   bool
	}{
		{
			name:   "正常系: ユーザー取得成功",
			userID: uuid.New(),
			setupMock: func(m *mock.MockUserRepository, userID uuid.UUID) {
				expectedUser := &entity.User{
					ID:          userID,
					Name:        "testuser",
					Email:       "testuser@example.com",
					DisplayName: "testuser",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(expectedUser, nil).
					Times(1)
			},
			wantErr: false,
		},
		{
			name:   "異常系: リポジトリエラー",
			userID: uuid.New(),
			setupMock: func(m *mock.MockUserRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockUserRepository(ctrl)
			tt.setupMock(mockRepo, tt.userID)

			uc := usecase.NewUserUseCase(mockRepo, mock.NewMockAssetRepository(ctrl))

			got, err := uc.GetByUserID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.userID, got.ID)
			}
		})
	}
}

func TestUserUseCase_UpdateUser(t *testing.T) {
	originalEmail := "old@example.com"
	originalDisplayName := "Old User"
	originalProfile := "Old profile"
	originalXUsername := "old-x-user"
	originalGithubID := "old-github"
	originalAvatarURL := "https://s3.example.com/bucket/dir/avatar/old.webp"

	updatedDisplayName := "Updated User"
	updatedProfile := "Updated profile"
	updatedXUsername := "xuser123"
	updatedGithubID := "github123"
	updatedAvatarURL := "https://s3.example.com/bucket/dir/avatar/new.png"

	newExistingUser := func(userID uuid.UUID) *entity.User {
		return &entity.User{
			ID:          userID,
			Name:        "testuser",
			Email:       originalEmail,
			DisplayName: originalDisplayName,
			Profile:     originalProfile,
			AvatarURL:   originalAvatarURL,
			XUsername:   originalXUsername,
			GithubID:    originalGithubID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	pngAvatar := &multipart.FileHeader{Filename: "avatar.PNG", Size: 1024}
	unsupportedAvatar := &multipart.FileHeader{Filename: "avatar.mp4", Size: 1024}
	noExtensionAvatar := &multipart.FileHeader{Filename: "avatar", Size: 1024}

	tests := []struct {
		name        string
		userID      uuid.UUID
		displayName *string
		profile     *string
		xUsername   *string
		githubID    *string
		avatar      *multipart.FileHeader
		setupMock   func(*mock.MockUserRepository, *mock.MockAssetRepository, uuid.UUID)
		wantErr     error
		verify      func(*testing.T, *entity.User)
	}{
		{
			name:        "正常系: 全フィールド更新",
			userID:      uuid.New(),
			displayName: &updatedDisplayName,
			profile:     &updatedProfile,
			xUsername:   &updatedXUsername,
			githubID:    &updatedGithubID,
			setupMock: func(m *mock.MockUserRepository, _ *mock.MockAssetRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(newExistingUser(userID), nil).
					Times(1)
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						return user, nil
					}).
					Times(1)
			},
			verify: func(t *testing.T, got *entity.User) {
				assert.Equal(t, originalEmail, got.Email)
				assert.Equal(t, updatedDisplayName, got.DisplayName)
				assert.Equal(t, updatedProfile, got.Profile)
				assert.Equal(t, updatedXUsername, got.XUsername)
				assert.Equal(t, updatedGithubID, got.GithubID)
				assert.Equal(t, originalAvatarURL, got.AvatarURL)
			},
		},
		{
			name:        "正常系: display_nameのみ更新し他のフィールドは現状維持",
			userID:      uuid.New(),
			displayName: &updatedDisplayName,
			setupMock: func(m *mock.MockUserRepository, _ *mock.MockAssetRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(newExistingUser(userID), nil).
					Times(1)
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						return user, nil
					}).
					Times(1)
			},
			verify: func(t *testing.T, got *entity.User) {
				assert.Equal(t, updatedDisplayName, got.DisplayName)
				assert.Equal(t, originalProfile, got.Profile)
				assert.Equal(t, originalXUsername, got.XUsername)
				assert.Equal(t, originalGithubID, got.GithubID)
			},
		},
		{
			name:   "正常系: アバター更新で旧アバターが削除される",
			userID: uuid.New(),
			avatar: pngAvatar,
			setupMock: func(m *mock.MockUserRepository, a *mock.MockAssetRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(newExistingUser(userID), nil).
					Times(1)
				a.EXPECT().
					UploadAvatarFile(gomock.Any(), pngAvatar, "png").
					Return(&updatedAvatarURL, nil).
					Times(1)
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						return user, nil
					}).
					Times(1)
				a.EXPECT().
					DeleteFile(gomock.Any(), originalAvatarURL).
					Return(nil).
					Times(1)
			},
			verify: func(t *testing.T, got *entity.User) {
				assert.Equal(t, updatedAvatarURL, got.AvatarURL)
				assert.Equal(t, originalDisplayName, got.DisplayName)
			},
		},
		{
			name:   "正常系: 旧アバターの削除に失敗しても更新は成功する",
			userID: uuid.New(),
			avatar: pngAvatar,
			setupMock: func(m *mock.MockUserRepository, a *mock.MockAssetRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(newExistingUser(userID), nil).
					Times(1)
				a.EXPECT().
					UploadAvatarFile(gomock.Any(), pngAvatar, "png").
					Return(&updatedAvatarURL, nil).
					Times(1)
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						return user, nil
					}).
					Times(1)
				a.EXPECT().
					DeleteFile(gomock.Any(), originalAvatarURL).
					Return(errors.New("invalid asset URL")).
					Times(1)
			},
			verify: func(t *testing.T, got *entity.User) {
				assert.Equal(t, updatedAvatarURL, got.AvatarURL)
			},
		},
		{
			name:      "異常系: 対応していないアバターの拡張子",
			userID:    uuid.New(),
			avatar:    unsupportedAvatar,
			setupMock: func(m *mock.MockUserRepository, a *mock.MockAssetRepository, userID uuid.UUID) {},
			wantErr:   domainerrors.ErrUnsupportedFileType,
		},
		{
			name:      "異常系: アバターのファイル名に拡張子がない",
			userID:    uuid.New(),
			avatar:    noExtensionAvatar,
			setupMock: func(m *mock.MockUserRepository, a *mock.MockAssetRepository, userID uuid.UUID) {},
			wantErr:   domainerrors.ErrInvalidFileName,
		},
		{
			name:   "異常系: アバターのアップロードに失敗",
			userID: uuid.New(),
			avatar: pngAvatar,
			setupMock: func(m *mock.MockUserRepository, a *mock.MockAssetRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(newExistingUser(userID), nil).
					Times(1)
				a.EXPECT().
					UploadAvatarFile(gomock.Any(), pngAvatar, "png").
					Return(nil, errors.New("s3 error")).
					Times(1)
			},
			wantErr: domainerrors.ErrFailedToUploadFile,
		},
		{
			name:   "異常系: DB更新に失敗した場合はアップロードしたアバターを削除する",
			userID: uuid.New(),
			avatar: pngAvatar,
			setupMock: func(m *mock.MockUserRepository, a *mock.MockAssetRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(newExistingUser(userID), nil).
					Times(1)
				a.EXPECT().
					UploadAvatarFile(gomock.Any(), pngAvatar, "png").
					Return(&updatedAvatarURL, nil).
					Times(1)
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil, domainerrors.ErrFailedToUpdateUser).
					Times(1)
				a.EXPECT().
					DeleteFile(gomock.Any(), updatedAvatarURL).
					Return(nil).
					Times(1)
			},
			wantErr: domainerrors.ErrFailedToUpdateUser,
		},
		{
			name:   "異常系: ユーザーが見つからない",
			userID: uuid.New(),
			setupMock: func(m *mock.MockUserRepository, _ *mock.MockAssetRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(nil, domainerrors.ErrUserNotFound).
					Times(1)
			},
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name:        "異常系: 更新に失敗",
			userID:      uuid.New(),
			displayName: &updatedDisplayName,
			setupMock: func(m *mock.MockUserRepository, _ *mock.MockAssetRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(newExistingUser(userID), nil).
					Times(1)
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil, domainerrors.ErrFailedToUpdateUser).
					Times(1)
			},
			wantErr: domainerrors.ErrFailedToUpdateUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockUserRepository(ctrl)
			mockAssetRepo := mock.NewMockAssetRepository(ctrl)
			tt.setupMock(mockRepo, mockAssetRepo, tt.userID)

			uc := usecase.NewUserUseCase(mockRepo, mockAssetRepo)

			got, err := uc.UpdateUser(context.Background(), tt.userID, tt.displayName, tt.profile, tt.xUsername, tt.githubID, tt.avatar)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.userID, got.ID)
				if tt.verify != nil {
					tt.verify(t, got)
				}
			}
		})
	}
}
