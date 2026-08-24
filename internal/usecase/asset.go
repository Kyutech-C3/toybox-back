package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/domain/repository"
)

type IAssetUseCase interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, userID uuid.UUID) (*entity.Asset, error)
}

type assetUseCase struct {
	assetRepo repository.AssetRepository
}

var allowedAssetExtensions = map[string]struct{}{
	"png":  {},
	"jpg":  {},
	"jpeg": {},
	"bmp":  {},
	"gif":  {},
	"webp": {},
	"mp4":  {},
	"mov":  {},
	"mp3":  {},
	"wav":  {},
	"m4a":  {},
	"zip":  {},
}

func NewAssetUseCase(assetRepo repository.AssetRepository) IAssetUseCase {
	return &assetUseCase{
		assetRepo: assetRepo,
	}
}

func (uc *assetUseCase) UploadFile(ctx context.Context, file *multipart.FileHeader, userID uuid.UUID) (*entity.Asset, error) {
	extension, err := extractExtension(file.Filename)
	if err != nil {
		return nil, err
	}
	if !isAllowedAssetExtension(extension) {
		return nil, domainerrors.ErrUnsupportedFileType
	}

	asset := entity.NewAsset("", userID, extension, "")

	assetURL, assetType, err := uc.assetRepo.UploadFile(ctx, file, asset.ID, extension)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	asset.URL = *assetURL
	asset.AssetType = *assetType

	createdAsset, err := uc.assetRepo.Create(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}
	return createdAsset, nil
}

func extractExtension(filename string) (string, error) {
	// TrimPrefixは.pngのような形式で返すので.を取り除く必要あり
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	if ext == "" {
		return "", domainerrors.ErrInvalidFileName
	}
	return strings.ToLower(ext), nil
}

func isAllowedAssetExtension(extension string) bool {
	_, ok := allowedAssetExtensions[extension]
	return ok
}
