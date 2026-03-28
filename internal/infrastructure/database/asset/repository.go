package asset

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/types"
	"github.com/uptrace/bun"
)

type mimeType string

const (
	png             mimeType = "image/png"
	jpeg            mimeType = "image/jpeg"
	jpg             mimeType = "image/jpge"
	bmp             mimeType = "image/bmp"
	gif             mimeType = "image/gif"
	webp            mimeType = "image/webp"
	mp4             mimeType = "video/mp4"
	mov             mimeType = "video/quicktime"
	avi             mimeType = "video/x-msvideo"
	flv             mimeType = "video/x-flv"
	mp3             mimeType = "audio/mpeg"
	wav             mimeType = "audio/wav"
	m4a             mimeType = "audio/aac"
	zip             mimeType = "application/zip"
	gilf            mimeType = "model/gltf+json"
	defaultMimeType mimeType = "application/octet-stream"
)

const discordAvatarEndpointFormat = "https://cdn.discordapp.com/avatars/%s/%s.webp?size=256"

var ExtensionToDirName = map[string]string{
	"png":  "image",
	"jpeg": "image",
	"jpg":  "image",
	"bmp":  "image",
	"gif":  "image",
	"webp": "image",
	"mp4":  "video",
	"mov":  "video",
	"avi":  "video",
	"flv":  "video",
	"mp3":  "music",
	"wav":  "music",
	"m4a":  "music",
	"zip":  "zip",
	"gltf": "model",
	"fbx":  "model",
}

type AssetRepository struct {
	db *bun.DB
	s3 *s3.Client
}

func NewAssetRepository(db *bun.DB, s3 *s3.Client) *AssetRepository {
	return &AssetRepository{
		db: db,
		s3: s3,
	}
}

func (r *AssetRepository) Create(ctx context.Context, asset *entity.Asset) (*entity.Asset, error) {
	dtoAsset := dto.ToAssetDTO(asset)

	dtoAsset.AssetType = types.AssetType(ExtensionToDirName[asset.Extension])
	_, err := r.db.NewInsert().Model(dtoAsset).Exec(ctx)
	if err != nil {
		return nil, domainerrors.ErrFailedToCreateAsset
	}
	return dtoAsset.ToAssetEntity(), nil
}

func (r *AssetRepository) UploadFile(ctx context.Context, file *multipart.FileHeader, assetUUID uuid.UUID, extension string) (assetURL *string, assetType *string, err error) {
	openFile, err := file.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer openFile.Close()

	//ファイルのmimeTypeと保存場所を指定
	mimeType := defineMimeType(extension)
	dirName := ExtensionToDirName[extension]
	if dirName == "" {
		dirName = "other"
	}

	_, err = r.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.S3_BUCKET),
		Key:         aws.String(config.S3_DIR + "/" + dirName + "/" + assetUUID.String() + "/origin." + extension),
		Body:        openFile,
		ContentType: aws.String(string(mimeType)),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload file: %w", err)
	}
	newAssetURL := config.S3_BASE_URL + "/" + config.S3_BUCKET + "/" + config.S3_DIR + "/" + dirName + "/" + assetUUID.String() + "/origin." + extension

	return &newAssetURL, &dirName, nil
}

func defineMimeType(extension string) mimeType {
	mimeType := mimeType(extension)
	if mimeType == "" {
		mimeType = defaultMimeType
	}
	return mimeType
}

func (r *AssetRepository) UploadAvatar(ctx context.Context, discordUserID string, avatarHash string) (avatarURL *string, err error) {
	if discordUserID == "" || avatarHash == "" {
		return nil, fmt.Errorf("discord user id or avatar hash is empty")
	}

	discordAvatarURL := fmt.Sprintf(discordAvatarEndpointFormat, discordUserID, avatarHash)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discordAvatarURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord avatar request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download discord avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download discord avatar: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read discord avatar: %w", err)
	}

	s3Key := fmt.Sprintf("%s/avatar/%s.webp", config.S3_DIR, avatarHash)
	_, err = r.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(config.S3_BUCKET),
		Key:           aws.String(s3Key),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(string(webp)),
		ContentLength: aws.Int64(int64(len(data))),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar to s3: %w", err)
	}

	newAvatarURL := fmt.Sprintf("%s/%s/%s", config.S3_BASE_URL, config.S3_BUCKET, s3Key)
	return &newAvatarURL, nil
}

func (r *AssetRepository) DeleteFile(ctx context.Context, url string) error {
	// S3の完全なURLからS3オブジェクトのキーを抽出する。
	// 例: "https://s3.REGION.amazonaws.com/BUCKET_NAME/path/to/object" から "path/to/object" を抽出
	prefix := config.S3_BASE_URL + "/" + config.S3_BUCKET + "/"
	key := ""
	if len(url) > len(prefix) && url[:len(prefix)] == prefix {
		key = url[len(prefix):]
	}

	if key == "" {
		return fmt.Errorf("invalid asset URL: %s", url)
	}

	_, err := r.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(config.S3_BUCKET),
		Key:    aws.String(key),
	})
	if err != nil {
		return domainerrors.ErrFailedToDeleteAsset
	}

	return nil
}
