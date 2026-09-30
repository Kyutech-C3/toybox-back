package dto

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Collaborator struct {
	bun.BaseModel `bun:"table:collaborator"`
	WorkID        uuid.UUID `bun:"work_id,pk"`
	UserID        uuid.UUID `bun:"user_id,pk"`
	Work          *Work     `bun:"rel:belongs-to,join:work_id=id"`
	User          *User     `bun:"rel:belongs-to,join:user_id=id"`
}
