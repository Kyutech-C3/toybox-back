package controller

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
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

// ResolveCurrentUser は JWT の Subject(Discord ユーザーID) から登録済みユーザーを引き、claims.UserID に設定するミドルウェア。
// 未登録の場合は UserID を空のまま次へ渡す。
func (uc *UserController) ResolveCurrentUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return next(c)
		}
		claims := token.Claims.(*schema.JWTCustomClaims)
		user, err := uc.userusecase.GetByDiscordUserID(c.Request().Context(), claims.Subject)
		if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
			return handleUserError(err)
		}
		if user != nil {
			claims.UserID = user.ID.String()
		}
		return next(c)
	}
}

// RegisterCurrentUser godoc
// @Summary Register the current authenticated user
// @Description Create the user from the JWT claims if not registered yet. Idempotent.
// @Tags users
// @Produce json
// @Success 200 {object} schema.GetCurrentUserResponse
// @Failure 500 {object} echo.HTTPError
// @Router /auth/users/me [put]
// @Security CookieAuth
func (uc *UserController) RegisterCurrentUser(c echo.Context) error {
	claims := c.Get("user").(*jwt.Token).Claims.(*schema.JWTCustomClaims)
	user, err := uc.userusecase.RegisterUser(c.Request().Context(), entity.DiscordUser{
		ID:         claims.Subject,
		Username:   claims.Username,
		AvatarHash: claims.Avatar,
		Email:      claims.Email,
	})
	if err != nil {
		c.Logger().Error("Failed to register user:", err)
		return handleUserError(err)
	}
	return c.JSON(http.StatusOK, schema.ToCurrentUserResponse(user))
}

// GetCurrentUser godoc
// @Summary Get the current authenticated user
// @Description Get the current authenticated user's ID, display name, and icon URL
// @Tags users
// @Produce json
// @Success 200 {object} schema.GetCurrentUserResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/users/me [get]
// @Security CookieAuth
func (uc *UserController) GetCurrentUser(c echo.Context) error {
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
	return c.JSON(http.StatusOK, schema.ToCurrentUserResponse(user))
}

// UpdateUser godoc
// @Summary Update a user
// @Description Partially update a user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param display_name formData string false "Display name (1-32 characters)"
// @Param profile formData string false "Profile (up to 500 characters)"
// @Param x_username formData string false "X username"
// @Param github_id formData string false "GitHub ID"
// @Param avatar formData file false "Avatar image. Allowed extensions: png, jpg, jpeg, gif, webp. Request body must not exceed 5MiB."
// @Success 200 {object} schema.GetUserOutput
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 413 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/users [patch]
// @Security CookieAuth
func (uc *UserController) UpdateUser(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	}

	var input schema.UpdateUserInput
	if err := c.Bind(&input); err != nil {
		if errors.Is(err, echo.ErrStatusRequestEntityTooLarge) {
			return echo.ErrStatusRequestEntityTooLarge
		}
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	}
	if err := c.Validate(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	}

	avatar, err := c.FormFile("avatar")
	if err != nil {
		switch {
		case errors.Is(err, echo.ErrStatusRequestEntityTooLarge):
			return echo.ErrStatusRequestEntityTooLarge
		case errors.Is(err, http.ErrMissingFile):
			avatar = nil
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
		}
	}

	newUser, err := uc.userusecase.UpdateUser(c.Request().Context(), userID, input.DisplayName, input.Profile, input.XUsername, input.GithubID, avatar)
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
	case errors.Is(err, domainerrors.ErrInvalidFileName):
		return echo.NewHTTPError(http.StatusBadRequest, "ファイル名が無効です")
	case errors.Is(err, domainerrors.ErrUnsupportedFileType):
		return echo.NewHTTPError(http.StatusBadRequest, "対応していないファイル形式です")
	case errors.Is(err, domainerrors.ErrFailedToUploadFile):
		return echo.NewHTTPError(http.StatusInternalServerError, "ファイルのアップロードに失敗しました")
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "サーバーエラーが発生しました")
}
