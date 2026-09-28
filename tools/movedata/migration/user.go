package migration

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/uptrace/bun"
)

func MigrateUsers(ctx context.Context, sourceDB, targetDB bun.IDB) error {
	fmt.Println("Migrating users...")

	var oldUsers []OldUser
	if err := sourceDB.NewSelect().Model(&oldUsers).Scan(ctx); err != nil {
		return fmt.Errorf("failed to select old users: %w", err)
	}

	newUsers := make([]*dto.User, 0, len(oldUsers))
	for _, old := range oldUsers {
		parsedOldID, err := uuid.Parse(old.ID)
		if err != nil {
			return fmt.Errorf("failed to parse user UUID %s: %w", old.ID, err)
		}

		newUser := &dto.User{
			ID:            parsedOldID,
			Name:          old.Name,
			Email:         old.Email,
			DisplayName:   old.DisplayName,
			DiscordUserID: derefString(old.DiscordUserID),
			Profile:       derefString(old.Profile),
			AvatarURL:     derefString(old.AvatarURL),
			XUsername:     derefString(old.TwitterID),
			GithubID:      derefString(old.GithubID),
			CreatedAt:     old.CreatedAt,
			UpdatedAt:     old.UpdatedAt,
		}
		newUsers = append(newUsers, newUser)
	}

	if len(newUsers) > 0 {
		if _, err := targetDB.NewInsert().Model(&newUsers).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert new users: %w", err)
		}
	}

	fmt.Printf("-> Migrated %d users.\n", len(newUsers))
	return nil
}
