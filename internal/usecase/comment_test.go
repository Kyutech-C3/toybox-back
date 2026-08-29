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
		{
			name:   "正常系: 削除済みで子孫がないコメントは除外される",
			workID: uuid.New(),
			setupMock: func(m *mock.MockCommentRepository, workID uuid.UUID) {
				comments := []*entity.Comment{
					{ID: uuid.New(), WorkID: workID, Status: "deleted"},
					{ID: uuid.New(), WorkID: workID, Status: "active"},
				}
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return(comments, nil).
					Times(1)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:   "正常系: 削除済みでも表示可能な子を持つ場合は残る",
			workID: uuid.New(),
			setupMock: func(m *mock.MockCommentRepository, workID uuid.UUID) {
				parentID := uuid.New()
				comments := []*entity.Comment{
					{ID: parentID, WorkID: workID, Status: "deleted"},
					{ID: uuid.New(), WorkID: workID, Status: "active", ReplyAt: parentID.String()},
				}
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return(comments, nil).
					Times(1)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "正常系: 削除済みの子しか持たない削除済み親は連鎖的に除外される",
			workID: uuid.New(),
			setupMock: func(m *mock.MockCommentRepository, workID uuid.UUID) {
				parentID := uuid.New()
				comments := []*entity.Comment{
					{ID: parentID, WorkID: workID, Status: "deleted"},
					{ID: uuid.New(), WorkID: workID, Status: "deleted", ReplyAt: parentID.String()},
				}
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return(comments, nil).
					Times(1)
			},
			wantCount: 0,
			wantErr:   false,
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

func TestCommentUsecase_DeleteComment(t *testing.T) {
	commentID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name      string
		commentID uuid.UUID
		userID    uuid.UUID
		setupMock func(*mock.MockCommentRepository, uuid.UUID, uuid.UUID)
		wantErr   bool
	}{
		{
			name:      "正常系: 削除成功",
			commentID: commentID,
			userID:    ownerID,
			setupMock: func(m *mock.MockCommentRepository, commentID, userID uuid.UUID) {
				m.EXPECT().
					FindByID(gomock.Any(), commentID).
					Return(&entity.Comment{ID: commentID, UserID: ownerID, Status: "active"}, nil)
				m.EXPECT().
					Delete(gomock.Any(), commentID, userID).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "異常系: コメント取得に失敗",
			commentID: commentID,
			userID:    ownerID,
			setupMock: func(m *mock.MockCommentRepository, commentID, userID uuid.UUID) {
				m.EXPECT().
					FindByID(gomock.Any(), commentID).
					Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:      "異常系: 既に削除済み",
			commentID: commentID,
			userID:    ownerID,
			setupMock: func(m *mock.MockCommentRepository, commentID, userID uuid.UUID) {
				m.EXPECT().
					FindByID(gomock.Any(), commentID).
					Return(&entity.Comment{ID: commentID, UserID: ownerID, Status: "deleted"}, nil)
			},
			wantErr: true,
		},
		{
			name:      "異常系: 本人のコメントでない",
			commentID: commentID,
			userID:    otherUserID,
			setupMock: func(m *mock.MockCommentRepository, commentID, userID uuid.UUID) {
				m.EXPECT().
					FindByID(gomock.Any(), commentID).
					Return(&entity.Comment{ID: commentID, UserID: ownerID, Status: "active"}, nil)
			},
			wantErr: true,
		},
		{
			name:      "異常系: リポジトリの削除に失敗",
			commentID: commentID,
			userID:    ownerID,
			setupMock: func(m *mock.MockCommentRepository, commentID, userID uuid.UUID) {
				m.EXPECT().
					FindByID(gomock.Any(), commentID).
					Return(&entity.Comment{ID: commentID, UserID: ownerID, Status: "active"}, nil)
				m.EXPECT().
					Delete(gomock.Any(), commentID, userID).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockCommentRepository(ctrl)
			tt.setupMock(mockRepo, tt.commentID, tt.userID)
			mockWorkRepo := mock.NewMockWorkRepository(ctrl)
			uc := usecase.NewCommentUsecase(mockRepo, mockWorkRepo, 30*time.Second)

			err := uc.DeleteComment(context.Background(), tt.commentID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
