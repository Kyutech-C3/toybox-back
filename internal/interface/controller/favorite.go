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

type FavoriteController struct {
	favoriteUsecase usecase.IFavoriteUsecase
}

func NewFavoriteController(favoriteUsecase usecase.IFavoriteUsecase) *FavoriteController {
	return &FavoriteController{favoriteUsecase: favoriteUsecase}
}

// CreateFavorite godoc
// @Summary Create a favorite
// @Description Create a favorite
// @Tags favorites
// @Accept json
// @Produce json
// @Param work_id path string true "Work ID"
// @Success 201
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/works/:work_id/favorite [post]
func (fc *FavoriteController) CreateFavorite(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.Logger().Error("Invalid user ID:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		c.Logger().Error("Invalid work ID format:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid work ID format")
	}
	err = fc.favoriteUsecase.CreateFavorite(c.Request().Context(), workID, userID)
	if err != nil {
		c.Logger().Error("Failed to create favorite:", err)
		return handleFavoriteError(err)
	}
	return c.NoContent(http.StatusCreated)
}

// DeleteFavorite godoc
// @Summary Delete a favorite
// @Description Delete a favorite
// @Tags favorites
// @Accept json
// @Produce json
// @Param work_id path string true "Work ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/works/:work_id/favorite [delete]
func (fc *FavoriteController) DeleteFavorite(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.Logger().Error("Invalid user ID:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		c.Logger().Error("Invalid work ID format:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid work ID format")
	}
	err = fc.favoriteUsecase.DeleteFavorite(c.Request().Context(), workID, userID)
	if err != nil {
		c.Logger().Error("Failed to delete favorite:", err)
		return handleFavoriteError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// CountFavoritesByWorkID godoc
// @Summary Count favorites by work ID
// @Description Count favorites by work ID
// @Tags favorites
// @Accept json
// @Produce json
// @Param work_id path string true "Work ID"
// @Success 200 {object} schema.CountFavoritesByWorkIDResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /works/:work_id/favorite [get]
func (fc *FavoriteController) CountFavoritesByWorkID(c echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		c.Logger().Error("Invalid work ID format:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid work ID format")
	}
	total, err := fc.favoriteUsecase.CountFavoritesByWorkID(c.Request().Context(), workID)
	if err != nil {
		c.Logger().Error("Failed to count favorites by work ID:", err)
		return handleFavoriteError(err)
	}
	return c.JSON(http.StatusOK, schema.CountFavoritesByWorkIDResponse{Total: total})
}

// IsFavorite godoc
// @Summary Check if a user has favorited a work
// @Description Check if a user has favorited a work
// @Tags favorites
// @Accept json
// @Produce json
// @Param work_id path string true "Work ID"
// @Success 200 {object} schema.IsFavoriteResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/works/:work_id/favorite/is-favorite [get]
func (fc *FavoriteController) IsFavorite(c echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		c.Logger().Error("Invalid work ID format:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid work ID format")
	}
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.Logger().Error("Invalid user ID:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}
	isFavorite := fc.favoriteUsecase.IsFavorite(c.Request().Context(), workID, userID)
	return c.JSON(http.StatusOK, schema.IsFavoriteResponse{IsFavorite: isFavorite})
}

func handleFavoriteError(err error) error {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	switch {
	case errors.Is(err, domainerrors.ErrInvalidRequestBody):
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	case errors.Is(err, domainerrors.ErrFailedToCreateFavorite):
		return echo.NewHTTPError(http.StatusInternalServerError, "いいねに失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToDeleteFavorite):
		return echo.NewHTTPError(http.StatusInternalServerError, "いいねの削除に失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToCountFavoritesByWorkID):
		return echo.NewHTTPError(http.StatusInternalServerError, "いいねのカウントに失敗しました")
	case errors.Is(err, domainerrors.ErrFavoriteAlreadyExists):
		return echo.NewHTTPError(http.StatusBadRequest, "既にいいねしています")
	case errors.Is(err, domainerrors.ErrFavoriteNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "いいねが見つかりませんでした")
	case errors.Is(err, domainerrors.ErrWorkNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "作品が見つかりませんでした")
	case errors.Is(err, domainerrors.ErrWorkNotViewable):
		return echo.NewHTTPError(http.StatusForbidden, "この作品にはいいねできません")
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "サーバーエラーが発生しました")
}
