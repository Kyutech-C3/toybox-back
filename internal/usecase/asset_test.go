package usecase_test

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/usecase"
	"github.com/simesaba80/toybox-back/internal/usecase/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAssetUseCase_UploadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		filename  string
		setup     func(t *testing.T, repo *mock.MockAssetRepository, file *multipart.FileHeader, userID uuid.UUID)
		wantErr   bool
		wantErrIs error
		wantExt   string
	}{
		{
			name:     "正常系: ファイルアップロード成功",
			filename: "test.png",
			setup: func(t *testing.T, repo *mock.MockAssetRepository, file *multipart.FileHeader, userID uuid.UUID) {
				t.Helper()

				assetURL := "https://example.com/assets/" + uuid.NewString() + ".png"
				assetType := "image"
				var capturedAssetID uuid.UUID

				repo.EXPECT().
					UploadFile(gomock.Any(), file, gomock.AssignableToTypeOf(uuid.UUID{}), "png").
					DoAndReturn(func(ctx context.Context, fh *multipart.FileHeader, assetUUID uuid.UUID, extension string) (*string, *string, error) {
						assert.Equal(t, file, fh)
						assert.Equal(t, "png", extension)
						capturedAssetID = assetUUID
						return &assetURL, &assetType, nil
					}).
					Times(1)

				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, asset *entity.Asset) (*entity.Asset, error) {
						assert.Equal(t, capturedAssetID, asset.ID)
						assert.Equal(t, userID, asset.UserID)
						assert.Equal(t, "png", asset.Extension)
						assert.Equal(t, assetURL, asset.URL)
						assert.Equal(t, assetType, asset.AssetType)
						return asset, nil
					}).
					Times(1)
			},
			wantExt: "png",
		},
		{
			name:     "正常系: 複数ドットを含むファイル名は最後の拡張子を使う",
			filename: "archive.photo.jpeg",
			setup: func(t *testing.T, repo *mock.MockAssetRepository, file *multipart.FileHeader, userID uuid.UUID) {
				t.Helper()

				assetURL := "https://example.com/assets/" + uuid.NewString() + ".jpeg"
				assetType := "image"

				repo.EXPECT().
					UploadFile(gomock.Any(), file, gomock.AssignableToTypeOf(uuid.UUID{}), "jpeg").
					Return(&assetURL, &assetType, nil).
					Times(1)

				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, asset *entity.Asset) (*entity.Asset, error) {
						assert.Equal(t, "jpeg", asset.Extension)
						return asset, nil
					}).
					Times(1)
			},
			wantExt: "jpeg",
		},
		{
			name:     "正常系: 大文字の拡張子を小文字に正規化する",
			filename: "example.PNG",
			setup: func(t *testing.T, repo *mock.MockAssetRepository, file *multipart.FileHeader, userID uuid.UUID) {
				t.Helper()

				assetURL := "https://example.com/assets/" + uuid.NewString() + ".png"
				assetType := "image"

				repo.EXPECT().
					UploadFile(gomock.Any(), file, gomock.AssignableToTypeOf(uuid.UUID{}), "png").
					Return(&assetURL, &assetType, nil).
					Times(1)

				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, asset *entity.Asset) (*entity.Asset, error) {
						assert.Equal(t, "png", asset.Extension)
						return asset, nil
					}).
					Times(1)
			},
			wantExt: "png",
		},
		{
			name:     "正常系: webp を許可する",
			filename: "thumb.webp",
			setup: func(t *testing.T, repo *mock.MockAssetRepository, file *multipart.FileHeader, userID uuid.UUID) {
				t.Helper()

				assetURL := "https://example.com/assets/" + uuid.NewString() + ".webp"
				assetType := "image"

				repo.EXPECT().
					UploadFile(gomock.Any(), file, gomock.AssignableToTypeOf(uuid.UUID{}), "webp").
					Return(&assetURL, &assetType, nil).
					Times(1)

				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Asset{Extension: "webp", AssetType: assetType, URL: assetURL}, nil).
					Times(1)
			},
			wantExt: "webp",
		},
		{
			name:      "異常系: 拡張子がない",
			filename:  "noextension",
			wantErr:   true,
			wantErrIs: domainerrors.ErrInvalidFileName,
		},
		{
			name:      "異常系: 未対応形式",
			filename:  "model.gltf",
			wantErr:   true,
			wantErrIs: domainerrors.ErrUnsupportedFileType,
		},
		{
			name:     "異常系: UploadFile でエラー",
			filename: "test.png",
			setup: func(t *testing.T, repo *mock.MockAssetRepository, file *multipart.FileHeader, userID uuid.UUID) {
				t.Helper()

				repo.EXPECT().
					UploadFile(gomock.Any(), file, gomock.AssignableToTypeOf(uuid.UUID{}), "png").
					Return(nil, nil, errors.New("upload failed")).
					Times(1)
			},
			wantErr: true,
		},
		{
			name:     "異常系: Create でエラー",
			filename: "test.png",
			setup: func(t *testing.T, repo *mock.MockAssetRepository, file *multipart.FileHeader, userID uuid.UUID) {
				t.Helper()

				assetURL := "https://example.com/assets/" + uuid.NewString() + ".png"
				assetType := "image"

				repo.EXPECT().
					UploadFile(gomock.Any(), file, gomock.AssignableToTypeOf(uuid.UUID{}), "png").
					Return(&assetURL, &assetType, nil).
					Times(1)

				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("create failed")).
					Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			file := &multipart.FileHeader{Filename: tt.filename}
			userID := uuid.New()

			mockRepo := mock.NewMockAssetRepository(ctrl)
			if tt.setup != nil {
				tt.setup(t, mockRepo, file, userID)
			}

			uc := usecase.NewAssetUseCase(mockRepo)

			got, err := uc.UploadFile(context.Background(), file, userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.wantExt != "" {
				assert.Equal(t, tt.wantExt, got.Extension)
			}
		})
	}
}
