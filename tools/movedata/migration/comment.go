package migration

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/uptrace/bun"
)

func MigrateComments(ctx context.Context, sourceDB, targetDB bun.IDB) error {
	fmt.Println("Migrating comments...")
	var oldComments []OldComment
	if err := sourceDB.NewSelect().Model(&oldComments).Scan(ctx); err != nil {
		return fmt.Errorf("failed to select old comments: %w", err)
	}

	newComments := make([]*dto.Comment, 0, len(oldComments))
	for _, old := range oldComments {
		parsedOldID, err := uuid.Parse(old.ID)
		if err != nil {
			return fmt.Errorf("failed to parse comment UUID %s: %w", old.ID, err)
		}
		parsedOldWorkID, err := uuid.Parse(old.WorkID)
		if err != nil {
			return fmt.Errorf("failed to parse work UUID %s for comment %s: %w", old.WorkID, old.ID, err)
		}

		parsedOldUserID, err := parseNullableUUID(old.UserID)
		if err != nil {
			return fmt.Errorf("failed to parse user UUID %s for comment %s: %w", *old.UserID, old.ID, err)
		}

		newComment := &dto.Comment{
			ID:         parsedOldID,
			Content:    old.Content,
			WorkID:     parsedOldWorkID,
			UserID:     parsedOldUserID,
			ReplyAt:    derefString(old.ReplyAt),
			CreatedAt:  old.CreatedAt,
			UpdatedAt:  old.UpdatedAt,
		}
		newComments = append(newComments, newComment)
	}

	if len(newComments) > 0 {
		if _, err := targetDB.NewInsert().Model(&newComments).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert new comments: %w", err)
		}
	}

	fmt.Printf("-> Migrated %d comments.\n", len(newComments))
	return nil
}
