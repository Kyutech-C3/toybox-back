package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/interface/controller/mock"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/pkg/echovalidator"
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

func TestCommentController_CreateComment(t *testing.T) {
	workID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	mockComment := &entity.Comment{
		ID:        uuid.New(),
		Content:   "コメント",
		WorkID:    workID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	validBody, _ := json.Marshal(schema.CreateCommentRequest{Content: "コメント", UserID: userID.String()})

	successResponseBytes, _ := json.Marshal(schema.ToCreateCommentResponse(mockComment))
	invalidWorkIDResponseBytes, _ := json.Marshal(map[string]string{"message": "Invalid work ID format"})
	invalidBodyResponseBytes, _ := json.Marshal(map[string]string{"message": "Invalid request body"})
	workNotFoundResponseBytes, _ := json.Marshal(map[string]string{"message": "作品が見つかりませんでした"})
	invalidReplyAtResponseBytes, _ := json.Marshal(map[string]string{"message": "返信先のコメントが不正です"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name       string
		workID     string
		body       []byte
		setupMock  func(mockCommentUsecase *mock.MockICommentUsecase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name:   "正常系: コメント作成成功",
			workID: workID.String(),
			body:   validBody,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase) {
				mockCommentUsecase.EXPECT().
					CreateComment(gomock.Any(), "コメント", workID, userID, "").
					Return(mockComment, nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   successResponseBytes,
		},
		{
			name:       "異常系: work_idが不正",
			workID:     "invalid-uuid",
			body:       validBody,
			setupMock:  func(mockCommentUsecase *mock.MockICommentUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidWorkIDResponseBytes,
		},
		{
			name:       "異常系: リクエストボディのバインド失敗",
			workID:     workID.String(),
			body:       []byte("invalid json"),
			setupMock:  func(mockCommentUsecase *mock.MockICommentUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidBodyResponseBytes,
		},
		{
			name:   "異常系: 指定したworkが存在しない",
			workID: workID.String(),
			body:   validBody,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase) {
				mockCommentUsecase.EXPECT().
					CreateComment(gomock.Any(), "コメント", workID, userID, "").
					Return(nil, domainerrors.ErrWorkNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   workNotFoundResponseBytes,
		},
		{
			name:   "異常系: reply_atが不正",
			workID: workID.String(),
			body:   validBody,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase) {
				mockCommentUsecase.EXPECT().
					CreateComment(gomock.Any(), "コメント", workID, userID, "").
					Return(nil, domainerrors.ErrInvalidReplyAt)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidReplyAtResponseBytes,
		},
		{
			name:   "異常系: Usecaseエラー",
			workID: workID.String(),
			body:   validBody,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase) {
				mockCommentUsecase.EXPECT().
					CreateComment(gomock.Any(), "コメント", workID, userID, "").
					Return(nil, errors.New("some error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   internalErrorResponseBytes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = echovalidator.NewValidator()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCommentUsecase := mock.NewMockICommentUsecase(ctrl)
			tt.setupMock(mockCommentUsecase)

			commentController := controller.NewCommentController(mockCommentUsecase)
			e.POST("/works/:work_id/comments", commentController.CreateComment)

			req := httptest.NewRequest(http.MethodPost, "/works/"+tt.workID+"/comments", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}
