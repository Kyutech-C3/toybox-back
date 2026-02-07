package controller

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/internal/usecase"
)

type UserController struct {
	userusecase usecase.IUserUseCase
}

func NewUserController(userusecase usecase.IUserUseCase) *UserController {
	return &UserController{
		userusecase: userusecase,
	}
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Get all users
// @Tags users
// @Produce  json
// @Success 200 {object} schema.UserListResponse
// @Router /users [get]
func (uc *UserController) GetAllUsers(c echo.Context) error {
	users, err := uc.userusecase.GetAllUser(c.Request().Context())
	if err != nil {
		c.Logger().Error("Failed to get all users:", err)
		return handleUserError(err)
	}

	response := make([]schema.GetUserOutput, len(users))
	for i, user := range users {
		response[i] = schema.ToUserResponse(user)
	}

	return c.JSON(http.StatusOK, schema.UserListResponse{Users: response})
}

// GetUserByID godoc
// @Summary Get a user by ID
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} schema.GetUserOutput
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /users/{id} [get]
func (uc *UserController) GetUserByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	}
	user, err := uc.userusecase.GetByUserID(c.Request().Context(), id)
	if err != nil {
		c.Logger().Error("Failed to get user by ID:", err)
		return handleUserError(err)
	}
	return c.JSON(http.StatusOK, schema.ToUserResponse(user))
}

// GetIconAndURLByUserID godoc
// @Summary Get icon and URL by user ID
// @Description Get icon and URL by user ID
// @Tags users
// @Produce json
// @Success 200 {object} schema.GetIconAndURLResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/users/me [get]
// @Security BearerAuth
func (uc *UserController) GetIconAndURLByUserID(c echo.Context) error {
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	}
	user, err := uc.userusecase.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		return handleUserError(err)
	}
	return c.JSON(http.StatusOK, schema.ToIconAndURLResponse(user))
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param user body schema.UpdateUserInput true "User to update"
// @Success 200 {object} schema.GetUserOutput
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/users [put]
// @Security BearerAuth
func (uc *UserController) UpdateUser(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	}

	var input schema.UpdateUserInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	}

	newUser, err := uc.userusecase.UpdateUser(c.Request().Context(), userID, input.Email, input.DisplayName, input.Profile, input.TwitterID, input.GithubID)
	if err != nil {
		c.Logger().Error("Failed to update user:", err)
		return handleUserError(err)
	}
	return c.JSON(http.StatusOK, schema.ToUserResponse(newUser))
}

func handleUserError(err error) error {
	switch {
	case errors.Is(err, domainerrors.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "ユーザーが見つかりませんでした")
	case errors.Is(err, domainerrors.ErrFailedToUpdateUser):
		return echo.NewHTTPError(http.StatusInternalServerError, "ユーザーの更新に失敗しました")
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "サーバーエラーが発生しました")
}
