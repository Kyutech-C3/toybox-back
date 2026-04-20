package controller

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/internal/usecase"
)

type WorkController struct {
	workUsecase usecase.IWorkUseCase
}

func NewWorkController(workUsecase usecase.IWorkUseCase) *WorkController {
	return &WorkController{
		workUsecase: workUsecase,
	}
}

// GetAllWorks godoc
// @Summary Get all works
// @Description Get all works with pagination and optional tag filter (OR search)
// @Tags works
// @Produce json
// @Param limit query int false "Limit per page (default: 20, max: 100)"
// @Param page query int false "Page number (default: 1)"
// @Param tag_ids query string false "Comma-separated tag IDs for filtering (OR search)"
// @Success 200 {object} schema.WorkListResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /works [get]
// @Security BearerAuth
func (wc *WorkController) GetAllWorks(c echo.Context) error {
	rawUser := c.Get("user")
	var userID uuid.UUID
	if rawUser == nil {
		userID = uuid.Nil
	} else {
		user := rawUser.(*jwt.Token)
		claims := user.Claims.(*schema.JWTCustomClaims)
		var err error
		userID, err = uuid.Parse(claims.UserID)
		if err != nil {
			return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
		}
	}
	var query schema.GetWorksQuery
	if err := c.Bind(&query); err != nil {
		return handleWorkError(c, err)
	}
	if err := c.Validate(&query); err != nil {
		return err
	}

	// タグIDsをパース
	var tagIDs []uuid.UUID
	if query.TagIDs != "" {
		tagIDStrs := strings.Split(query.TagIDs, ",")
		tagIDs = make([]uuid.UUID, 0, len(tagIDStrs))
		for _, idStr := range tagIDStrs {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			id, err := uuid.Parse(idStr)
			if err != nil {
				return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
			}
			tagIDs = append(tagIDs, id)
		}
	}

	works, total, limit, page, err := wc.workUsecase.GetAll(c.Request().Context(), query.Limit, query.Page, userID, tagIDs)
	if err != nil {
		return handleWorkError(c, err)
	}

	response := make([]schema.GetWorkOutput, len(works))
	for i, work := range works {
		response[i] = schema.ToWorkResponse(work)
	}

	return c.JSON(http.StatusOK, schema.WorkListResponse{
		Works:      response,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	})
}

// GetWorkByID godoc
// @Summary Get a work by ID
// @Description Get a work by ID
// @Tags works
// @Produce json
// @Param work_id path string true "Work ID"
// @Success 200 {object} schema.GetWorkOutput
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /works/{work_id} [get]
func (wc *WorkController) GetWorkByID(c echo.Context) error {
	idStr := c.Param("work_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	}

	work, err := wc.workUsecase.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "Work not found")
		}
		return handleWorkError(c, err)
	}

	return c.JSON(http.StatusOK, schema.ToWorkResponse(work))
}

// GetWorksByUserID godoc
// @Summary Get works by user ID
// @Description Get works by user ID
// @Tags works
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} schema.WorkListResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /works/users/{user_id} [get]
// @Security BearerAuth
func (wc *WorkController) GetWorksByUserID(c echo.Context) error {
	rawUser := c.Get("user")
	var authenticatedUserID uuid.UUID
	if rawUser == nil {
		authenticatedUserID = uuid.Nil
	} else {
		authenticatedUser := rawUser.(*jwt.Token)
		authenticatedClaims := authenticatedUser.Claims.(*schema.JWTCustomClaims)
		var err error
		authenticatedUserID, err = uuid.Parse(authenticatedClaims.UserID)
		if err != nil {
			return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
		}
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}
	works, err := wc.workUsecase.GetByUserID(c.Request().Context(), userID, authenticatedUserID)
	if err != nil {
		return handleWorkError(c, err)
	}
	return c.JSON(http.StatusOK, schema.ToWorkListResponse(works))
}

// CreateWork godoc
// @Summary Create a new work
// @Description Create a new work with the input payload
// @Tags works
// @Accept json
// @Produce json
// @Param work body schema.CreateWorkInput true "Work to create"
// @Success 201 {object} schema.CreateWorkOutput
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/works [post]
func (wc *WorkController) CreateWork(c echo.Context) error {
	var input schema.CreateWorkInput
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}
	if err := c.Bind(&input); err != nil {
		c.Logger().Error("Bind error:", err)
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}
	if err := c.Validate(&input); err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}
	// リクエストボディからuser_idを取得し、UUIDにパース

	createdWork, err := wc.workUsecase.CreateWork(
		c.Request().Context(),
		input.Title,
		input.Description,
		input.Visibility,
		input.ThumbnailAssetID,
		input.AssetIDs,
		input.URLs,
		userID,
		input.TagIDs,
	)
	if err != nil {
		c.Logger().Error("WorkUseCase.CreateWork error:", err)
		return handleWorkError(c, err)
	}

	return c.JSON(http.StatusCreated, schema.ToCreateWorkOutput(createdWork))
}

// UpdateWork godoc
// @Summary Update a work
// @Description Update a work by ID (only owner can update)
// @Tags works
// @Accept json
// @Produce json
// @Param work_id path string true "Work ID"
// @Param work body schema.UpdateWorkInput true "Work to update"
// @Success 200 {object} schema.GetWorkOutput
// @Failure 400 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/works/{work_id} [patch]
func (wc *WorkController) UpdateWork(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}

	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}

	var input schema.UpdateWorkInput
	if err := c.Bind(&input); err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}
	if err := c.Validate(&input); err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}

	updatedWork, err := wc.workUsecase.UpdateWork(c.Request().Context(), workID, userID, input.Title, input.Description, input.Visibility, input.ThumbnailAssetID, input.AssetIDs, input.URLs, input.TagIDs)
	if err != nil {
		return handleWorkError(c, err)
	}

	return c.JSON(http.StatusOK, schema.ToWorkResponse(updatedWork))
}

// DeleteWork godoc
// @Summary Delete a work
// @Description Delete a work by ID (only owner can delete)
// @Tags works
// @Param work_id path string true "Work ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/works/{work_id} [delete]
func (wc *WorkController) DeleteWork(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}
	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		return handleWorkError(c, domainerrors.ErrInvalidRequestBody)
	}

	err = wc.workUsecase.DeleteWork(c.Request().Context(), workID, userID)
	if err != nil {
		return handleWorkError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func handleWorkError(c echo.Context, err error) error {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	switch {
	case errors.Is(err, domainerrors.ErrInvalidRequestBody):
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストボディです")
	case errors.Is(err, domainerrors.ErrFailedToGetWorkById):
		return echo.NewHTTPError(http.StatusNotFound, "作品が見つかりませんでした")
	case errors.Is(err, domainerrors.ErrFailedToGetAllWorksByLimitAndOffset):
		return echo.NewHTTPError(http.StatusInternalServerError, "作品の取得に失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToGetWorkById):
		return echo.NewHTTPError(http.StatusInternalServerError, "作品の取得に失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToGetWorksByUserID):
		return echo.NewHTTPError(http.StatusInternalServerError, "作品の取得に失敗しました")
	case errors.Is(err, domainerrors.ErrWorkNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "作品が見つかりませんでした")
	case errors.Is(err, domainerrors.ErrFailedToBeginTransaction):
		return echo.NewHTTPError(http.StatusInternalServerError, "トランザクションの開始に失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToCommitTransaction):
		return echo.NewHTTPError(http.StatusInternalServerError, "トランザクションのコミットに失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToRollbackTransaction):
		return echo.NewHTTPError(http.StatusInternalServerError, "トランザクションのロールバックに失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToCreateWork):
		return echo.NewHTTPError(http.StatusBadRequest, "作品の作成に失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToUpdateWork):
		return echo.NewHTTPError(http.StatusBadRequest, "作品の更新に失敗しました")
	case errors.Is(err, domainerrors.ErrTagNotFound):
		return echo.NewHTTPError(http.StatusBadRequest, "存在しないタグIDが含まれています")
	case errors.Is(err, domainerrors.ErrInvalidTagIDs):
		return echo.NewHTTPError(http.StatusBadRequest, "タグが指定されていません")
	case errors.Is(err, domainerrors.ErrWorkNotOwnedByUser):
		return echo.NewHTTPError(http.StatusForbidden, "この作品を削除する権限がありません")
	case errors.Is(err, domainerrors.ErrFailedToDeleteWork):
		return echo.NewHTTPError(http.StatusInternalServerError, "作品の削除に失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToDeleteAsset):
		return echo.NewHTTPError(http.StatusInternalServerError, "アセットの削除に失敗しました")
	default:
		c.Logger().Error("Work error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "サーバーエラーが発生しました")
	}
}
