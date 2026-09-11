package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
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

			uc := usecase.NewUserUseCase(mockRepo)

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

			uc := usecase.NewUserUseCase(mockRepo)

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
	originalTwitterID := "old-twitter"
	originalGithubID := "old-github"

	updatedDisplayName := "Updated User"
	updatedProfile := "Updated profile"
	updatedTwitterID := "twitter123"
	updatedGithubID := "github123"

	newExistingUser := func(userID uuid.UUID) *entity.User {
		return &entity.User{
			ID:          userID,
			Name:        "testuser",
			Email:       originalEmail,
			DisplayName: originalDisplayName,
			Profile:     originalProfile,
			TwitterID:   originalTwitterID,
			GithubID:    originalGithubID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	tests := []struct {
		name        string
		userID      uuid.UUID
		displayName *string
		profile     *string
		twitterID   *string
		githubID    *string
		setupMock   func(*mock.MockUserRepository, uuid.UUID)
		wantErr     bool
		verify      func(*testing.T, *entity.User)
	}{
		{
			name:        "正常系: 全フィールド更新",
			userID:      uuid.New(),
			displayName: &updatedDisplayName,
			profile:     &updatedProfile,
			twitterID:   &updatedTwitterID,
			githubID:    &updatedGithubID,
			setupMock: func(m *mock.MockUserRepository, userID uuid.UUID) {
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
			wantErr: false,
			verify: func(t *testing.T, got *entity.User) {
				assert.Equal(t, originalEmail, got.Email)
				assert.Equal(t, updatedDisplayName, got.DisplayName)
				assert.Equal(t, updatedProfile, got.Profile)
				assert.Equal(t, updatedTwitterID, got.TwitterID)
				assert.Equal(t, updatedGithubID, got.GithubID)
			},
		},
		{
			name:        "正常系: display_nameのみ更新し他のフィールドは現状維持",
			userID:      uuid.New(),
			displayName: &updatedDisplayName,
			setupMock: func(m *mock.MockUserRepository, userID uuid.UUID) {
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
			wantErr: false,
			verify: func(t *testing.T, got *entity.User) {
				assert.Equal(t, updatedDisplayName, got.DisplayName)
				assert.Equal(t, originalProfile, got.Profile)
				assert.Equal(t, originalTwitterID, got.TwitterID)
				assert.Equal(t, originalGithubID, got.GithubID)
			},
		},
		{
			name:   "異常系: ユーザーが見つからない",
			userID: uuid.New(),
			setupMock: func(m *mock.MockUserRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(nil, errors.New("user not found")).
					Times(1)
			},
			wantErr: true,
		},
		{
			name:        "異常系: 更新に失敗",
			userID:      uuid.New(),
			displayName: &updatedDisplayName,
			setupMock: func(m *mock.MockUserRepository, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), gomock.Eq(userID)).
					Return(newExistingUser(userID), nil).
					Times(1)
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("update failed")).
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

			uc := usecase.NewUserUseCase(mockRepo)

			got, err := uc.UpdateUser(context.Background(), tt.userID, tt.displayName, tt.profile, tt.twitterID, tt.githubID)

			if tt.wantErr {
				assert.Error(t, err)
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
