package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/interface/controller/mock"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/pkg/echovalidator"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserController_GetAllUsers(t *testing.T) {
	mockUser := &entity.User{ID: uuid.New(), Name: "testuser"}
	successResponseBytes, _ := json.Marshal(schema.UserListResponse{
		Users: []schema.GetUserOutput{schema.ToUserResponse(mockUser)},
	})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name       string
		setupMock  func(mockUserUsecase *mock.MockIUserUseCase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "正常系",
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					GetAllUser(gomock.Any()).
					Return([]*entity.User{mockUser}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "異常系: Usecaseエラー",
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					GetAllUser(gomock.Any()).
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

			mockUsecase := mock.NewMockIUserUseCase(ctrl)
			tt.setupMock(mockUsecase)

			userController := controller.NewUserController(mockUsecase)
			e.GET("/users", userController.GetAllUsers)

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}

func TestUserController_GetUserByID(t *testing.T) {
	mockUser := &entity.User{ID: uuid.New(), Name: "testuser"}
	successResponseBytes, _ := json.Marshal(schema.ToUserResponse(mockUser))
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name       string
		setupMock  func(mockUserUsecase *mock.MockIUserUseCase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "正常系",
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					GetByUserID(gomock.Any(), gomock.Eq(mockUser.ID)).
					Return(mockUser, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "異常系: Usecaseエラー",
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					GetByUserID(gomock.Any(), gomock.Eq(mockUser.ID)).
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

			mockUsecase := mock.NewMockIUserUseCase(ctrl)
			tt.setupMock(mockUsecase)

			userController := controller.NewUserController(mockUsecase)
			e.GET("/users/:id", userController.GetUserByID)

			req := httptest.NewRequest(http.MethodGet, "/users/"+mockUser.ID.String(), nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}

func TestUserController_GetCurrentUser(t *testing.T) {
	userID := uuid.New()
	mockUser := &entity.User{
		ID:          userID,
		Name:        "testuser",
		DisplayName: "Test User",
		AvatarURL:   "https://example.com/icon.png",
	}
	successResponseBytes, _ := json.Marshal(schema.ToCurrentUserResponse(mockUser))
	notFoundResponseBytes, _ := json.Marshal(map[string]string{"message": "ユーザーが見つかりませんでした"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})
	tests := []struct {
		name       string
		setupMock  func(mockUserUsecase *mock.MockIUserUseCase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "正常系",
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					GetByUserID(gomock.Any(), gomock.Eq(userID)).
					Return(mockUser, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "異常系: ユーザーが見つからない",
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					GetByUserID(gomock.Any(), gomock.Eq(userID)).
					Return(nil, domainerrors.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   notFoundResponseBytes,
		},
		{
			name: "異常系: Usecaseエラー",
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					GetByUserID(gomock.Any(), gomock.Eq(userID)).
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

			mockUsecase := mock.NewMockIUserUseCase(ctrl)
			tt.setupMock(mockUsecase)

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
				UserID: userID.String(),
			})

			userController := controller.NewUserController(mockUsecase)
			e.GET("/auth/users/me", func(c echo.Context) error {
				c.Set("user", token)
				return userController.GetCurrentUser(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/auth/users/me", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}

type updateUserAvatar struct {
	filename string
	content  []byte
}

func newUpdateUserRequest(t *testing.T, fields map[string]string, avatar *updateUserAvatar) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("failed to write field: %v", err)
		}
	}
	if avatar != nil {
		part, err := writer.CreateFormFile("avatar", avatar.filename)
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		if _, err := part.Write(avatar.content); err != nil {
			t.Fatalf("failed to write file content: %v", err)
		}
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/auth/user", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	return req
}

func TestUserController_UpdateUser(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	mockUser := &entity.User{
		ID:          userID,
		Name:        "testuser",
		Email:       "updated@example.com",
		DisplayName: "Updated User",
		Profile:     "Updated profile",
		AvatarURL:   "https://s3.example.com/bucket/dir/avatar/new.png",
		XUsername:   "xuser123",
		GithubID:    "github123",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	updatedDisplayName := "Updated User"
	updatedProfile := "Updated profile"
	updatedXUsername := "xuser123"
	updatedGithubID := "github123"
	fullFields := map[string]string{
		"display_name": updatedDisplayName,
		"profile":      updatedProfile,
		"x_username":   updatedXUsername,
		"github_id":    updatedGithubID,
	}
	avatar := &updateUserAvatar{filename: "avatar.png", content: []byte("dummy image content")}

	successResponseBytes, _ := json.Marshal(schema.ToUserResponse(mockUser))
	badRequestResponseBytes, _ := json.Marshal(map[string]string{"message": "無効なリクエストです"})
	notFoundResponseBytes, _ := json.Marshal(map[string]string{"message": "ユーザーが見つかりませんでした"})
	failedToUpdateResponseBytes, _ := json.Marshal(map[string]string{"message": "ユーザーの更新に失敗しました"})
	unsupportedFileTypeResponseBytes, _ := json.Marshal(map[string]string{"message": "対応していないファイル形式です"})
	failedToUploadResponseBytes, _ := json.Marshal(map[string]string{"message": "ファイルのアップロードに失敗しました"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	partialDisplayName := "New Name Only"

	tests := []struct {
		name       string
		newRequest func(t *testing.T) *http.Request
		setupMock  func(mockUserUsecase *mock.MockIUserUseCase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "正常系: ユーザー更新成功",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, fullFields, nil)
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, &updatedDisplayName, &updatedProfile, &updatedXUsername, &updatedGithubID, (*multipart.FileHeader)(nil)).
					Return(mockUser, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "正常系: アバターと他フィールドを同時に更新",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, fullFields, avatar)
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, &updatedDisplayName, &updatedProfile, &updatedXUsername, &updatedGithubID, gomock.Any()).
					DoAndReturn(func(ctx any, uid uuid.UUID, displayName, profile, xUsername, githubID *string, file *multipart.FileHeader) (*entity.User, error) {
						assert.NotNil(t, file)
						assert.Equal(t, avatar.filename, file.Filename)
						assert.Equal(t, int64(len(avatar.content)), file.Size)
						return mockUser, nil
					})
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "正常系: アバターのみ更新（他フィールドは省略）",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, nil, avatar)
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil), gomock.Not(gomock.Nil())).
					Return(mockUser, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "正常系: display_nameのみ更新（部分更新、他フィールドは省略）",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, map[string]string{"display_name": partialDisplayName}, nil)
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, &partialDisplayName, (*string)(nil), (*string)(nil), (*string)(nil), (*multipart.FileHeader)(nil)).
					Return(mockUser, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name: "異常系: multipart/form-data以外のリクエスト",
			newRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPatch, "/auth/user", bytes.NewReader([]byte(`{"display_name": "json"}`)))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				return req
			},
			setupMock:  func(mockUserUsecase *mock.MockIUserUseCase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   badRequestResponseBytes,
		},
		{
			name: "異常系: display_nameが空文字（バリデーションエラー）",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, map[string]string{"display_name": ""}, nil)
			},
			setupMock:  func(mockUserUsecase *mock.MockIUserUseCase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   badRequestResponseBytes,
		},
		{
			name: "異常系: 対応していないアバターの形式",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, nil, &updateUserAvatar{filename: "avatar.mp4", content: []byte("dummy")})
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil), gomock.Any()).
					Return(nil, domainerrors.ErrUnsupportedFileType)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   unsupportedFileTypeResponseBytes,
		},
		{
			name: "異常系: アバターのアップロード失敗",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, nil, avatar)
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil), gomock.Any()).
					Return(nil, domainerrors.ErrFailedToUploadFile)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   failedToUploadResponseBytes,
		},
		{
			name: "異常系: ユーザーが見つからない",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, fullFields, nil)
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, &updatedDisplayName, &updatedProfile, &updatedXUsername, &updatedGithubID, (*multipart.FileHeader)(nil)).
					Return(nil, domainerrors.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   notFoundResponseBytes,
		},
		{
			name: "異常系: ユーザー更新失敗",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, fullFields, nil)
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, &updatedDisplayName, &updatedProfile, &updatedXUsername, &updatedGithubID, (*multipart.FileHeader)(nil)).
					Return(nil, domainerrors.ErrFailedToUpdateUser)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   failedToUpdateResponseBytes,
		},
		{
			name: "異常系: 予期しないエラー",
			newRequest: func(t *testing.T) *http.Request {
				return newUpdateUserRequest(t, fullFields, nil)
			},
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, &updatedDisplayName, &updatedProfile, &updatedXUsername, &updatedGithubID, (*multipart.FileHeader)(nil)).
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

			mockUsecase := mock.NewMockIUserUseCase(ctrl)
			tt.setupMock(mockUsecase)

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
				UserID: userID.String(),
			})

			userController := controller.NewUserController(mockUsecase)
			e.PATCH("/auth/user", func(c echo.Context) error {
				c.Set("user", token)
				return userController.UpdateUser(c)
			})

			req := tt.newRequest(t)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}

func TestUserController_UpdateUser_BodyLimitWithoutContentLength(t *testing.T) {
	e := echo.New()
	e.Validator = echovalidator.NewValidator()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockIUserUseCase(ctrl)
	mockUsecase.EXPECT().
		UpdateUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	userController := controller.NewUserController(mockUsecase)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
		UserID: uuid.NewString(),
	})

	e.PATCH("/auth/user", func(c echo.Context) error {
		c.Set("user", token)
		return userController.UpdateUser(c)
	}, middleware.BodyLimit("1KiB"))

	req := newUpdateUserRequest(t, nil, &updateUserAvatar{filename: "avatar.png", content: bytes.Repeat([]byte("a"), 8*1024)})
	req.ContentLength = -1
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.JSONEq(t, `{"message":"Request Entity Too Large"}`, rec.Body.String())
}
