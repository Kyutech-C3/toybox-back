package migration

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/simesaba80/toybox-back/internal/domain/entity"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
	"github.com/uptrace/bun"
)

func MigrateTags(ctx context.Context, sourceDB, targetDB bun.IDB) (map[uuid.UUID]uuid.UUID, error) {
	fmt.Println("Migrating tags...")

	var oldTags []OldTag
	if err := sourceDB.NewSelect().Model(&oldTags).Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to select old tags: %w", err)
	}

	normalizedNameToRepresentativeId := make(map[string]uuid.UUID)
	oldIdToRepresentativeId := make(map[uuid.UUID]uuid.UUID)
	representativeTagsToInsert := make([]*dto.Tag, 0)

	for _, old := range oldTags {
		parsedOldID, err := uuid.Parse(old.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tag UUID %s: %w", old.ID, err)
		}

		convertedName := entity.NormalizeTagName(old.Name)

		representativeID, exists := normalizedNameToRepresentativeId[convertedName]
		if !exists {
			representativeID = parsedOldID
			normalizedNameToRepresentativeId[convertedName] = representativeID
			representativeTagsToInsert = append(representativeTagsToInsert, &dto.Tag{
				ID:        representativeID,
				Name:      convertedName,
				CreatedAt: old.CreatedAt,
				UpdatedAt: old.UpdatedAt,
			})
		}
		oldIdToRepresentativeId[parsedOldID] = representativeID
	}

	if len(representativeTagsToInsert) > 0 {
		if _, err := targetDB.NewInsert().Model(&representativeTagsToInsert).Exec(ctx); err != nil {
			return nil, fmt.Errorf("failed to insert new representative tags: %w", err)
		}
	}

	fmt.Printf("-> Migrated %d tags.\n", len(representativeTagsToInsert))
	return oldIdToRepresentativeId, nil
}
