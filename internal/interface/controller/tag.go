package controller

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/internal/usecase"
)

type TagController struct {
	tagUsecase usecase.ITagUseCase
}

func NewTagController(tagUsecase usecase.ITagUseCase) *TagController {
	return &TagController{
		tagUsecase: tagUsecase,
	}
}

// CreateTag godoc
// @Summary Create a new tag
// @Description Create a new tag (authentication required)
// @Tags tags
// @Accept json
// @Produce json
// @Param tag body schema.CreateTagInput true "Tag to create"
// @Success 201 {object} schema.TagResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/tags [post]
func (tc *TagController) CreateTag(c echo.Context) error {
	var input schema.CreateTagInput
	if err := c.Bind(&input); err != nil {
		return handleTagError(c, domainerrors.ErrInvalidRequestBody)
	}
	if err := c.Validate(&input); err != nil {
		return handleTagError(c, domainerrors.ErrInvalidRequestBody)
	}

	createdTag, err := tc.tagUsecase.Create(c.Request().Context(), input.Name)
	if err != nil {
		return handleTagError(c, err)
	}

	return c.JSON(http.StatusCreated, schema.ToTagDetailResponse(createdTag))
}

// GetAllTags godoc
// @Summary Get all tags
// @Description Get all tags
// @Tags tags
// @Produce json
// @Success 200 {object} schema.TagListResponse
// @Failure 500 {object} echo.HTTPError
// @Router /tags [get]
func (tc *TagController) GetAllTags(c echo.Context) error {
	tags, err := tc.tagUsecase.GetAll(c.Request().Context())
	if err != nil {
		return handleTagError(c, err)
	}

	return c.JSON(http.StatusOK, schema.ToTagListResponse(tags))
}

func handleTagError(c echo.Context, err error) error {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	switch {
	case errors.Is(err, domainerrors.ErrInvalidRequestBody):
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストボディです")
	case errors.Is(err, domainerrors.ErrInvalidTagName):
		return echo.NewHTTPError(http.StatusBadRequest, "タグ名が無効です")
	case errors.Is(err, domainerrors.ErrFailedToCreateTag):
		return echo.NewHTTPError(http.StatusInternalServerError, "タグの作成に失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToGetAllTags):
		return echo.NewHTTPError(http.StatusInternalServerError, "タグの取得に失敗しました")
	case errors.Is(err, domainerrors.ErrTagAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, "タグが既に存在します")
	default:
		c.Logger().Error("Tag error:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "サーバーエラーが発生しました")
	}
}

