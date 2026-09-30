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
	workID := uuid.New()
	ownerID := uuid.New()

	tests := []struct {
		name             string
		workID           uuid.UUID
		userID           uuid.UUID
		setupWorkMock    func(*mock.MockWorkRepository)
		setupCommentMock func(*mock.MockCommentRepository)
		wantCount        int
		wantErr          bool
		errIs            error
	}{
		{
			name:   "正常系: public作品は未認証でも取得成功",
			workID: workID,
			userID: uuid.Nil,
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
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
			workID: workID,
			userID: uuid.Nil,
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return([]*entity.Comment{}, nil).
					Times(1)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "異常系: コメント取得でリポジトリエラー",
			workID: workID,
			userID: uuid.Nil,
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return(nil, errors.New("database connection failed")).
					Times(1)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:   "正常系: private作品は認証済みユーザーなら取得成功",
			workID: workID,
			userID: uuid.New(),
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "private"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					FindByWorkID(gomock.Any(), gomock.Eq(workID)).
					Return([]*entity.Comment{}, nil).
					Times(1)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "異常系: private作品は未認証だと取得不可",
			workID: workID,
			userID: uuid.Nil,
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "private"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {},
			wantErr:          true,
			errIs:            domainerrors.ErrWorkNotViewable,
		},
		{
			name:   "異常系: draft作品は取得不可",
			workID: workID,
			userID: ownerID,
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, UserID: ownerID, Visibility: "draft"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {},
			wantErr:          true,
			errIs:            domainerrors.ErrWorkNotViewable,
		},
		{
			name:   "異常系: workが見つからない",
			workID: workID,
			userID: uuid.Nil,
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(nil, domainerrors.ErrWorkNotFound)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {},
			wantErr:          true,
			errIs:            domainerrors.ErrWorkNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockCommentRepository(ctrl)
			mockWorkRepo := mock.NewMockWorkRepository(ctrl)
			tt.setupWorkMock(mockWorkRepo)
			tt.setupCommentMock(mockRepo)

			uc := usecase.NewCommentUsecase(mockRepo, mockWorkRepo, 30*time.Second)
			got, err := uc.GetCommentsByWorkID(context.Background(), tt.workID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
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
	ownerID := uuid.New()
	otherWorkID := uuid.New()
	replyID := uuid.New()

	tests := []struct {
		name             string
		userID           uuid.UUID
		replyAt          string
		setupCommentMock func(*mock.MockCommentRepository)
		setupWorkMock    func(*mock.MockWorkRepository)
		wantErr          bool
		errIs            error
	}{
		{
			name:    "正常系: reply_atなしで作成成功",
			userID:  userID,
			replyAt: "",
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
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
			userID:  userID,
			replyAt: replyID.String(),
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
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
			name:    "正常系: private作品への認証済み投稿は成功",
			userID:  userID,
			replyAt: "",
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "private"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(&entity.Comment{})).
					DoAndReturn(func(_ context.Context, c *entity.Comment) (*entity.Comment, error) {
						return c, nil
					})
			},
			wantErr: false,
		},
		{
			name:    "異常系: private作品への匿名投稿は不可",
			userID:  uuid.Nil,
			replyAt: "",
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "private"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {},
			wantErr:          true,
			errIs:            domainerrors.ErrWorkNotViewable,
		},
		{
			name:    "異常系: draft作品は投稿不可",
			userID:  ownerID,
			replyAt: "",
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, UserID: ownerID, Visibility: "draft"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {},
			wantErr:          true,
			errIs:            domainerrors.ErrWorkNotViewable,
		},
		{
			name:    "異常系: workが存在しない",
			userID:  userID,
			replyAt: "",
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(nil, domainerrors.ErrWorkNotFound)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {},
			wantErr:          true,
			errIs:            domainerrors.ErrWorkNotFound,
		},
		{
			name:    "異常系: workの取得でリポジトリエラー",
			userID:  userID,
			replyAt: "",
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(nil, errors.New("database error"))
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {},
			wantErr:          true,
		},
		{
			name:    "異常系: reply_atのUUID形式が不正",
			userID:  userID,
			replyAt: "invalid-uuid",
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {},
			wantErr:          true,
		},
		{
			name:    "異常系: reply_at先のコメントが見つからない",
			userID:  userID,
			replyAt: replyID.String(),
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					FindByID(gomock.Any(), replyID).
					Return(nil, domainerrors.ErrCommentNotFound)
			},
			wantErr: true,
			errIs:   domainerrors.ErrCommentNotFound,
		},
		{
			name:    "異常系: reply_at先のコメントが別workのものである",
			userID:  userID,
			replyAt: replyID.String(),
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
				m.EXPECT().
					FindByID(gomock.Any(), replyID).
					Return(&entity.Comment{ID: replyID, WorkID: otherWorkID}, nil)
			},
			wantErr: true,
			errIs:   domainerrors.ErrInvalidReplyAt,
		},
		{
			name:    "異常系: リポジトリのCreateエラー",
			userID:  userID,
			replyAt: "",
			setupWorkMock: func(m *mock.MockWorkRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), workID).
					Return(&entity.Work{ID: workID, Visibility: "public"}, nil)
			},
			setupCommentMock: func(m *mock.MockCommentRepository) {
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
			tt.setupWorkMock(mockWorkRepo)
			tt.setupCommentMock(mockCommentRepo)

			uc := usecase.NewCommentUsecase(mockCommentRepo, mockWorkRepo, 30*time.Second)

			got, err := uc.CreateComment(context.Background(), "コメント", workID, tt.userID, tt.replyAt)

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
