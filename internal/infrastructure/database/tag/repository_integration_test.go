//go:build integration

package tag_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/tag"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/testutil"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testutil.Teardown()
	os.Exit(code)
}

func TestTagRepository_ExistAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := tag.NewTagRepository(db)

	ctx := context.Background()
	tag1ID := insertTestTag(t, db, "go")
	tag2ID := insertTestTag(t, db, "rust")
	tag3ID := insertTestTag(t, db, "python")

	tests := []struct {
		name    string
		ids     []uuid.UUID
		want    bool
		wantErr bool
	}{
		{
			name: "正常系: 全てのタグが存在する",
			ids:  []uuid.UUID{tag1ID, tag2ID},
			want: true,
		},
		{
			name: "正常系: 1つのタグが存在する",
			ids:  []uuid.UUID{tag1ID},
			want: true,
		},
		{
			name: "正常系: 空のスライス",
			ids:  []uuid.UUID{},
			want: true,
		},
		{
			name: "異常系: 存在しないタグが含まれる",
			ids:  []uuid.UUID{tag1ID, uuid.New()},
			want: false,
		},
		{
			name: "異常系: 全て存在しない",
			ids:  []uuid.UUID{uuid.New(), uuid.New()},
			want: false,
		},
		{
			name: "正常系: 全てのタグが存在する（3つ）",
			ids:  []uuid.UUID{tag1ID, tag2ID, tag3ID},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.ExistAll(ctx, tt.ids)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, exists)
			}
		})
	}
}

func TestTagRepository_FindAllByIDs(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := tag.NewTagRepository(db)

	ctx := context.Background()
	tag1ID := insertTestTag(t, db, "go")
	tag2ID := insertTestTag(t, db, "rust")
	tag3ID := insertTestTag(t, db, "python")

	tests := []struct {
		name      string
		ids       []uuid.UUID
		wantCount int
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "正常系: 2つのタグを取得",
			ids:       []uuid.UUID{tag1ID, tag2ID},
			wantCount: 2,
			wantNames: []string{"go", "rust"},
		},
		{
			name:      "正常系: 1つのタグを取得",
			ids:       []uuid.UUID{tag3ID},
			wantCount: 1,
			wantNames: []string{"python"},
		},
		{
			name:      "正常系: 全てのタグを取得",
			ids:       []uuid.UUID{tag1ID, tag2ID, tag3ID},
			wantCount: 3,
			wantNames: []string{"go", "rust", "python"},
		},
		{
			name:      "正常系: 存在しないIDを含む（部分一致）",
			ids:       []uuid.UUID{tag1ID, uuid.New()},
			wantCount: 1,
			wantNames: []string{"go"},
		},
		{
			name:      "正常系: 全て存在しないID",
			ids:       []uuid.UUID{uuid.New(), uuid.New()},
			wantCount: 0,
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := repo.FindAllByIDs(ctx, tt.ids)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantCount, len(tags))

				// タグ名の確認
				tagNames := make([]string, len(tags))
				for i, tag := range tags {
					tagNames[i] = tag.Name
				}
				for _, wantName := range tt.wantNames {
					require.Contains(t, tagNames, wantName)
				}
			}
		})
	}
}

func TestTagRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := tag.NewTagRepository(db)

	ctx := context.Background()

	tests := []struct {
		name    string
		tagName string
		wantErr bool
	}{
		{
			name:    "正常系: タグ作成成功",
			tagName: "新しいタグ",
			wantErr: false,
		},
		{
			name:    "正常系: 日本語タグ名",
			tagName: "日本語タグ",
			wantErr: false,
		},
		{
			name:    "正常系: 英数字タグ名",
			tagName: "tag123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Second)
			inputTag := &entity.Tag{
				ID:        uuid.New(),
				Name:      tt.tagName,
				CreatedAt: now,
				UpdatedAt: now,
			}

			createdTag, err := repo.Create(ctx, inputTag)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, createdTag)
				require.Equal(t, inputTag.ID, createdTag.ID)
				require.Equal(t, tt.tagName, createdTag.Name)
			}
		})
	}
}

func TestTagRepository_FindAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := tag.NewTagRepository(db)

	ctx := context.Background()

	// テスト用のタグを挿入（名前順でソートされることを確認するため、順序をバラバラに）
	insertTestTag(t, db, "Rust")
	insertTestTag(t, db, "Go")
	insertTestTag(t, db, "Python")

	tests := []struct {
		name           string
		wantMinCount   int
		wantSortedByName bool
	}{
		{
			name:             "正常系: 全タグ取得（名前順ソート）",
			wantMinCount:     3,
			wantSortedByName: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := repo.FindAll(ctx)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(tags), tt.wantMinCount)

			if tt.wantSortedByName && len(tags) > 1 {
				// 名前順にソートされていることを確認
				for i := 0; i < len(tags)-1; i++ {
					require.LessOrEqual(t, tags[i].Name, tags[i+1].Name, "タグが名前順でソートされていません")
				}
			}
		})
	}
}

func insertTestTag(t *testing.T, db *bun.DB, name string) uuid.UUID {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	tag := &dto.Tag{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := db.NewInsert().Model(tag).Exec(context.Background())
	require.NoError(t, err)

	return tag.ID
}
