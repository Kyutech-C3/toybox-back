//go:build integration

package user_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/testutil"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/user"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testutil.Teardown()
	os.Exit(code)
}

func TestUserRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := user.NewUserRepository(db)

	ctx := context.Background()
	user := &entity.User{
		ID:            uuid.New(),
		Name:          "testuser",
		Email:         "testuser@example.com",
		DisplayName:   "testuser",
		AvatarURL:     "https://example.com/avatar.png",
		DiscordUserID: "testuser",
	}
	_, err := repo.Create(ctx, user)
	require.NoError(t, err)
}

func TestUserRepository_GetAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := user.NewUserRepository(db)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		user := &entity.User{
			ID:            uuid.New(),
			Name:          "testuser" + strconv.Itoa(i),
			Email:         "testuser" + strconv.Itoa(i) + "@example.com",
			DisplayName:   "testuser",
			AvatarURL:     "https://example.com/avatar.png",
			DiscordUserID: "testuser",
		}
		_, err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	_, err := repo.GetAll(ctx)
	require.NoError(t, err)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := user.NewUserRepository(db)

	ctx := context.Background()
	user := &entity.User{
		ID:            uuid.New(),
		Name:          "testuser",
		Email:         "testuser@example.com",
		DisplayName:   "testuser",
		AvatarURL:     "https://example.com/avatar.png",
		DiscordUserID: "testuser",
	}
	created, err := repo.Create(ctx, user)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.Equal(t, created.Name, found.Name)
	require.Equal(t, created.Email, found.Email)
	require.Equal(t, created.DisplayName, found.DisplayName)
	require.Equal(t, created.AvatarURL, found.AvatarURL)
	require.Equal(t, created.DiscordUserID, found.DiscordUserID)
}

func TestUserRepository_GetUserByDiscordUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := user.NewUserRepository(db)

	ctx := context.Background()
	user := &entity.User{
		ID:            uuid.New(),
		Name:          "testuser",
		Email:         "testuser@example.com",
		DisplayName:   "testuser",
		AvatarURL:     "https://example.com/avatar.png",
		DiscordUserID: "testuser",
	}
	created, err := repo.Create(ctx, user)
	require.NoError(t, err)

	found, err := repo.GetUserByDiscordUserID(ctx, created.DiscordUserID)
	require.NoError(t, err)
	require.Equal(t, created.ID, found.ID)
	require.Equal(t, created.Name, found.Name)
	require.Equal(t, created.Email, found.Email)
	require.Equal(t, created.DisplayName, found.DisplayName)
	require.Equal(t, created.AvatarURL, found.AvatarURL)
	require.Equal(t, created.DiscordUserID, found.DiscordUserID)

	found, err = repo.GetUserByDiscordUserID(ctx, "notfound")
	require.ErrorIs(t, err, domainerrors.ErrUserNotFound)
	require.Nil(t, found)
}

func TestUserRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := user.NewUserRepository(db)

	ctx := context.Background()
	testUser := &entity.User{
		ID:            uuid.New(),
		Name:          "testuser",
		Email:         "testuser@example.com",
		DisplayName:   "testuser",
		AvatarURL:     "https://example.com/avatar.png",
		DiscordUserID: "testuser",
		Profile:       "",
		XUsername:     "",
		GithubID:      "",
	}
	created, err := repo.Create(ctx, testUser)
	require.NoError(t, err)

	// 更新するデータを設定
	created.Email = "updated@example.com"
	created.DisplayName = "Updated User"
	created.Profile = "Updated profile"
	created.XUsername = "xuser123"
	created.GithubID = "github123"

	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, created.ID, updated.ID)
	require.Equal(t, "updated@example.com", updated.Email)
	require.Equal(t, "Updated User", updated.DisplayName)
	require.Equal(t, "Updated profile", updated.Profile)
	require.Equal(t, "xuser123", updated.XUsername)
	require.Equal(t, "github123", updated.GithubID)

	// DBから再取得して確認
	found, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "updated@example.com", found.Email)
	require.Equal(t, "Updated User", found.DisplayName)
	require.Equal(t, "Updated profile", found.Profile)
	require.Equal(t, "xuser123", found.XUsername)
	require.Equal(t, "github123", found.GithubID)
}
