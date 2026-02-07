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

func TestTagController_GetAllTags(t *testing.T) {
	now := time.Now()
	mockTags := []*entity.Tag{
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
	successResponseBytes, _ := json.Marshal(schema.ToTagListResponse(mockTags))
	emptyResponseBytes, _ := json.Marshal(schema.ToTagListResponse([]*entity.Tag{}))
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "タグの取得に失敗しました"})

	tests := []struct {
		name       string
		setupMock  func(mockTagUsecase *mock.MockITagUseCase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "正常系: タグ一覧取得成功",
			setupMock: func(mockTagUsecase *mock.MockITagUseCase) {
				mockTagUsecase.EXPECT().
					GetAll(gomock.Any()).
					Return(mockTags, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "正常系: タグが0件",
			setupMock: func(mockTagUsecase *mock.MockITagUseCase) {
				mockTagUsecase.EXPECT().
					GetAll(gomock.Any()).
					Return([]*entity.Tag{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   emptyResponseBytes,
		},
		{
			name: "異常系: Usecaseエラー",
			setupMock: func(mockTagUsecase *mock.MockITagUseCase) {
				mockTagUsecase.EXPECT().
					GetAll(gomock.Any()).
					Return(nil, domainerrors.ErrFailedToGetAllTags)
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

			mockUsecase := mock.NewMockITagUseCase(ctrl)
			tt.setupMock(mockUsecase)

			tagController := controller.NewTagController(mockUsecase)
			e.GET("/tags", tagController.GetAllTags)

			req := httptest.NewRequest(http.MethodGet, "/tags", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}

func TestTagController_CreateTag(t *testing.T) {
	now := time.Now()
	tagID := uuid.New()
	mockTag := &entity.Tag{
		ID:        tagID,
		Name:      "Go",
		CreatedAt: now,
		UpdatedAt: now,
	}

	input := schema.CreateTagInput{Name: "Go"}
	inputJSON, _ := json.Marshal(input)

	successResponseBytes, _ := json.Marshal(schema.ToTagDetailResponse(mockTag))
	badRequestResponseBytes, _ := json.Marshal(map[string]string{"message": "無効なリクエストボディです"})
	invalidTagNameResponseBytes, _ := json.Marshal(map[string]string{"message": "タグ名が無効です"})
	failedToCreateResponseBytes, _ := json.Marshal(map[string]string{"message": "タグの作成に失敗しました"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name       string
		body       []byte
		setupMock  func(mockTagUsecase *mock.MockITagUseCase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "正常系: タグ作成成功",
			body: inputJSON,
			setupMock: func(mockTagUsecase *mock.MockITagUseCase) {
				mockTagUsecase.EXPECT().
					Create(gomock.Any(), "Go").
					Return(mockTag, nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   successResponseBytes,
		},
		{
			name:       "異常系: 不正なリクエストボディ",
			body:       []byte("invalid json"),
			setupMock:  func(mockTagUsecase *mock.MockITagUseCase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   badRequestResponseBytes,
		},
		{
			name:       "異常系: タグ名が空（バリデーションエラー）",
			body:       []byte(`{"name": ""}`),
			setupMock:  func(mockTagUsecase *mock.MockITagUseCase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   badRequestResponseBytes,
		},
		{
			name: "異常系: タグ名が無効",
			body: inputJSON,
			setupMock: func(mockTagUsecase *mock.MockITagUseCase) {
				mockTagUsecase.EXPECT().
					Create(gomock.Any(), "Go").
					Return(nil, domainerrors.ErrInvalidTagName)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidTagNameResponseBytes,
		},
		{
			name: "異常系: タグ作成失敗",
			body: inputJSON,
			setupMock: func(mockTagUsecase *mock.MockITagUseCase) {
				mockTagUsecase.EXPECT().
					Create(gomock.Any(), "Go").
					Return(nil, domainerrors.ErrFailedToCreateTag)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   failedToCreateResponseBytes,
		},
		{
			name: "異常系: 予期しないエラー",
			body: inputJSON,
			setupMock: func(mockTagUsecase *mock.MockITagUseCase) {
				mockTagUsecase.EXPECT().
					Create(gomock.Any(), "Go").
					Return(nil, errors.New("unexpected error"))
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

			mockUsecase := mock.NewMockITagUseCase(ctrl)
			tt.setupMock(mockUsecase)

			tagController := controller.NewTagController(mockUsecase)
			e.POST("/auth/tags", tagController.CreateTag)

			req := httptest.NewRequest(http.MethodPost, "/auth/tags", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}
