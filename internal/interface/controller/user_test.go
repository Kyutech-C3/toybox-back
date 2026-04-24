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

func TestUserController_GetIconAndURLByUserID(t *testing.T) {
	userID := uuid.New()
	mockUser := &entity.User{ID: userID, Name: "testuser"}
	successResponseBytes, _ := json.Marshal(schema.ToIconAndURLResponse(mockUser))
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
				return userController.GetIconAndURLByUserID(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/auth/users/me", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
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
		TwitterID:   "twitter123",
		GithubID:    "github123",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	input := schema.UpdateUserInput{
		Email:       "updated@example.com",
		DisplayName: "Updated User",
		Profile:     "Updated profile",
		TwitterID:   "twitter123",
		GithubID:    "github123",
	}
	inputJSON, _ := json.Marshal(input)

	successResponseBytes, _ := json.Marshal(schema.ToUserResponse(mockUser))
	badRequestResponseBytes, _ := json.Marshal(map[string]string{"message": "無効なリクエストです"})
	notFoundResponseBytes, _ := json.Marshal(map[string]string{"message": "ユーザーが見つかりませんでした"})
	failedToUpdateResponseBytes, _ := json.Marshal(map[string]string{"message": "ユーザーの更新に失敗しました"})
	internalErrorResponseBytes, _ := json.Marshal(map[string]string{"message": "サーバーエラーが発生しました"})

	tests := []struct {
		name       string
		body       []byte
		setupMock  func(mockUserUsecase *mock.MockIUserUseCase)
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "正常系: ユーザー更新成功",
			body: inputJSON,
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, input.Email, input.DisplayName, input.Profile, input.TwitterID, input.GithubID).
					Return(mockUser, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   successResponseBytes,
		},
		{
			name:       "異常系: 不正なリクエストボディ",
			body:       []byte("invalid json"),
			setupMock:  func(mockUserUsecase *mock.MockIUserUseCase) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   badRequestResponseBytes,
		},
		{
			name: "異常系: ユーザーが見つからない",
			body: inputJSON,
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, input.Email, input.DisplayName, input.Profile, input.TwitterID, input.GithubID).
					Return(nil, domainerrors.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   notFoundResponseBytes,
		},
		{
			name: "異常系: ユーザー更新失敗",
			body: inputJSON,
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, input.Email, input.DisplayName, input.Profile, input.TwitterID, input.GithubID).
					Return(nil, domainerrors.ErrFailedToUpdateUser)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   failedToUpdateResponseBytes,
		},
		{
			name: "異常系: 予期しないエラー",
			body: inputJSON,
			setupMock: func(mockUserUsecase *mock.MockIUserUseCase) {
				mockUserUsecase.EXPECT().
					UpdateUser(gomock.Any(), userID, input.Email, input.DisplayName, input.Profile, input.TwitterID, input.GithubID).
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
			e.PUT("/auth/user", func(c echo.Context) error {
				c.Set("user", token)
				return userController.UpdateUser(c)
			})

			req := httptest.NewRequest(http.MethodPut, "/auth/user", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, string(tt.wantBody), rec.Body.String())
		})
	}
}
