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

type AssetController struct {
	assetUsecase usecase.IAssetUseCase
}

func NewAssetController(assetUsecase usecase.IAssetUseCase) *AssetController {
	return &AssetController{
		assetUsecase: assetUsecase,
	}
}

// UploadAsset godoc
// @Summary Upload an asset
// @Description Upload an asset. Allowed extensions: png, jpg, jpeg, bmp, gif, webp, mp4, mov, mp3, wav, m4a, zip. Maximum request body size is 2GiB.
// @Tags assets
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload. Allowed extensions: png, jpg, jpeg, bmp, gif, webp, mp4, mov, mp3, wav, m4a, zip. Request body must not exceed 2GiB."
// @Success 200 {object} schema.UploadAssetResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 413 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /auth/works/asset [post]
func (ac *AssetController) UploadAsset(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*schema.JWTCustomClaims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return handleAssetError(c, domainerrors.ErrInvalidRequestBody)
	}
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "File is required")
	}
	asset, err := ac.assetUsecase.UploadFile(c.Request().Context(), file, userID)
	if err != nil {
		c.Logger().Error("Failed to upload asset: %w", err)
		return handleAssetError(c, err)
	}
	return c.JSON(http.StatusOK, schema.ToUploadAssetResponse(asset))
}

func handleAssetError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domainerrors.ErrInvalidRequestBody):
		return echo.NewHTTPError(http.StatusBadRequest, "無効なリクエストです")
	case errors.Is(err, domainerrors.ErrInvalidFileName):
		return echo.NewHTTPError(http.StatusBadRequest, "ファイル名が無効です")
	case errors.Is(err, domainerrors.ErrUnsupportedFileType):
		return echo.NewHTTPError(http.StatusBadRequest, "対応していないファイル形式です")
	case errors.Is(err, domainerrors.ErrFailedToOpenFile):
		return echo.NewHTTPError(http.StatusInternalServerError, "ファイルの読み込みに失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToUploadFile):
		return echo.NewHTTPError(http.StatusInternalServerError, "ファイルのアップロードに失敗しました")
	case errors.Is(err, domainerrors.ErrFailedToCreateAsset):
		return echo.NewHTTPError(http.StatusInternalServerError, "アセットの作成に失敗しました")
	}
	c.Logger().Error("Failed to upload asset: %w", err)
	return echo.NewHTTPError(http.StatusInternalServerError, "サーバーエラーが発生しました")
}
