package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/interface/controller/mock"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCommentController_GetCommentsByWorkID(t *testing.T) {
	workID := uuid.New()

	mockComments := []*entity.Comment{
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			WorkID:    workID,
			Content:   "コメント",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	successResponseBytes, _ := json.Marshal(schema.ToCommentListResponse(mockComments))
	badRequestResponseBytes, _ := json.Marshal(map[string]string{"message": "Invalid work ID format"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name       string
		workID     string
		setupMock  func(mockCommentUsecase *mock.MockICommentUsecase, mockWorkUsecase *mock.MockIWorkUseCase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name:   "正常系: コメント取得成功",
			workID: workID.String(),
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, mockWorkUsecase *mock.MockIWorkUseCase) {
				mockCommentUsecase.EXPECT().
					GetCommentsByWorkID(gomock.Any(), workID).
					Return(mockComments, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name:   "異常系: work_idが不正",
			workID: "invalid-uuid",
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, mockWorkUsecase *mock.MockIWorkUseCase) {
				mockCommentUsecase.EXPECT().
					GetCommentsByWorkID(gomock.Any(), workID).
					Return(nil, errors.New("some db error")).
					Times(0)
			},

			wantStatus: http.StatusBadRequest,
			wantBody:   badRequestResponseBytes,
		},
		{
			name:   "異常系: Usecaseエラー",
			workID: workID.String(),
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, mockWorkUsecase *mock.MockIWorkUseCase) {
				mockCommentUsecase.EXPECT().
					GetCommentsByWorkID(gomock.Any(), workID).
					Return(nil, errors.New("some error"))
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

			mockCommentUsecase := mock.NewMockICommentUsecase(ctrl)
			mockWorkUsecase := mock.NewMockIWorkUseCase(ctrl)
			tt.setupMock(mockCommentUsecase, mockWorkUsecase)

			commentController := controller.NewCommentController(mockCommentUsecase)
			e.GET("/works/:work_id/comments", commentController.GetCommentsByWorkID)

			req := httptest.NewRequest(http.MethodGet, "/works/"+tt.workID+"/comments", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}

func TestCommentController_DeleteComment(t *testing.T) {
	commentID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		userID     string
		commentID  string
		setupMock  func(*mock.MockICommentUsecase)
		wantStatus int
		wantBody   string
		wantJSON   bool
	}{
		{
			name:      "正常系: コメント削除成功",
			userID:    userID.String(),
			commentID: commentID.String(),
			setupMock: func(m *mock.MockICommentUsecase) {
				m.EXPECT().
					DeleteComment(gomock.Any(), commentID, userID).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
		{
			name:      "異常系: ユーザーIDがUUID形式でない",
			userID:    "invalid-uuid",
			commentID: commentID.String(),
			setupMock: func(m *mock.MockICommentUsecase) {
				m.EXPECT().
					DeleteComment(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid user ID"}`,
			wantJSON:   true,
		},
		{
			name:      "異常系: comment_idがUUID形式でない",
			userID:    userID.String(),
			commentID: "invalid-comment-id",
			setupMock: func(m *mock.MockICommentUsecase) {
				m.EXPECT().
					DeleteComment(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid comment ID format"}`,
			wantJSON:   true,
		},
		{
			name:      "異常系: 本人のコメントでない",
			userID:    userID.String(),
			commentID: commentID.String(),
			setupMock: func(m *mock.MockICommentUsecase) {
				m.EXPECT().
					DeleteComment(gomock.Any(), commentID, userID).
					Return(domainerrors.ErrCommentNotOwnedByUser)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   `{"message":"このコメントを削除する権限がありません"}`,
			wantJSON:   true,
		},
		{
			name:      "異常系: コメントが見つからない",
			userID:    userID.String(),
			commentID: commentID.String(),
			setupMock: func(m *mock.MockICommentUsecase) {
				m.EXPECT().
					DeleteComment(gomock.Any(), commentID, userID).
					Return(domainerrors.ErrCommentNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"message":"コメントが見つかりませんでした"}`,
			wantJSON:   true,
		},
		{
			name:      "異常系: 予期しないエラー",
			userID:    userID.String(),
			commentID: commentID.String(),
			setupMock: func(m *mock.MockICommentUsecase) {
				m.EXPECT().
					DeleteComment(gomock.Any(), commentID, userID).
					Return(errors.New("unexpected error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"message":"サーバーエラーが発生しました"}`,
			wantJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCommentUsecase := mock.NewMockICommentUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockCommentUsecase)
			}

			commentController := controller.NewCommentController(mockCommentUsecase)

			e.DELETE("/auth/comments/:comment_id", func(c echo.Context) error {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
					UserID: tt.userID,
				})
				c.Set("user", token)
				return commentController.DeleteComment(c)
			})

			req := httptest.NewRequest(http.MethodDelete, "/auth/comments/"+tt.commentID, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantJSON {
				assert.JSONEq(t, tt.wantBody, rec.Body.String())
			} else {
				assert.Equal(t, tt.wantBody, rec.Body.String())
			}
		})
	}
}
