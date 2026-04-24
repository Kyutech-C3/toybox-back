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
					Create(gomock.Any(), gomock.Any()).
					Return(nil, domainerrors.ErrFailedToCreateTag).
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
	tests := []struct {
		name      string
		setupMock func(*mock.MockTagRepository)
		wantCount int
		wantErr   bool
	}{
		{
			name: "正常系: タグ一覧取得成功",
			setupMock: func(m *mock.MockTagRepository) {
				expectedTags := []*entity.Tag{
					{
						ID:        uuid.New(),
						Name:      "Go",
						CreatedAt: now,
						UpdatedAt: now,
					},
					{
						ID:        uuid.New(),
						Name:      "Rust",
						CreatedAt: now,
						UpdatedAt: now,
					},
				}
				m.EXPECT().
					FindAll(gomock.Any()).
					Return(expectedTags, nil).
					Times(1)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "正常系: タグが0件",
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					FindAll(gomock.Any()).
					Return([]*entity.Tag{}, nil).
					Times(1)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "異常系: リポジトリエラー",
			setupMock: func(m *mock.MockTagRepository) {
				m.EXPECT().
					FindAll(gomock.Any()).
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

			got, err := uc.GetAll(context.Background())

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
