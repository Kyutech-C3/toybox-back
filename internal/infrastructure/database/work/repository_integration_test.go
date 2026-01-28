//go:build integration

package work_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"github.com/simesaba80/toybox-back/internal/domain/entity"
	domainerrors "github.com/simesaba80/toybox-back/internal/domain/errors"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/testutil"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/work"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testutil.Teardown()
	os.Exit(code)
}

func TestWorkRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag1 := insertTestTag(t, db, "go")
	tag2 := insertTestTag(t, db, "rust")

	work := newTestWork(user.ID, "create-title")
	work.TagIDs = []uuid.UUID{tag1.ID, tag2.ID}
	work.Tags = []*entity.Tag{tag1, tag2}

	created, err := repo.Create(ctx, work)
	require.NoError(t, err)
	require.Equal(t, work.Title, created.Title)
	require.Equal(t, work.Description, created.Description)
	require.Equal(t, work.UserID, created.UserID)
	require.Equal(t, work.URLs, created.URLs)
	require.Equal(t, 2, len(created.Tags))
	require.Equal(t, 2, len(created.TagIDs))

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, fetched.ID)
	require.Equal(t, 2, len(fetched.Tags))
}

func TestWorkRepository_GetAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test")

	for i := 0; i < 3; i++ {
		asset := insertTestAsset(t, db, user.ID)
		work := newTestWork(user.ID, "title-"+uuid.NewString())
		work.Assets = []*entity.Asset{asset}
		work.TagIDs = []uuid.UUID{tag.ID}
		work.Tags = []*entity.Tag{tag}
		work.CreatedAt = work.CreatedAt.Add(time.Duration(i) * time.Minute)
		work.UpdatedAt = work.CreatedAt
		_, err := repo.Create(ctx, work)
		require.NoError(t, err)
	}

	works, total, err := repo.GetAll(ctx, 10, 0, nil)
	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, works, 3)
	require.True(t, works[0].CreatedAt.After(works[1].CreatedAt) || works[0].CreatedAt.Equal(works[1].CreatedAt))

	for _, w := range works {
		require.Equal(t, 1, len(w.Tags))
		require.Equal(t, tag.Name, w.Tags[0].Name)
	}
}

func TestWorkRepository_GetAllPublic(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test-tag")

	// 公開作品を2つ作成
	for i := 0; i < 2; i++ {
		asset := insertTestAsset(t, db, user.ID)
		work := newTestWork(user.ID, "public-title-"+uuid.NewString())
		work.Visibility = "public"
		work.Assets = []*entity.Asset{asset}
		work.TagIDs = []uuid.UUID{tag.ID}
		work.Tags = []*entity.Tag{tag}
		work.CreatedAt = work.CreatedAt.Add(time.Duration(i) * time.Minute)
		work.UpdatedAt = work.CreatedAt
		_, err := repo.Create(ctx, work)
		require.NoError(t, err)
	}

	// 非公開作品を1つ作成
	asset := insertTestAsset(t, db, user.ID)
	privateWork := newTestWork(user.ID, "private-title-"+uuid.NewString())
	privateWork.Visibility = "private"
	privateWork.Assets = []*entity.Asset{asset}
	privateWork.TagIDs = []uuid.UUID{tag.ID}
	privateWork.Tags = []*entity.Tag{tag}
	_, err := repo.Create(ctx, privateWork)
	require.NoError(t, err)

	// 下書き作品を1つ作成
	asset = insertTestAsset(t, db, user.ID)
	draftWork := newTestWork(user.ID, "draft-title-"+uuid.NewString())
	draftWork.Visibility = "draft"
	draftWork.Assets = []*entity.Asset{asset}
	draftWork.TagIDs = []uuid.UUID{tag.ID}
	draftWork.Tags = []*entity.Tag{tag}
	_, err = repo.Create(ctx, draftWork)
	require.NoError(t, err)

	// GetAllPublicは公開作品のみを取得する
	works, total, err := repo.GetAllPublic(ctx, 10, 0, nil)
	require.NoError(t, err)
	require.Equal(t, 2, total, "公開作品のみカウントされる")
	require.Len(t, works, 2, "公開作品のみ取得される")

	// 全ての取得した作品が公開であることを確認
	for _, work := range works {
		require.Equal(t, "public", work.Visibility, "取得した作品は全て公開である")
	}

	// 作成日時の降順でソートされていることを確認
	require.True(t, works[0].CreatedAt.After(works[1].CreatedAt) || works[0].CreatedAt.Equal(works[1].CreatedAt))
}

func TestWorkRepository_GetAllPublic_WithTagFilter(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag1 := insertTestTag(t, db, "go")
	tag2 := insertTestTag(t, db, "rust")
	tag3 := insertTestTag(t, db, "python")

	// tag1のみを持つ作品を2つ作成
	for i := 0; i < 2; i++ {
		asset := insertTestAsset(t, db, user.ID)
		w := newTestWork(user.ID, "go-work-"+uuid.NewString())
		w.Visibility = "public"
		w.Assets = []*entity.Asset{asset}
		w.TagIDs = []uuid.UUID{tag1.ID}
		w.Tags = []*entity.Tag{tag1}
		_, err := repo.Create(ctx, w)
		require.NoError(t, err)
	}

	// tag2のみを持つ作品を1つ作成
	asset := insertTestAsset(t, db, user.ID)
	rustWork := newTestWork(user.ID, "rust-work")
	rustWork.Visibility = "public"
	rustWork.Assets = []*entity.Asset{asset}
	rustWork.TagIDs = []uuid.UUID{tag2.ID}
	rustWork.Tags = []*entity.Tag{tag2}
	_, err := repo.Create(ctx, rustWork)
	require.NoError(t, err)

	// tag3のみを持つ作品を1つ作成
	asset = insertTestAsset(t, db, user.ID)
	pythonWork := newTestWork(user.ID, "python-work")
	pythonWork.Visibility = "public"
	pythonWork.Assets = []*entity.Asset{asset}
	pythonWork.TagIDs = []uuid.UUID{tag3.ID}
	pythonWork.Tags = []*entity.Tag{tag3}
	_, err = repo.Create(ctx, pythonWork)
	require.NoError(t, err)

	// 単一タグでフィルタリング（tag1のみ）
	works, total, err := repo.GetAllPublic(ctx, 10, 0, []uuid.UUID{tag1.ID})
	require.NoError(t, err)
	require.Equal(t, 2, total, "tag1を持つ作品は2件")
	require.Len(t, works, 2)

	// OR検索: tag1またはtag2を持つ作品
	works, total, err = repo.GetAllPublic(ctx, 10, 0, []uuid.UUID{tag1.ID, tag2.ID})
	require.NoError(t, err)
	require.Equal(t, 3, total, "tag1またはtag2を持つ作品は3件")
	require.Len(t, works, 3)

	// OR検索: 全タグを指定
	works, total, err = repo.GetAllPublic(ctx, 10, 0, []uuid.UUID{tag1.ID, tag2.ID, tag3.ID})
	require.NoError(t, err)
	require.Equal(t, 4, total, "いずれかのタグを持つ作品は4件")
	require.Len(t, works, 4)

	// 存在しないタグでフィルタリング
	nonExistentTagID := uuid.New()
	works, total, err = repo.GetAllPublic(ctx, 10, 0, []uuid.UUID{nonExistentTagID})
	require.NoError(t, err)
	require.Equal(t, 0, total, "存在しないタグでは0件")
	require.Len(t, works, 0)

	// タグフィルタなし（nil）は全作品を取得
	works, total, err = repo.GetAllPublic(ctx, 10, 0, nil)
	require.NoError(t, err)
	require.Equal(t, 4, total, "フィルタなしでは全4件")
	require.Len(t, works, 4)

	// 空のタグスライスも全作品を取得
	works, total, err = repo.GetAllPublic(ctx, 10, 0, []uuid.UUID{})
	require.NoError(t, err)
	require.Equal(t, 4, total, "空のタグスライスでも全4件")
	require.Len(t, works, 4)
}

func TestWorkRepository_GetAll_WithTagFilter(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag1 := insertTestTag(t, db, "frontend")
	tag2 := insertTestTag(t, db, "backend")

	// tag1を持つ公開作品
	asset1 := insertTestAsset(t, db, user.ID)
	publicWork := newTestWork(user.ID, "frontend-public")
	publicWork.Visibility = "public"
	publicWork.Assets = []*entity.Asset{asset1}
	publicWork.TagIDs = []uuid.UUID{tag1.ID}
	publicWork.Tags = []*entity.Tag{tag1}
	_, err := repo.Create(ctx, publicWork)
	require.NoError(t, err)

	// tag2を持つ非公開作品
	asset2 := insertTestAsset(t, db, user.ID)
	privateWork := newTestWork(user.ID, "backend-private")
	privateWork.Visibility = "private"
	privateWork.Assets = []*entity.Asset{asset2}
	privateWork.TagIDs = []uuid.UUID{tag2.ID}
	privateWork.Tags = []*entity.Tag{tag2}
	_, err = repo.Create(ctx, privateWork)
	require.NoError(t, err)

	// GetAll（認証済みユーザー向け）でtag1フィルタ
	works, total, err := repo.GetAll(ctx, 10, 0, []uuid.UUID{tag1.ID})
	require.NoError(t, err)
	require.Equal(t, 1, total, "tag1を持つ作品は1件")
	require.Len(t, works, 1)
	require.Equal(t, "frontend-public", works[0].Title)

	// GetAllでtag2フィルタ
	works, total, err = repo.GetAll(ctx, 10, 0, []uuid.UUID{tag2.ID})
	require.NoError(t, err)
	require.Equal(t, 1, total, "tag2を持つ作品は1件")
	require.Len(t, works, 1)
	require.Equal(t, "backend-private", works[0].Title)

	// GetAllでOR検索（tag1またはtag2）
	works, total, err = repo.GetAll(ctx, 10, 0, []uuid.UUID{tag1.ID, tag2.ID})
	require.NoError(t, err)
	require.Equal(t, 2, total, "tag1またはtag2を持つ作品は2件")
	require.Len(t, works, 2)
}

func TestWorkRepository_GetAllPublic_WithPagination(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test-tag")

	// 公開作品を5つ作成
	for i := 0; i < 5; i++ {
		asset := insertTestAsset(t, db, user.ID)
		work := newTestWork(user.ID, "public-title-"+uuid.NewString())
		work.Visibility = "public"
		work.Assets = []*entity.Asset{asset}
		work.TagIDs = []uuid.UUID{tag.ID}
		work.Tags = []*entity.Tag{tag}
		work.CreatedAt = work.CreatedAt.Add(time.Duration(i) * time.Minute)
		work.UpdatedAt = work.CreatedAt
		_, err := repo.Create(ctx, work)
		require.NoError(t, err)
	}

	// ページネーション: limit=2, offset=0
	works, total, err := repo.GetAllPublic(ctx, 2, 0, nil)
	require.NoError(t, err)
	require.Equal(t, 5, total, "全体の公開作品数は5")
	require.Len(t, works, 2, "limit=2なので2件取得")

	// ページネーション: limit=2, offset=2
	works, total, err = repo.GetAllPublic(ctx, 2, 2, nil)
	require.NoError(t, err)
	require.Equal(t, 5, total, "全体の公開作品数は5")
	require.Len(t, works, 2, "limit=2なので2件取得")

	// ページネーション: limit=2, offset=4
	works, total, err = repo.GetAllPublic(ctx, 2, 4, nil)
	require.NoError(t, err)
	require.Equal(t, 5, total, "全体の公開作品数は5")
	require.Len(t, works, 1, "残り1件のみ取得")
}

func TestWorkRepository_GetAllPublic_Empty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test-tag")

	// 非公開作品のみ作成
	asset := insertTestAsset(t, db, user.ID)
	privateWork := newTestWork(user.ID, "private-title")
	privateWork.Visibility = "private"
	privateWork.Assets = []*entity.Asset{asset}
	privateWork.TagIDs = []uuid.UUID{tag.ID}
	privateWork.Tags = []*entity.Tag{tag}
	_, err := repo.Create(ctx, privateWork)
	require.NoError(t, err)

	// 公開作品がない場合
	works, total, err := repo.GetAllPublic(ctx, 10, 0, nil)
	require.NoError(t, err)
	require.Equal(t, 0, total, "公開作品が0件")
	require.Len(t, works, 0, "空のスライスが返される")
}

func TestWorkRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	require.ErrorIs(t, err, domainerrors.ErrWorkNotFound)
}

func TestWorkRepository_GetByUserID_Public(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test-tag")

	// 公開作品を2つ作成
	asset1 := insertTestAsset(t, db, user.ID)
	publicWork1 := newTestWork(user.ID, "public-work-1")
	publicWork1.Visibility = "public"
	publicWork1.Assets = []*entity.Asset{asset1}
	publicWork1.TagIDs = []uuid.UUID{tag.ID}
	publicWork1.Tags = []*entity.Tag{tag}
	publicWork1.CreatedAt = publicWork1.CreatedAt.Add(1 * time.Minute)
	publicWork1.UpdatedAt = publicWork1.CreatedAt
	_, err := repo.Create(ctx, publicWork1)
	require.NoError(t, err)

	asset2 := insertTestAsset(t, db, user.ID)
	publicWork2 := newTestWork(user.ID, "public-work-2")
	publicWork2.Visibility = "public"
	publicWork2.Assets = []*entity.Asset{asset2}
	publicWork2.TagIDs = []uuid.UUID{tag.ID}
	publicWork2.Tags = []*entity.Tag{tag}
	_, err = repo.Create(ctx, publicWork2)
	require.NoError(t, err)

	// 非公開作品を1つ作成
	asset3 := insertTestAsset(t, db, user.ID)
	privateWork := newTestWork(user.ID, "private-work")
	privateWork.Visibility = "private"
	privateWork.Assets = []*entity.Asset{asset3}
	privateWork.TagIDs = []uuid.UUID{tag.ID}
	privateWork.Tags = []*entity.Tag{tag}
	_, err = repo.Create(ctx, privateWork)
	require.NoError(t, err)

	// 下書き作品を1つ作成
	asset4 := insertTestAsset(t, db, user.ID)
	draftWork := newTestWork(user.ID, "draft-work")
	draftWork.Visibility = "draft"
	draftWork.Assets = []*entity.Asset{asset4}
	draftWork.TagIDs = []uuid.UUID{tag.ID}
	draftWork.Tags = []*entity.Tag{tag}
	_, err = repo.Create(ctx, draftWork)
	require.NoError(t, err)

	// public=trueの場合、公開作品のみ取得
	works, err := repo.GetByUserID(ctx, user.ID, true)
	require.NoError(t, err)
	require.Len(t, works, 2, "公開作品のみ取得される")

	// 全ての取得した作品が公開であることを確認
	for _, work := range works {
		require.Equal(t, "public", work.Visibility, "取得した作品は全て公開である")
		require.Equal(t, user.ID, work.UserID, "全ての作品が指定したユーザーのもの")
	}
}

func TestWorkRepository_GetByUserID_All(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test-tag")

	// 公開作品を1つ作成
	asset1 := insertTestAsset(t, db, user.ID)
	publicWork := newTestWork(user.ID, "public-work")
	publicWork.Visibility = "public"
	publicWork.Assets = []*entity.Asset{asset1}
	publicWork.TagIDs = []uuid.UUID{tag.ID}
	publicWork.Tags = []*entity.Tag{tag}
	_, err := repo.Create(ctx, publicWork)
	require.NoError(t, err)

	// 非公開作品を1つ作成
	asset2 := insertTestAsset(t, db, user.ID)
	privateWork := newTestWork(user.ID, "private-work")
	privateWork.Visibility = "private"
	privateWork.Assets = []*entity.Asset{asset2}
	privateWork.TagIDs = []uuid.UUID{tag.ID}
	privateWork.Tags = []*entity.Tag{tag}
	_, err = repo.Create(ctx, privateWork)
	require.NoError(t, err)

	// 下書き作品を1つ作成
	asset3 := insertTestAsset(t, db, user.ID)
	draftWork := newTestWork(user.ID, "draft-work")
	draftWork.Visibility = "draft"
	draftWork.Assets = []*entity.Asset{asset3}
	draftWork.TagIDs = []uuid.UUID{tag.ID}
	draftWork.Tags = []*entity.Tag{tag}
	_, err = repo.Create(ctx, draftWork)
	require.NoError(t, err)

	// public=falseの場合、公開・非公開両方取得（下書きは除外）
	works, err := repo.GetByUserID(ctx, user.ID, false)
	require.NoError(t, err)
	require.Len(t, works, 2, "公開・非公開作品が取得される（下書きは除外）")

	// 取得した作品の可視性を確認
	visibilities := make(map[string]bool)
	for _, work := range works {
		visibilities[work.Visibility] = true
		require.Equal(t, user.ID, work.UserID, "全ての作品が指定したユーザーのもの")
	}
	require.True(t, visibilities["public"], "公開作品が含まれる")
	require.True(t, visibilities["private"], "非公開作品が含まれる")
	require.False(t, visibilities["draft"], "下書きは含まれない")
}

func TestWorkRepository_GetByUserID_Empty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)

	// 作品を作成しない状態でテスト
	works, err := repo.GetByUserID(ctx, user.ID, true)
	require.NoError(t, err)
	require.Len(t, works, 0, "作品が0件の場合は空のスライスが返される")

	works, err = repo.GetByUserID(ctx, user.ID, false)
	require.NoError(t, err)
	require.Len(t, works, 0, "作品が0件の場合は空のスライスが返される")
}

func TestWorkRepository_GetByUserID_DifferentUsers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user1 := insertTestUser(t, db)
	user2 := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test-tag")

	// user1の作品を作成
	asset1 := insertTestAsset(t, db, user1.ID)
	work1 := newTestWork(user1.ID, "user1-work")
	work1.Visibility = "public"
	work1.Assets = []*entity.Asset{asset1}
	work1.TagIDs = []uuid.UUID{tag.ID}
	work1.Tags = []*entity.Tag{tag}
	_, err := repo.Create(ctx, work1)
	require.NoError(t, err)

	// user2の作品を作成
	asset2 := insertTestAsset(t, db, user2.ID)
	work2 := newTestWork(user2.ID, "user2-work")
	work2.Visibility = "public"
	work2.Assets = []*entity.Asset{asset2}
	work2.TagIDs = []uuid.UUID{tag.ID}
	work2.Tags = []*entity.Tag{tag}
	_, err = repo.Create(ctx, work2)
	require.NoError(t, err)

	// user1の作品のみ取得
	works, err := repo.GetByUserID(ctx, user1.ID, true)
	require.NoError(t, err)
	require.Len(t, works, 1, "user1の作品のみ取得される")
	require.Equal(t, user1.ID, works[0].UserID)
	require.Equal(t, "user1-work", works[0].Title)

	// user2の作品のみ取得
	works, err = repo.GetByUserID(ctx, user2.ID, true)
	require.NoError(t, err)
	require.Len(t, works, 1, "user2の作品のみ取得される")
	require.Equal(t, user2.ID, works[0].UserID)
	require.Equal(t, "user2-work", works[0].Title)
}

func TestWorkRepository_ExistsByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test")

	work := newTestWork(user.ID, "exists-title")
	work.TagIDs = []uuid.UUID{tag.ID}
	work.Tags = []*entity.Tag{tag}
	created, err := repo.Create(ctx, work)
	require.NoError(t, err)

	exists, err := repo.ExistsById(ctx, created.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsById(ctx, uuid.New())
	require.NoError(t, err)
	require.False(t, exists)
}

func TestWorkRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag1 := insertTestTag(t, db, "original-tag")
	tag2 := insertTestTag(t, db, "updated-tag")

	workEntity := newTestWork(user.ID, "original-title")
	workEntity.Description = "original description"
	workEntity.TagIDs = []uuid.UUID{tag1.ID}
	workEntity.Tags = []*entity.Tag{tag1}
	created, err := repo.Create(ctx, workEntity)
	require.NoError(t, err)

	created.Title = "updated-title"
	created.Description = "updated description"
	created.Visibility = "private"
	created.TagIDs = []uuid.UUID{tag2.ID}
	created.Tags = []*entity.Tag{tag2}

	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, "updated-title", updated.Title)
	require.Equal(t, "updated description", updated.Description)
	require.Equal(t, "private", updated.Visibility)
	require.Equal(t, 1, len(updated.Tags))
	require.Equal(t, tag2.Name, updated.Tags[0].Name)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "updated-title", fetched.Title)
	require.Equal(t, "updated description", fetched.Description)
	require.Equal(t, "private", fetched.Visibility)
}

func TestWorkRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test-tag")
	asset := insertTestAsset(t, db, user.ID)

	workEntity := newTestWork(user.ID, "to-be-deleted")
	workEntity.TagIDs = []uuid.UUID{tag.ID}
	workEntity.Tags = []*entity.Tag{tag}
	workEntity.Assets = []*entity.Asset{asset}
	created, err := repo.Create(ctx, workEntity)
	require.NoError(t, err)

	err = repo.Delete(ctx, created.ID, user.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, created.ID)
	require.ErrorIs(t, err, domainerrors.ErrWorkNotFound)

	exists, err := repo.ExistsById(ctx, created.ID)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestWorkRepository_Delete_NotOwned(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user1 := insertTestUser(t, db)
	user2 := insertTestUser(t, db)

	workEntity := newTestWork(user1.ID, "user1-work")
	created, err := repo.Create(ctx, workEntity)
	require.NoError(t, err)

	err = repo.Delete(ctx, created.ID, user2.ID)
	require.ErrorIs(t, err, domainerrors.ErrWorkNotOwnedByUser)

	exists, err := repo.ExistsById(ctx, created.ID)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestWorkRepository_Delete_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	user := insertTestUser(t, db)

	nonExistentID := uuid.New()
	err := repo.Delete(ctx, nonExistentID, user.ID)
	require.ErrorIs(t, err, domainerrors.ErrWorkNotFound)
}

func insertTestUser(t *testing.T, db *bun.DB) *entity.User {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	userID := uuid.New()
	// UUIDの最初の8文字を使用して短い識別子を作成
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

func newTestWork(userID uuid.UUID, title string) *entity.Work {
	now := time.Now().UTC().Truncate(time.Second)
	return &entity.Work{
		ID:               uuid.New(),
		Title:            title,
		Description:      "description",
		UserID:           userID,
		Visibility:       "public",
		ThumbnailAssetID: uuid.Nil,
		Assets:           []*entity.Asset{},
		URLs:             []*string{},
		TagIDs:           []uuid.UUID{},
		Tags:             []*entity.Tag{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
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

func TestWorkRepository_Create_WithCollaborators(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	owner := insertTestUser(t, db)
	collaborator1 := insertTestUser(t, db)
	collaborator2 := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test")

	work := newTestWork(owner.ID, "work-with-collaborators")
	work.TagIDs = []uuid.UUID{tag.ID}
	work.Tags = []*entity.Tag{tag}
	work.Collaborators = []*entity.User{collaborator1, collaborator2}

	created, err := repo.Create(ctx, work)
	require.NoError(t, err)
	require.Equal(t, 2, len(created.Collaborators))

	// DBから取得して共同制作者が正しく保存されているか確認
	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(fetched.Collaborators))

	// 共同制作者のIDを確認
	collaboratorIDs := make(map[uuid.UUID]bool)
	for _, c := range fetched.Collaborators {
		collaboratorIDs[c.ID] = true
	}
	require.True(t, collaboratorIDs[collaborator1.ID])
	require.True(t, collaboratorIDs[collaborator2.ID])
}

func TestWorkRepository_GetByID_WithCollaborators(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	owner := insertTestUser(t, db)
	collaborator := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test")

	work := newTestWork(owner.ID, "work-with-one-collaborator")
	work.TagIDs = []uuid.UUID{tag.ID}
	work.Tags = []*entity.Tag{tag}
	work.Collaborators = []*entity.User{collaborator}

	created, err := repo.Create(ctx, work)
	require.NoError(t, err)

	// GetByIDで共同制作者がロードされることを確認
	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, 1, len(fetched.Collaborators))
	require.Equal(t, collaborator.ID, fetched.Collaborators[0].ID)
	require.Equal(t, collaborator.DisplayName, fetched.Collaborators[0].DisplayName)
}

func TestWorkRepository_Update_Collaborators(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	owner := insertTestUser(t, db)
	collaborator1 := insertTestUser(t, db)
	collaborator2 := insertTestUser(t, db)
	collaborator3 := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test")

	// 共同制作者1,2で作品作成
	work := newTestWork(owner.ID, "work-to-update-collaborators")
	work.TagIDs = []uuid.UUID{tag.ID}
	work.Tags = []*entity.Tag{tag}
	work.Collaborators = []*entity.User{collaborator1, collaborator2}

	created, err := repo.Create(ctx, work)
	require.NoError(t, err)
	require.Equal(t, 2, len(created.Collaborators))

	// 共同制作者を2,3に更新
	created.Collaborators = []*entity.User{collaborator2, collaborator3}
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, 2, len(updated.Collaborators))

	// DBから取得して確認
	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(fetched.Collaborators))

	// 共同制作者が正しく更新されているか確認
	collaboratorIDs := make(map[uuid.UUID]bool)
	for _, c := range fetched.Collaborators {
		collaboratorIDs[c.ID] = true
	}
	require.False(t, collaboratorIDs[collaborator1.ID], "collaborator1は削除されている")
	require.True(t, collaboratorIDs[collaborator2.ID], "collaborator2は残っている")
	require.True(t, collaboratorIDs[collaborator3.ID], "collaborator3が追加されている")
}

func TestWorkRepository_Delete_CascadeCollaborators(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := work.NewWorkRepository(db)

	ctx := context.Background()
	owner := insertTestUser(t, db)
	collaborator := insertTestUser(t, db)
	tag := insertTestTag(t, db, "test")

	work := newTestWork(owner.ID, "work-to-delete-with-collaborators")
	work.TagIDs = []uuid.UUID{tag.ID}
	work.Tags = []*entity.Tag{tag}
	work.Collaborators = []*entity.User{collaborator}

	created, err := repo.Create(ctx, work)
	require.NoError(t, err)

	// 作品削除
	err = repo.Delete(ctx, created.ID, owner.ID)
	require.NoError(t, err)

	// collaboratorテーブルのレコードも削除されていることを確認
	count, err := db.NewSelect().
		Table("collaborator").
		Where("work_id = ?", created.ID).
		Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, count, "作品削除時にcollaboratorレコードも削除される")
}
