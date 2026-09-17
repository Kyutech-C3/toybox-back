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
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /works/{work_id}/comments [get]
func (cc *CommentController) GetCommentsByWorkID(c echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := uuid.Parse(workIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid work ID format")
	}

	comments, err := cc.commentUsecase.GetCommentsByWorkID(c.Request().Context(), workID)
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
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /works/{work_id}/comments [post]
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

	var userID uuid.UUID
	if input.UserID != "" {
		userID, err = uuid.Parse(input.UserID)
		if err != nil {
			c.Logger().Error("Invalid UserID format:", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid UserID format")
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

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment. Only the author of the comment can delete it.
// @Tags comments
// @Param comment_id path string true "Comment ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/comments/{comment_id} [delete]
func (cc *CommentController) DeleteComment(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.Logger().Error("Invalid user ID:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	commentIDStr := c.Param("comment_id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID format")
	}

	if err := cc.commentUsecase.DeleteComment(c.Request().Context(), commentID, userID); err != nil {
		c.Logger().Error("CommentUsecase.DeleteComment error:", err)
		return handleCommentError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
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
	case errors.Is(err, domainerrors.ErrCommentNotOwnedByUser):
		return echo.NewHTTPError(http.StatusForbidden, "このコメントを削除する権限がありません")
	case errors.Is(err, domainerrors.ErrFailedToDeleteComment):
		return echo.NewHTTPError(http.StatusInternalServerError, "コメントの削除に失敗しました")
	}

	c.Logger().Error("Comment error:", err)
	return echo.NewHTTPError(http.StatusInternalServerError, "サーバーエラーが発生しました")
}
