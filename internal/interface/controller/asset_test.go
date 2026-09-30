package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/interface/controller/mock"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newAssetUploadRequest(t *testing.T, path string, includeFile bool) *http.Request {
	return newAssetUploadRequestWithContent(t, path, includeFile, []byte("dummy data"))
}

func newAssetUploadRequestWithContent(t *testing.T, path string, includeFile bool, content []byte) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if includeFile {
		part, err := writer.CreateFormFile("file", "test.png")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("failed to write file content: %v", err)
		}
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	return req
}

func TestAssetController_UploadAsset(t *testing.T) {
	assetID := uuid.New()
	successResponseBytes, _ := json.Marshal(schema.ToUploadAssetResponse(&entity.Asset{
		ID:  assetID,
		URL: "https://example.com/assets/test.png",
	}))
	fileRequiredResponseBytes, _ := json.Marshal(map[string]string{"message": "File is required"})
	invalidRequestResponseBytes, _ := json.Marshal(map[string]string{"message": "無効なリクエストです"})
	invalidFileNameResponseBytes, _ := json.Marshal(map[string]string{"message": "ファイル名が無効です"})
	unsupportedFileTypeResponseBytes, _ := json.Marshal(map[string]string{"message": "対応していないファイル形式です"})
	failedUploadResponseBytes, _ := json.Marshal(map[string]string{"message": "ファイルのアップロードに失敗しました"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMock     func(mockAssetUsecase *mock.MockIAssetUseCase, userID uuid.UUID)
		request       func(t *testing.T) *http.Request
		wantStatus    int
		wantBody      []byte
		expectSuccess bool
	}{
		{
			name:   "正常系: アセットアップロード成功",
			userID: uuid.New(),
			setupMock: func(mockAssetUsecase *mock.MockIAssetUseCase, userID uuid.UUID) {
				mockAssetUsecase.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), userID).
					DoAndReturn(func(ctx context.Context, file *multipart.FileHeader, uid uuid.UUID) (*entity.Asset, error) {
						assert.Equal(t, userID, uid)
						return &entity.Asset{
							ID:  assetID,
							URL: "https://example.com/assets/test.png",
						}, nil
					})
			},
			request: func(t *testing.T) *http.Request {
				return newAssetUploadRequest(t, "/works/asset", true)
			},
			wantStatus:    http.StatusOK,
			wantBody:      successResponseBytes,
			expectSuccess: true,
		},
		{
			name:   "異常系: ファイル未指定",
			userID: uuid.New(),
			setupMock: func(mockAssetUsecase *mock.MockIAssetUseCase, userID uuid.UUID) {
				mockAssetUsecase.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			request: func(t *testing.T) *http.Request {
				return newAssetUploadRequest(t, "/works/asset", false)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   fileRequiredResponseBytes,
		},
		{
			name:   "異常系: 無効なリクエスト",
			userID: uuid.New(),
			setupMock: func(mockAssetUsecase *mock.MockIAssetUseCase, userID uuid.UUID) {
				mockAssetUsecase.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), userID).
					Return(nil, domainerrors.ErrInvalidRequestBody)
			},
			request: func(t *testing.T) *http.Request {
				return newAssetUploadRequest(t, "/works/asset", true)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidRequestResponseBytes,
		},
		{
			name:   "異常系: 無効なファイル名",
			userID: uuid.New(),
			setupMock: func(mockAssetUsecase *mock.MockIAssetUseCase, userID uuid.UUID) {
				mockAssetUsecase.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), userID).
					Return(nil, domainerrors.ErrInvalidFileName)
			},
			request: func(t *testing.T) *http.Request {
				return newAssetUploadRequest(t, "/works/asset", true)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   invalidFileNameResponseBytes,
		},
		{
			name:   "異常系: 未対応のファイル形式",
			userID: uuid.New(),
			setupMock: func(mockAssetUsecase *mock.MockIAssetUseCase, userID uuid.UUID) {
				mockAssetUsecase.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), userID).
					Return(nil, domainerrors.ErrUnsupportedFileType)
			},
			request: func(t *testing.T) *http.Request {
				return newAssetUploadRequest(t, "/works/asset", true)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   unsupportedFileTypeResponseBytes,
		},
		{
			name:   "異常系: アップロード失敗",
			userID: uuid.New(),
			setupMock: func(mockAssetUsecase *mock.MockIAssetUseCase, userID uuid.UUID) {
				mockAssetUsecase.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), userID).
					Return(nil, domainerrors.ErrFailedToUploadFile)
			},
			request: func(t *testing.T) *http.Request {
				return newAssetUploadRequest(t, "/works/asset", true)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   failedUploadResponseBytes,
		},
		{
			name:   "異常系: 想定外のエラー",
			userID: uuid.New(),
			setupMock: func(mockAssetUsecase *mock.MockIAssetUseCase, userID uuid.UUID) {
				mockAssetUsecase.EXPECT().
					UploadFile(gomock.Any(), gomock.Any(), userID).
					Return(nil, errors.New("unexpected error"))
			},
			request: func(t *testing.T) *http.Request {
				return newAssetUploadRequest(t, "/works/asset", true)
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

			mockAssetUsecase := mock.NewMockIAssetUseCase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockAssetUsecase, tt.userID)
			}

			assetController := controller.NewAssetController(mockAssetUsecase)
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
				UserID: tt.userID.String(),
			})

			e.POST("/works/asset", func(c echo.Context) error {
				c.Set("user", token)
				return assetController.UploadAsset(c)
			})

			req := tt.request(t)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}

func TestAssetController_UploadAsset_BodyLimitWithoutContentLength(t *testing.T) {
	e := echo.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAssetUsecase := mock.NewMockIAssetUseCase(ctrl)
	mockAssetUsecase.EXPECT().
		UploadFile(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	assetController := controller.NewAssetController(mockAssetUsecase)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
		UserID: uuid.NewString(),
	})

	e.POST("/works/asset", func(c echo.Context) error {
		c.Set("user", token)
		return assetController.UploadAsset(c)
	}, middleware.BodyLimit("1KiB"))

	req := newAssetUploadRequestWithContent(t, "/works/asset", true, bytes.Repeat([]byte("a"), 8*1024))
	req.ContentLength = -1
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.JSONEq(t, `{"message":"Request Entity Too Large"}`, rec.Body.String())
}
