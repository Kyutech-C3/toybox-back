package controller_test

import (
	"bytes"
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
	"github.com/simesaba80/toybox-back/pkg/echovalidator"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCommentController_GetCommentsByWorkID(t *testing.T) {
	workID := uuid.New()
	userID := uuid.New()

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
	forbiddenResponseBytes, _ := json.Marshal(map[string]string{"message": "この作品にはコメントできません"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name       string
		workID     string
		withAuth   bool
		userID     uuid.UUID
		setupMock  func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID)
		wantStatus int
		wantBody   []byte
	}{
		{
			name:     "正常系: 認証なしでコメント取得成功",
			workID:   workID.String(),
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					GetCommentsByWorkID(gomock.Any(), workID, userID).
					Return(mockComments, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name:     "正常系: 認証ありでコメント取得成功",
			workID:   workID.String(),
			withAuth: true,
			userID:   userID,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					GetCommentsByWorkID(gomock.Any(), workID, userID).
					Return(mockComments, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name:     "異常系: work_idが不正",
			workID:   "invalid-uuid",
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					GetCommentsByWorkID(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   badRequestResponseBytes,
		},
		{
			name:     "異常系: 閲覧権限がない作品",
			workID:   workID.String(),
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					GetCommentsByWorkID(gomock.Any(), workID, userID).
					Return(nil, domainerrors.ErrWorkNotViewable)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   forbiddenResponseBytes,
		},
		{
			name:     "異常系: Usecaseエラー",
			workID:   workID.String(),
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					GetCommentsByWorkID(gomock.Any(), workID, userID).
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
			tt.setupMock(mockCommentUsecase, tt.userID)

			commentController := controller.NewCommentController(mockCommentUsecase)
			e.GET("/works/:work_id/comments", func(c echo.Context) error {
				if tt.withAuth {
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
						UserID: tt.userID.String(),
					})
					c.Set("user", token)
				}
				return commentController.GetCommentsByWorkID(c)
			})

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
	validBody, _ := json.Marshal(schema.CreateCommentRequest{Content: "コメント"})

	successResponseBytes, _ := json.Marshal(schema.ToCreateCommentResponse(mockComment))
	invalidWorkIDResponseBytes, _ := json.Marshal(map[string]string{"message": "Invalid work ID format"})
	invalidBodyResponseBytes, _ := json.Marshal(map[string]string{"message": "Invalid request body"})
	workNotFoundResponseBytes, _ := json.Marshal(map[string]string{"message": "作品が見つかりませんでした"})
	forbiddenResponseBytes, _ := json.Marshal(map[string]string{"message": "この作品にはコメントできません"})
	invalidReplyAtResponseBytes, _ := json.Marshal(map[string]string{"message": "返信先のコメントが不正です"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name       string
		workID     string
		body       []byte
		withAuth   bool
		userID     uuid.UUID
		setupMock  func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID)
		wantStatus int
		wantBody   []byte
	}{
		{
			name:     "正常系: 認証なし（匿名）で作成成功",
			workID:   workID.String(),
			body:     validBody,
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					CreateComment(gomock.Any(), "コメント", workID, userID, "").
					Return(mockComment, nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   successResponseBytes,
		},
		{
			name:     "正常系: 認証ありで作成成功",
			workID:   workID.String(),
			body:     validBody,
			withAuth: true,
			userID:   userID,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
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
			withAuth:   false,
			userID:     uuid.Nil,
			setupMock:  func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidWorkIDResponseBytes,
		},
		{
			name:       "異常系: リクエストボディのバインド失敗",
			workID:     workID.String(),
			body:       []byte("invalid json"),
			withAuth:   false,
			userID:     uuid.Nil,
			setupMock:  func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidBodyResponseBytes,
		},
		{
			name:     "異常系: 指定したworkが存在しない",
			workID:   workID.String(),
			body:     validBody,
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					CreateComment(gomock.Any(), "コメント", workID, userID, "").
					Return(nil, domainerrors.ErrWorkNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   workNotFoundResponseBytes,
		},
		{
			name:     "異常系: 閲覧権限がない作品",
			workID:   workID.String(),
			body:     validBody,
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					CreateComment(gomock.Any(), "コメント", workID, userID, "").
					Return(nil, domainerrors.ErrWorkNotViewable)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   forbiddenResponseBytes,
		},
		{
			name:     "異常系: reply_atが不正",
			workID:   workID.String(),
			body:     validBody,
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
				mockCommentUsecase.EXPECT().
					CreateComment(gomock.Any(), "コメント", workID, userID, "").
					Return(nil, domainerrors.ErrInvalidReplyAt)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidReplyAtResponseBytes,
		},
		{
			name:     "異常系: Usecaseエラー",
			workID:   workID.String(),
			body:     validBody,
			withAuth: false,
			userID:   uuid.Nil,
			setupMock: func(mockCommentUsecase *mock.MockICommentUsecase, userID uuid.UUID) {
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
			tt.setupMock(mockCommentUsecase, tt.userID)

			commentController := controller.NewCommentController(mockCommentUsecase)
			e.POST("/works/:work_id/comments", func(c echo.Context) error {
				if tt.withAuth {
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
						UserID: tt.userID.String(),
					})
					c.Set("user", token)
				}
				return commentController.CreateComment(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/works/"+tt.workID+"/comments", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}
