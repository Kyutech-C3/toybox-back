//go:build integration

package comment_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/comment"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/testutil"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/work"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testutil.Teardown()
	os.Exit(code)
}

func TestCommentRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	commentRepo := comment.NewCommentRepository(db)

	ctx := context.Background()
	work := insertTestWork(t, db)
	workUUID := work.ID
	comment := &entity.Comment{
		ID:        uuid.New(),
		Content:   "create-content",
		WorkID:    workUUID,
		UserID:    uuid.New(),
		ReplyAt:   "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := commentRepo.Create(ctx, comment)
	require.NoError(t, err)

	anonymousComment := &entity.Comment{
		ID:        uuid.New(),
		Content:   "anonymous-content",
		WorkID:    workUUID,
		UserID:    uuid.Nil,
		ReplyAt:   "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = commentRepo.Create(ctx, anonymousComment)
	require.NoError(t, err)
}

func TestCommentRepository_FindByWorkID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	commentRepo := comment.NewCommentRepository(db)

	ctx := context.Background()
	work := insertTestWork(t, db)
	workUUID := work.ID
	comment := &entity.Comment{
		ID:        uuid.New(),
		Content:   "find-content",
		WorkID:    workUUID,
		UserID:    uuid.New(),
		ReplyAt:   "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := commentRepo.Create(ctx, comment)
	require.NoError(t, err)

	comments, err := commentRepo.FindByWorkID(ctx, workUUID)
	require.NoError(t, err)
	require.Equal(t, 1, len(comments))
	require.Equal(t, comment.Content, comments[0].Content)
	require.Equal(t, comment.WorkID, comments[0].WorkID)
}

func TestCommentRepository_FindByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	commentRepo := comment.NewCommentRepository(db)

	ctx := context.Background()
	work := insertTestWork(t, db)
	workUUID := work.ID
	comment := &entity.Comment{
		ID:        uuid.New(),
		Content:   "find-content",
		WorkID:    workUUID,
		UserID:    uuid.New(),
		ReplyAt:   "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	created, err := commentRepo.Create(ctx, comment)
	require.NoError(t, err)

	found, err := commentRepo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, comment.Content, found.Content)
	require.Equal(t, comment.WorkID, found.WorkID)
}

func insertTestWork(t *testing.T, db *bun.DB) *entity.Work {
	t.Helper()

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test-tag")
	asset := insertTestAsset(t, db, user.ID)
	thumbnailAsset := insertTestAsset(t, db, user.ID)

	workRepo := work.NewWorkRepository(db)
	testWork := &entity.Work{
		ID:               uuid.New(),
		Title:            "create-title",
		Description:      "create-description",
		UserID:           user.ID,
		Visibility:       "public",
		ThumbnailAssetID: thumbnailAsset.ID,
		Assets:           []*entity.Asset{asset},
		TagIDs:           []uuid.UUID{tag.ID},
		Tags:             []*entity.Tag{tag},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	created, err := workRepo.Create(ctx, testWork)
	require.NoError(t, err)
	return created
}

func insertTestUser(t *testing.T, db *bun.DB) *entity.User {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	userID := uuid.New()
	shortID := userID.String()[:8]
	user := &entity.User{
		ID:            userID,
		Name:          "user-" + shortID,
		Email:         "user-" + shortID + "@example.com",
		DisplayName:   "User " + shortID,
		DiscordUserID: "discord-" + shortID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	dtoUser := dto.ToUserDTO(user)
	_, err := db.NewInsert().Model(dtoUser).Exec(context.Background())
	require.NoError(t, err)

	return user
}

func insertTestTag(t *testing.T, db *bun.DB, name string) *entity.Tag {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	tag := &entity.Tag{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dtoTag := dto.ToTagDTO(tag)
	_, err := db.NewInsert().Model(dtoTag).Exec(context.Background())
	require.NoError(t, err)

	return tag
}

func insertTestAsset(t *testing.T, db *bun.DB, userID uuid.UUID) *entity.Asset {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	asset := &entity.Asset{
		ID:        uuid.New(),
		AssetType: "image",
		UserID:    userID,
		Extension: "png",
		URL:       "https://example.com/test.png",
		CreatedAt: now,
		UpdatedAt: now,
	}

	dtoAsset := dto.ToAssetDTO(asset)
	_, err := db.NewInsert().Model(dtoAsset).Exec(context.Background())
	require.NoError(t, err)

	return asset
}
