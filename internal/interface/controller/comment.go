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

type CommentController struct {
	commentUsecase usecase.ICommentUsecase
}

func NewCommentController(commentUsecase usecase.ICommentUsecase) *CommentController {
	return &CommentController{
		commentUsecase: commentUsecase,
	}
}

// GetCommentsByWorkID godoc
// @Summary Get comments for a work
// @Description Get all comments for a specific work
// @Tags comments
// @Produce json
// @Param work_id path string true "Work ID"
// @Success 200 {array} schema.CommentResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /works/{work_id}/comments [get]
// @Security BearerAuth
func (cc *CommentController) GetCommentsByWorkID(c echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid work ID format")
	}

	rawUser := c.Get("user")
	var userID uuid.UUID
	if rawUser == nil {
		userID = uuid.Nil
	} else {
		user := rawUser.(*jwt.Token)
		claims := user.Claims.(*schema.JWTCustomClaims)
		userID, err = uuid.Parse(claims.UserID)
		if err != nil {
			return handleCommentError(c, domainerrors.ErrInvalidRequestBody)
		}
	}

	comments, err := cc.commentUsecase.GetCommentsByWorkID(c.Request().Context(), workID, userID)
	if err != nil {
		c.Logger().Error("CommentUsecase.GetCommentsByWorkID error:", err)
		return handleCommentError(c, err)
	}

	return c.JSON(http.StatusOK, schema.ToCommentListResponse(comments))
}

// CreateComment godoc
// @Summary Create a comment for a work
// @Description Create a new comment for a specific work. Can be anonymous or by a logged-in user.
// @Tags comments
// @Accept json
// @Produce json
// @Param work_id path string true "Work ID"
// @Param comment body schema.CreateCommentRequest true "Comment to create"
// @Success 201 {object} schema.CreateCommentResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /works/{work_id}/comments [post]
// @Security BearerAuth
func (cc *CommentController) CreateComment(c echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid work ID format")
	}

	var input schema.CreateCommentRequest
	if err := c.Bind(&input); err != nil {
		c.Logger().Error("Bind error:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&input); err != nil {
		return err
	}

	rawUser := c.Get("user")
	var userID uuid.UUID
	if rawUser == nil {
		userID = uuid.Nil
	} else {
		user := rawUser.(*jwt.Token)
		claims := user.Claims.(*schema.JWTCustomClaims)
		userID, err = uuid.Parse(claims.UserID)
		if err != nil {
			return handleCommentError(c, domainerrors.ErrInvalidRequestBody)
		}
	}

	createdComment, err := cc.commentUsecase.CreateComment(
		c.Request().Context(),
		input.Content,
		workID,
		userID,
		input.ReplyAt,
	)
	if err != nil {
		c.Logger().Error("CommentUsecase.CreateComment error:", err)
		return handleCommentError(c, err)
	}

	return c.JSON(http.StatusCreated, schema.ToCreateCommentResponse(createdComment))
}

func handleCommentError(c echo.Context, err error) error {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	switch {
	case errors.Is(err, domainerrors.ErrInvalidRequestBody):
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	case errors.Is(err, domainerrors.ErrFailedToGetCommentsByWorkID):
		return echo.NewHTTPError(http.StatusInternalServerError, "コメントの取得に失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToGetCommentById):
		return echo.NewHTTPError(http.StatusInternalServerError, "コメントの取得に失敗しました")
	case errors.Is(err, domainerrors.ErrCommentNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "コメントが見つかりませんでした")
	case errors.Is(err, domainerrors.ErrFailedToCreateComment):
		return echo.NewHTTPError(http.StatusInternalServerError, "コメントの作成に失敗しました")
	case errors.Is(err, domainerrors.ErrWorkNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "作品が見つかりませんでした")
	case errors.Is(err, domainerrors.ErrWorkNotViewable):
		return echo.NewHTTPError(http.StatusForbidden, "この作品にはコメントできません")
	case errors.Is(err, domainerrors.ErrInvalidReplyAt):
		return echo.NewHTTPError(http.StatusBadRequest, "返信先のコメントが不正です")
	}

	c.Logger().Error("Comment error:", err)
	return echo.NewHTTPError(http.StatusInternalServerError, "サーバーエラーが発生しました")
}
