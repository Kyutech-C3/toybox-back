//go:build integration

package favorite_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/favorite"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/testutil"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testutil.Teardown()
	os.Exit(code)
}

func TestFavoriteRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := favorite.NewFavoriteRepository(db)

	ctx := context.Background()

	user := insertTestUser(t, db)
	work := insertTestWork(t, db, user.ID)
	fav := entity.NewFavorite(work.ID, user.ID)

	created, err := repo.Create(ctx, fav)
	require.NoError(t, err)
	require.Equal(t, work.ID, created.WorkID)
	require.Equal(t, user.ID, created.UserID)
	require.WithinDuration(t, fav.CreatedAt, created.CreatedAt, time.Second)

	exists := repo.Exists(ctx, fav)
	require.True(t, exists)
}

func TestFavoriteRepository_Create_Duplicate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := favorite.NewFavoriteRepository(db)

	ctx := context.Background()

	user := insertTestUser(t, db)
	work := insertTestWork(t, db, user.ID)
	fav := entity.NewFavorite(work.ID, user.ID)

	_, err := repo.Create(ctx, fav)
	require.NoError(t, err)

	_, err = repo.Create(ctx, fav)
	require.ErrorIs(t, err, domainerrors.ErrFailedToCreateFavorite)
}

func TestFavoriteRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := favorite.NewFavoriteRepository(db)

	ctx := context.Background()

	user := insertTestUser(t, db)
	work := insertTestWork(t, db, user.ID)
	fav := entity.NewFavorite(work.ID, user.ID)

	_, err := repo.Create(ctx, fav)
	require.NoError(t, err)

	err = repo.Delete(ctx, fav)
	require.NoError(t, err)

	require.False(t, repo.Exists(ctx, fav))
}

func TestFavoriteRepository_CountByWorkID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := favorite.NewFavoriteRepository(db)

	ctx := context.Background()

	userA := insertTestUser(t, db)
	userB := insertTestUser(t, db)
	work := insertTestWork(t, db, userA.ID)

	favA := entity.NewFavorite(work.ID, userA.ID)
	favB := entity.NewFavorite(work.ID, userB.ID)

	_, err := repo.Create(ctx, favA)
	require.NoError(t, err)
	_, err = repo.Create(ctx, favB)
	require.NoError(t, err)

	total, err := repo.CountByWorkID(ctx, work.ID)
	require.NoError(t, err)
	require.Equal(t, 2, total)
}

func TestFavoriteRepository_Exists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := favorite.NewFavoriteRepository(db)

	ctx := context.Background()

	user := insertTestUser(t, db)
	otherUser := insertTestUser(t, db)
	work := insertTestWork(t, db, user.ID)

	fav := entity.NewFavorite(work.ID, user.ID)
	_, err := repo.Create(ctx, fav)
	require.NoError(t, err)

	require.True(t, repo.Exists(ctx, fav))

	otherFav := entity.NewFavorite(work.ID, otherUser.ID)
	require.False(t, repo.Exists(ctx, otherFav))
}

func TestFavoriteRepository_FindFavoritedWorkIDs(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := favorite.NewFavoriteRepository(db)

	ctx := context.Background()

	user := insertTestUser(t, db)
	otherUser := insertTestUser(t, db)
	favoritedWork := insertTestWork(t, db, user.ID)
	notFavoritedWork := insertTestWork(t, db, user.ID)
	otherUsersFavoritedWork := insertTestWork(t, db, user.ID)

	_, err := repo.Create(ctx, entity.NewFavorite(favoritedWork.ID, user.ID))
	require.NoError(t, err)
	_, err = repo.Create(ctx, entity.NewFavorite(otherUsersFavoritedWork.ID, otherUser.ID))
	require.NoError(t, err)

	got, err := repo.FindFavoritedWorkIDs(ctx, user.ID, []uuid.UUID{favoritedWork.ID, notFavoritedWork.ID, otherUsersFavoritedWork.ID})
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{favoritedWork.ID}, got)
}

func TestFavoriteRepository_FindFavoritedWorkIDs_EmptyWorkIDs(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := favorite.NewFavoriteRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)

	got, err := repo.FindFavoritedWorkIDs(ctx, user.ID, []uuid.UUID{})
	require.NoError(t, err)
	require.Empty(t, got)
}

func insertTestUser(t *testing.T, db *bun.DB) *entity.User {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	shortID := uuid.New().String()[:8]
	user := &entity.User{
		ID:            uuid.New(),
		Name:          fmt.Sprintf("user-%s", shortID),
		Email:         fmt.Sprintf("test-%s@example.com", uuid.New().String()),
		DisplayName:   fmt.Sprintf("tester-%s", shortID),
		DiscordUserID: fmt.Sprintf("discord-%s", shortID),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	dtoUser := dto.ToUserDTO(user)
	_, err := db.NewInsert().Model(dtoUser).Exec(context.Background())
	require.NoError(t, err)

	return user
}

func insertTestWork(t *testing.T, db *bun.DB, userID uuid.UUID) *entity.Work {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	shortID := uuid.New().String()[:8]
	work := &entity.Work{
		ID:          uuid.New(),
		Title:       fmt.Sprintf("test-work-%s", shortID),
		Description: "description",
		UserID:      userID,
		Visibility:  "public",
		Assets:      []*entity.Asset{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	dtoWork := dto.ToWorkDTO(work)
	_, err := db.NewInsert().Model(dtoWork).Exec(context.Background())
	require.NoError(t, err)

	return work
}
