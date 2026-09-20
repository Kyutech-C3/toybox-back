package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/interface/controller/mock"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFavoriteController_CreateFavorite(t *testing.T) {
	successWorkID := uuid.New()
	successUserID := uuid.New()

	tests := []struct {
		name       string
		userID     string
		workID     string
		setupMock  func(*mock.MockIFavoriteUsecase)
		wantStatus int
		wantBody   string
		wantJSON   bool
	}{
		{
			name:   "正常系: いいね作成が成功する",
			userID: successUserID.String(),
			workID: successWorkID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CreateFavorite(gomock.Any(), successWorkID, successUserID).
					Return(nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   "",
		},
		{
			name:   "異常系: ユーザーIDがUUID形式でない",
			userID: "invalid-uuid",
			workID: successWorkID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CreateFavorite(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid user ID"}`,
			wantJSON:   true,
		},
		{
			name:   "異常系: work_idパラメータがUUID形式でない",
			userID: successUserID.String(),
			workID: "invalid-work-id",
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CreateFavorite(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid work ID format"}`,
			wantJSON:   true,
		},
		{
			name:   "異常系: 既に存在するいいねの場合はエラーを返す",
			userID: successUserID.String(),
			workID: successWorkID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CreateFavorite(gomock.Any(), successWorkID, successUserID).
					Return(domainerrors.ErrFavoriteAlreadyExists)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"既にいいねしています"}`,
			wantJSON:   true,
		},
		{
			name:   "異常系: 作品が存在しない場合はエラーを返す",
			userID: successUserID.String(),
			workID: successWorkID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CreateFavorite(gomock.Any(), successWorkID, successUserID).
					Return(domainerrors.ErrWorkNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"message":"作品が見つかりませんでした"}`,
			wantJSON:   true,
		},
		{
			name:   "異常系: 閲覧できない作品にはいいねできない",
			userID: successUserID.String(),
			workID: successWorkID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CreateFavorite(gomock.Any(), successWorkID, successUserID).
					Return(domainerrors.ErrWorkNotViewable)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   `{"message":"この作品にはいいねできません"}`,
			wantJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mock.NewMockIFavoriteUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			favoriteController := controller.NewFavoriteController(mockUsecase)

			e.POST("/auth/works/:work_id/favorite", func(c echo.Context) error {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
					UserID: tt.userID,
				})
				c.Set("user", token)
				return favoriteController.CreateFavorite(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/auth/works/"+tt.workID+"/favorite", nil)
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

func TestFavoriteController_DeleteFavorite(t *testing.T) {
	workID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		userID     string
		workID     string
		setupMock  func(*mock.MockIFavoriteUsecase)
		wantStatus int
		wantBody   string
		wantJSON   bool
	}{
		{
			name:   "正常系: いいね削除が成功する",
			userID: userID.String(),
			workID: workID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					DeleteFavorite(gomock.Any(), workID, userID).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
		{
			name:   "異常系: ユーザーIDがUUID形式でない",
			userID: "invalid-uuid",
			workID: workID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					DeleteFavorite(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid user ID"}`,
			wantJSON:   true,
		},
		{
			name:   "異常系: work_idパラメータがUUID形式でない",
			userID: userID.String(),
			workID: "invalid-work-id",
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					DeleteFavorite(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid work ID format"}`,
			wantJSON:   true,
		},
		{
			name:   "異常系: 削除対象が存在しない場合はエラーを返す",
			userID: userID.String(),
			workID: workID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					DeleteFavorite(gomock.Any(), workID, userID).
					Return(domainerrors.ErrFavoriteNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"message":"いいねが見つかりませんでした"}`,
			wantJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mock.NewMockIFavoriteUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			favoriteController := controller.NewFavoriteController(mockUsecase)

			e.DELETE("/auth/works/:work_id/favorite", func(c echo.Context) error {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
					UserID: tt.userID,
				})
				c.Set("user", token)
				return favoriteController.DeleteFavorite(c)
			})

			req := httptest.NewRequest(http.MethodDelete, "/auth/works/"+tt.workID+"/favorite", nil)
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

func TestFavoriteController_CountFavoritesByWorkID(t *testing.T) {
	workID := uuid.New()

	successBody, _ := json.Marshal(schema.CountFavoritesByWorkIDResponse{Total: 3})

	tests := []struct {
		name       string
		workID     string
		setupMock  func(*mock.MockIFavoriteUsecase)
		wantStatus int
		wantBody   string
		wantJSON   bool
	}{
		{
			name:   "正常系: 指定作品のいいね数を取得できる",
			workID: workID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CountFavoritesByWorkID(gomock.Any(), workID).
					Return(3, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   string(successBody),
			wantJSON:   true,
		},
		{
			name:   "異常系: work_idパラメータがUUID形式でない",
			workID: "invalid-work-id",
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CountFavoritesByWorkID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid work ID format"}`,
			wantJSON:   true,
		},
		{
			name:   "異常系: いいね数取得時にエラーが発生した場合は500を返す",
			workID: workID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					CountFavoritesByWorkID(gomock.Any(), workID).
					Return(0, domainerrors.ErrFailedToCountFavoritesByWorkID)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"message":"いいねのカウントに失敗しました"}`,
			wantJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mock.NewMockIFavoriteUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			favoriteController := controller.NewFavoriteController(mockUsecase)

			e.GET("/works/:work_id/favorite", func(c echo.Context) error {
				return favoriteController.CountFavoritesByWorkID(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/works/"+tt.workID+"/favorite", nil)
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

func TestFavoriteController_IsFavorite(t *testing.T) {
	workID := uuid.New()
	userID := uuid.New()

	trueResponse, _ := json.Marshal(schema.IsFavoriteResponse{IsFavorite: true})
	falseResponse, _ := json.Marshal(schema.IsFavoriteResponse{IsFavorite: false})

	tests := []struct {
		name       string
		userID     string
		workID     string
		setupMock  func(*mock.MockIFavoriteUsecase)
		wantStatus int
		wantBody   string
		wantJSON   bool
	}{
		{
			name:   "正常系: いいね済みかどうかをtrueで返す",
			userID: userID.String(),
			workID: workID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					IsFavorite(gomock.Any(), workID, userID).
					Return(true)
			},
			wantStatus: http.StatusOK,
			wantBody:   string(trueResponse),
			wantJSON:   true,
		},
		{
			name:   "正常系: いいね済みかどうかをfalseで返す",
			userID: userID.String(),
			workID: workID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					IsFavorite(gomock.Any(), workID, userID).
					Return(false)
			},
			wantStatus: http.StatusOK,
			wantBody:   string(falseResponse),
			wantJSON:   true,
		},
		{
			name:   "異常系: work_idパラメータがUUID形式でない",
			userID: userID.String(),
			workID: "invalid-work-id",
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					IsFavorite(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid work ID format"}`,
			wantJSON:   true,
		},
		{
			name:   "異常系: ユーザーIDがUUID形式でない",
			userID: "invalid-uuid",
			workID: workID.String(),
			setupMock: func(m *mock.MockIFavoriteUsecase) {
				m.EXPECT().
					IsFavorite(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"Invalid user ID"}`,
			wantJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mock.NewMockIFavoriteUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			favoriteController := controller.NewFavoriteController(mockUsecase)

			e.GET("/auth/works/:work_id/favorite/is-favorite", func(c echo.Context) error {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &schema.JWTCustomClaims{
					UserID: tt.userID,
				})
				c.Set("user", token)
				return favoriteController.IsFavorite(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/auth/works/"+tt.workID+"/favorite/is-favorite", nil)
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
