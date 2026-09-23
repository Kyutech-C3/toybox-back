package usecase_test

import (
	"context"
	"errors"
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

func TestTagUseCase_Create(t *testing.T) {
	tests := []struct {
		name      string
		tagName   string
		setupMock func(*mock.MockTagRepository)
		wantErr   bool
		wantName  string
	}{
		{
			name:    "正常系: タグ作成成功",
			tagName: "Go",
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					ExistsByName(gomock.Any(), "go").
					Return(false, nil).
					Times(1)
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
						return tag, nil
					}).
					Times(1)
			},
			wantErr:  false,
			wantName: "go",
		},
		{
			name:      "異常系: タグ名が空",
			tagName:   "",
			setupMock: func(m *mock.MockTagRepository) {},
			wantErr:   true,
		},
		{
			name:    "異常系: リポジトリエラー",
			tagName: "Rust",
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					ExistsByName(gomock.Any(), "rust").
					Return(false, nil).
					Times(1)
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, domainerrors.ErrFailedToCreateTag).
					Times(1)
			},
			wantErr: true,
		},
		{
			name:    "異常系: 既に存在するタグ名",
			tagName: "Go",
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					ExistsByName(gomock.Any(), "go").
					Return(true, nil).
					Times(1)
			},
			wantErr: true,
		},
		{
			name:    "異常系: 存在確認のリポジトリエラー",
			tagName: "Go",
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					ExistsByName(gomock.Any(), "go").
					Return(false, domainerrors.ErrFailedToCheckTagExists).
					Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockTagRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewTagUseCase(mockRepo)

			got, err := uc.Create(context.Background(), tt.tagName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantName, got.Name)
				assert.NotEqual(t, uuid.Nil, got.ID)
			}
		})
	}
}

func TestTagUseCase_GetAll(t *testing.T) {
	now := time.Now()
	tag1ID := uuid.New()
	tag2ID := uuid.New()
	tests := []struct {
		name          string
		authenticated bool
		setupMock     func(*mock.MockTagRepository)
		wantCount     int
		wantWorkCount map[uuid.UUID]int
		wantErr       bool
	}{
		{
			name:          "正常系: タグ一覧取得成功（未認証、作品数がマージされる）",
			authenticated: false,
			setupMock: func(m *mock.MockTagRepository) {
				expectedTags := []*entity.Tag{
					{ID: tag1ID, Name: "Go", CreatedAt: now, UpdatedAt: now},
					{ID: tag2ID, Name: "Rust", CreatedAt: now, UpdatedAt: now},
				}
				m.EXPECT().
					FindAll(gomock.Any()).
					Return(expectedTags, nil).
					Times(1)
				m.EXPECT().
					CountWorksByTag(gomock.Any(), false).
					Return(map[uuid.UUID]int{tag1ID: 3}, nil).
					Times(1)
			},
			wantCount:     2,
			wantWorkCount: map[uuid.UUID]int{tag1ID: 3, tag2ID: 0},
			wantErr:       false,
		},
		{
			name:          "正常系: 認証済みの場合CountWorksByTagにtrueが渡る",
			authenticated: true,
			setupMock: func(m *mock.MockTagRepository) {
				expectedTags := []*entity.Tag{
					{ID: tag1ID, Name: "Go", CreatedAt: now, UpdatedAt: now},
				}
				m.EXPECT().
					FindAll(gomock.Any()).
					Return(expectedTags, nil).
					Times(1)
				m.EXPECT().
					CountWorksByTag(gomock.Any(), true).
					Return(map[uuid.UUID]int{tag1ID: 5}, nil).
					Times(1)
			},
			wantCount:     1,
			wantWorkCount: map[uuid.UUID]int{tag1ID: 5},
			wantErr:       false,
		},
		{
			name:          "正常系: タグが0件",
			authenticated: false,
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					FindAll(gomock.Any()).
					Return([]*entity.Tag{}, nil).
					Times(1)
				m.EXPECT().
					CountWorksByTag(gomock.Any(), false).
					Return(map[uuid.UUID]int{}, nil).
					Times(1)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:          "異常系: FindAllのリポジトリエラー",
			authenticated: false,
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					FindAll(gomock.Any()).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:          "異常系: CountWorksByTagのリポジトリエラー",
			authenticated: false,
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					FindAll(gomock.Any()).
					Return([]*entity.Tag{{ID: tag1ID, Name: "Go", CreatedAt: now, UpdatedAt: now}}, nil).
					Times(1)
				m.EXPECT().
					CountWorksByTag(gomock.Any(), false).
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

			mockRepo := mock.NewMockTagRepository(ctrl)
			tt.setupMock(mockRepo)

			uc := usecase.NewTagUseCase(mockRepo)

			got, err := uc.GetAll(context.Background(), tt.authenticated)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Len(t, got, tt.wantCount)
				for _, tag := range got {
					assert.Equal(t, tt.wantWorkCount[tag.ID], tag.WorkCount)
				}
			}
		})
	}
}
