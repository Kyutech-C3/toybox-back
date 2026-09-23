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

func TestCommentUsecase_GetCommentsByWorkID(t *testing.T) {
	tests := []struct {
		name      string
		workID    uuid.UUID
		setupMock func(*mock.MockCommentRepository, uuid.UUID)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "正常系: コメント取得成功",
			workID: uuid.New(),
			setupMock: func(m *mock.MockCommentRepository, workID uuid.UUID) {
				expectedComments := []*entity.Comment{
					{
						ID:        uuid.New(),
						Content:   "Great work!",
						WorkID:    workID,
						UserID:    uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					{
						ID:        uuid.New(),
						Content:   "Nice!",
						WorkID:    workID,
						UserID:    uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return(expectedComments, nil).
					Times(1)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "正常系: コメントが0件",
			workID: uuid.New(),
			setupMock: func(m *mock.MockCommentRepository, workID uuid.UUID) {
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return([]*entity.Comment{}, nil).
					Times(1)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "異常系: リポジトリエラー",
			workID: uuid.New(),
			setupMock: func(m *mock.MockCommentRepository, workID uuid.UUID) {
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return(nil, errors.New("database connection failed")).
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

			mockRepo := mock.NewMockCommentRepository(ctrl)
			tt.setupMock(mockRepo, tt.workID)
			mockWorkRepo := mock.NewMockWorkRepository(ctrl)
			uc := usecase.NewCommentUsecase(mockRepo, mockWorkRepo, 30*time.Second)
			got, err := uc.GetCommentsByWorkID(context.Background(), tt.workID)

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

func TestCommentUsecase_CreateComment(t *testing.T) {
	workID := uuid.New()
	userID := uuid.New()
	otherWorkID := uuid.New()
	replyID := uuid.New()

	tests := []struct {
		name         string
		replyAt      string
		setupComment func(*mock.MockCommentRepository)
		setupWork    func(*mock.MockWorkRepository)
		wantErr      bool
		errIs        error
	}{
		{
			name:    "正常系: reply_atなしで作成成功",
			replyAt: "",
			setupWork: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					ExistsById(gomock.Any(), workID).
					Return(true, nil)
			},
			setupComment: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(&entity.Comment{})).
					DoAndReturn(func(_ context.Context, c *entity.Comment) (*entity.Comment, error) {
						return c, nil
					})
			},
			wantErr: false,
		},
		{
			name:    "正常系: 同一work内のコメントへの返信は成功",
			replyAt: replyID.String(),
			setupWork: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					ExistsById(gomock.Any(), workID).
					Return(true, nil)
			},
			setupComment: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					FindByID(gomock.Any(), replyID).
					Return(&entity.Comment{ID: replyID, WorkID: workID}, nil)
				m.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(&entity.Comment{})).
					DoAndReturn(func(_ context.Context, c *entity.Comment) (*entity.Comment, error) {
						return c, nil
					})
			},
			wantErr: false,
		},
		{
			name:    "異常系: workが存在しない",
			replyAt: "",
			setupWork: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					ExistsById(gomock.Any(), workID).
					Return(false, nil)
			},
			setupComment: func(m *mock.MockCommentRepository) {},
			wantErr:      true,
			errIs:        domainerrors.ErrWorkNotFound,
		},
		{
			name:    "異常系: workの存在確認でリポジトリエラー",
			replyAt: "",
			setupWork: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					ExistsById(gomock.Any(), workID).
					Return(false, errors.New("database error"))
			},
			setupComment: func(m *mock.MockCommentRepository) {},
			wantErr:      true,
		},
		{
			name:    "異常系: reply_atのUUID形式が不正",
			replyAt: "invalid-uuid",
			setupWork: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					ExistsById(gomock.Any(), workID).
					Return(true, nil)
			},
			setupComment: func(m *mock.MockCommentRepository) {},
			wantErr:      true,
		},
		{
			name:    "異常系: reply_at先のコメントが見つからない",
			replyAt: replyID.String(),
			setupWork: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					ExistsById(gomock.Any(), workID).
					Return(true, nil)
			},
			setupComment: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					FindByID(gomock.Any(), replyID).
					Return(nil, domainerrors.ErrCommentNotFound)
			},
			wantErr: true,
			errIs:   domainerrors.ErrCommentNotFound,
		},
		{
			name:    "異常系: reply_at先のコメントが別workのものである",
			replyAt: replyID.String(),
			setupWork: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					ExistsById(gomock.Any(), workID).
					Return(true, nil)
			},
			setupComment: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					FindByID(gomock.Any(), replyID).
					Return(&entity.Comment{ID: replyID, WorkID: otherWorkID}, nil)
			},
			wantErr: true,
			errIs:   domainerrors.ErrInvalidReplyAt,
		},
		{
			name:    "異常系: リポジトリのCreateエラー",
			replyAt: "",
			setupWork: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					ExistsById(gomock.Any(), workID).
					Return(true, nil)
			},
			setupComment: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(&entity.Comment{})).
					Return(nil, domainerrors.ErrFailedToCreateComment)
			},
			wantErr: true,
			errIs:   domainerrors.ErrFailedToCreateComment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCommentRepo := mock.NewMockCommentRepository(ctrl)
			mockWorkRepo := mock.NewMockWorkRepository(ctrl)
			tt.setupWork(mockWorkRepo)
			tt.setupComment(mockCommentRepo)

			uc := usecase.NewCommentUsecase(mockCommentRepo, mockWorkRepo, 30*time.Second)

			got, err := uc.CreateComment(context.Background(), "コメント", workID, userID, tt.replyAt)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, workID, got.WorkID)
				assert.Equal(t, tt.replyAt, got.ReplyAt)
			}
		})
	}
}
