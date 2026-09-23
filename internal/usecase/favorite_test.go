package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/usecase"
	"github.com/simesaba80/toybox-back/internal/usecase/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFavoriteUsecase_CreateFavorite(t *testing.T) {
	tests := []struct {
		name              string
		setupWorkMock     func(*mock.MockWorkRepository, uuid.UUID, uuid.UUID)
		setupFavoriteMock func(*mock.MockFavoriteRepository, uuid.UUID, uuid.UUID)
		wantErr           bool
		errIs             error
	}{
		{
			name: "正常系: 新規でいいねを作成できる",
			setupWorkMock: func(m *mock.MockWorkRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, UserID: uuid.New(), Visibility: "public"}, nil)
			},
			setupFavoriteMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					DoAndReturn(func(_ context.Context, fav *entity.Favorite) (bool, error) {
						assert.Equal(t, workID, fav.WorkID)
						assert.Equal(t, userID, fav.UserID)
						return false, nil
					})
				m.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					DoAndReturn(func(_ context.Context, fav *entity.Favorite) (*entity.Favorite, error) {
						assert.Equal(t, workID, fav.WorkID)
						assert.Equal(t, userID, fav.UserID)
						return fav, nil
					})
			},
			wantErr: false,
		},
		{
			name: "正常系: 自分のdraftにはいいねできる",
			setupWorkMock: func(m *mock.MockWorkRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, UserID: userID, Visibility: "draft"}, nil)
			},
			setupFavoriteMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(false, nil)
				m.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					DoAndReturn(func(_ context.Context, fav *entity.Favorite) (*entity.Favorite, error) {
						return fav, nil
					})
			},
			wantErr: false,
		},
		{
			name: "異常系: 作品が存在しない",
			setupWorkMock: func(m *mock.MockWorkRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(nil, domainerrors.ErrWorkNotFound)
			},
			setupFavoriteMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {},
			wantErr:           true,
			errIs:             domainerrors.ErrWorkNotFound,
		},
		{
			name: "異常系: 他人のdraftにはいいねできない",
			setupWorkMock: func(m *mock.MockWorkRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, UserID: uuid.New(), Visibility: "draft"}, nil)
			},
			setupFavoriteMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {},
			wantErr:           true,
			errIs:             domainerrors.ErrWorkNotViewable,
		},
		{
			name: "異常系: 既に存在するいいねは作成しない",
			setupWorkMock: func(m *mock.MockWorkRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, UserID: uuid.New(), Visibility: "public"}, nil)
			},
			setupFavoriteMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					DoAndReturn(func(_ context.Context, fav *entity.Favorite) (bool, error) {
						assert.Equal(t, workID, fav.WorkID)
						assert.Equal(t, userID, fav.UserID)
						return true, nil
					})
			},
			wantErr: true,
			errIs:   domainerrors.ErrFavoriteAlreadyExists,
		},
		{
			name: "異常系: リポジトリの作成エラーをラップして返す",
			setupWorkMock: func(m *mock.MockWorkRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, UserID: uuid.New(), Visibility: "public"}, nil)
			},
			setupFavoriteMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(false, nil)
				m.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(nil, domainerrors.ErrFailedToCreateFavorite)
			},
			wantErr: true,
			errIs:   domainerrors.ErrFailedToCreateFavorite,
		},
		{
			name: "異常系: 存在確認のリポジトリエラーをラップして返す",
			setupWorkMock: func(m *mock.MockWorkRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, UserID: uuid.New(), Visibility: "public"}, nil)
			},
			setupFavoriteMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(false, domainerrors.ErrFailedToCheckFavoriteExists)
			},
			wantErr: true,
			errIs:   domainerrors.ErrFailedToCheckFavoriteExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			workID := uuid.New()
			userID := uuid.New()

			mockFavoriteRepo := mock.NewMockFavoriteRepository(ctrl)
			tt.setupFavoriteMock(mockFavoriteRepo, workID, userID)
			mockWorkRepo := mock.NewMockWorkRepository(ctrl)
			tt.setupWorkMock(mockWorkRepo, workID, userID)

			uc := usecase.NewFavoriteUsecase(mockFavoriteRepo, mockWorkRepo)

			err := uc.CreateFavorite(context.Background(), workID, userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFavoriteUsecase_DeleteFavorite(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mock.MockFavoriteRepository, uuid.UUID, uuid.UUID)
		wantErr   bool
		errIs     error
	}{
		{
			name: "正常系: 既存のいいねを削除できる",
			setupMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(true, nil)
				m.EXPECT().
					Delete(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "異常系: いいねが存在しない場合は削除しない",
			setupMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(false, nil)
			},
			wantErr: true,
			errIs:   domainerrors.ErrFavoriteNotFound,
		},
		{
			name: "異常系: 削除時のリポジトリエラーをそのまま返す",
			setupMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(true, nil)
				m.EXPECT().
					Delete(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(domainerrors.ErrFailedToDeleteFavorite)
			},
			wantErr: true,
			errIs:   domainerrors.ErrFailedToDeleteFavorite,
		},
		{
			name: "異常系: 存在確認のリポジトリエラーをラップして返す",
			setupMock: func(m *mock.MockFavoriteRepository, workID, userID uuid.UUID) {
				m.EXPECT().
					Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
					Return(false, domainerrors.ErrFailedToCheckFavoriteExists)
			},
			wantErr: true,
			errIs:   domainerrors.ErrFailedToCheckFavoriteExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			workID := uuid.New()
			userID := uuid.New()

			mockFavoriteRepo := mock.NewMockFavoriteRepository(ctrl)
			tt.setupMock(mockFavoriteRepo, workID, userID)
			mockWorkRepo := mock.NewMockWorkRepository(ctrl)

			uc := usecase.NewFavoriteUsecase(mockFavoriteRepo, mockWorkRepo)

			err := uc.DeleteFavorite(context.Background(), workID, userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFavoriteUsecase_CountFavoritesByWorkID(t *testing.T) {
	workID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mock.NewMockFavoriteRepository(ctrl)

	gomock.InOrder(
		mockFavoriteRepo.EXPECT().
			CountByWorkID(gomock.Any(), workID).
			Return(3, nil),
		mockFavoriteRepo.EXPECT().
			CountByWorkID(gomock.Any(), workID).
			Return(0, domainerrors.ErrFailedToCountFavoritesByWorkID),
	)

	mockWorkRepo := mock.NewMockWorkRepository(ctrl)
	uc := usecase.NewFavoriteUsecase(mockFavoriteRepo, mockWorkRepo)

	total, err := uc.CountFavoritesByWorkID(context.Background(), workID)
	assert.NoError(t, err)
	assert.Equal(t, 3, total)

	total, err = uc.CountFavoritesByWorkID(context.Background(), workID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrFailedToCountFavoritesByWorkID)
	assert.Equal(t, 0, total)
}

func TestFavoriteUsecase_IsFavorite(t *testing.T) {
	workID := uuid.New()
	userID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFavoriteRepo := mock.NewMockFavoriteRepository(ctrl)
	gomock.InOrder(
		mockFavoriteRepo.EXPECT().
			Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
			Return(true, nil),
		mockFavoriteRepo.EXPECT().
			Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
			Return(false, nil),
		mockFavoriteRepo.EXPECT().
			Exists(gomock.Any(), gomock.AssignableToTypeOf(&entity.Favorite{})).
			Return(false, domainerrors.ErrFailedToCheckFavoriteExists),
	)

	mockWorkRepo := mock.NewMockWorkRepository(ctrl)
	uc := usecase.NewFavoriteUsecase(mockFavoriteRepo, mockWorkRepo)

	isFavorite, err := uc.IsFavorite(context.Background(), workID, userID)
	assert.NoError(t, err)
	assert.True(t, isFavorite)

	isFavorite, err = uc.IsFavorite(context.Background(), workID, userID)
	assert.NoError(t, err)
	assert.False(t, isFavorite)

	isFavorite, err = uc.IsFavorite(context.Background(), workID, userID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrFailedToCheckFavoriteExists)
	assert.False(t, isFavorite)
}
