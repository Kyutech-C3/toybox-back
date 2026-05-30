//go:build integration

package asset_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/asset"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/testutil"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testutil.Teardown()
	testutil.TeardownS3()
	os.Exit(code)
}

func TestAssetRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	s3Client := testutil.SetupTestS3(t)
	repo := asset.NewAssetRepository(db, s3Client)

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	testAsset := &entity.Asset{
		ID:        uuid.New(),
		WorkID:    uuid.Nil,
		UserID:    userID,
		Extension: "png",
		URL:       "https://example.com/assets/sample.png",
		CreatedAt: now,
		UpdatedAt: now,
	}

	created, err := repo.Create(ctx, testAsset)
	require.NoError(t, err)
	require.Equal(t, "image", created.AssetType)
	require.Equal(t, testAsset.URL, created.URL)
	require.Equal(t, testAsset.Extension, created.Extension)
	require.Equal(t, testAsset.UserID, created.UserID)

	var stored dto.Asset
	err = db.NewSelect().Model(&stored).Where("id = ?", testAsset.ID).Scan(ctx)
	require.NoError(t, err)
	require.Equal(t, testAsset.URL, stored.URL)
	require.Equal(t, testAsset.Extension, stored.Extension)
	require.Equal(t, testAsset.UserID, stored.UserID)
	require.Equal(t, "image", string(stored.AssetType))
}

func TestAssetRepository_UploadFile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	s3Client := testutil.SetupTestS3(t)
	repo := asset.NewAssetRepository(db, s3Client)

	ctx := context.Background()
	fileHeader := newTestFileHeader(t, "test.png", []byte("dummy data"))
	assetID := uuid.New()

	assetURL, assetType, err := repo.UploadFile(ctx, fileHeader, assetID, "png")
	require.NoError(t, err)
	require.NotNil(t, assetURL)
	require.NotNil(t, assetType)
	require.Equal(t, "image", *assetType)

	expectedKey := config.S3_DIR + "/image/" + assetID.String() + "/origin.png"
	expectedURL := config.S3_BASE_URL + "/" + config.S3_BUCKET + "/" + expectedKey
	require.Equal(t, expectedURL, *assetURL)

	resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.S3_BUCKET),
		Key:    aws.String(expectedKey),
	})
	require.NoError(t, err)
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, []byte("dummy data"), content)
}

func TestAssetRepository_DeleteFile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	s3Client := testutil.SetupTestS3(t)
	repo := asset.NewAssetRepository(db, s3Client)

	ctx := context.Background()
	fileHeader := newTestFileHeader(t, "test-delete.png", []byte("to be deleted"))
	assetID := uuid.New()

	assetURL, _, err := repo.UploadFile(ctx, fileHeader, assetID, "png")
	require.NoError(t, err)
	require.NotNil(t, assetURL)

	expectedKey := config.S3_DIR + "/image/" + assetID.String() + "/origin.png"
	_, err = s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.S3_BUCKET),
		Key:    aws.String(expectedKey),
	})
	require.NoError(t, err)

	err = repo.DeleteFile(ctx, *assetURL)
	require.NoError(t, err)

	_, err = s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.S3_BUCKET),
		Key:    aws.String(expectedKey),
	})
	require.Error(t, err)
}

func TestAssetRepository_DeleteFile_InvalidURL(t *testing.T) {
	db := testutil.SetupTestDB(t)
	s3Client := testutil.SetupTestS3(t)
	repo := asset.NewAssetRepository(db, s3Client)

	ctx := context.Background()

	err := repo.DeleteFile(ctx, "invalid-url")
	require.Error(t, err)
}

func newTestFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(int64(len(content))+1024))

	_, fileHeader, err := req.FormFile("file")
	require.NoError(t, err)

	return fileHeader
}
