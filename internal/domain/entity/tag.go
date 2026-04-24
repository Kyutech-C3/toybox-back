package entity

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/width"
)

type Tag struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTag(name string) *Tag {
	return &Tag{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (t *Tag) NormalizeName() {
	var builder strings.Builder
	for _, r := range t.Name {
		// カタカナはそのまま保持（半角・全角どちらも変換しない）
		if unicode.In(r, unicode.Katakana) {
			builder.WriteRune(r)
			continue
		}
		// それ以外は半角に変換（変換できない場合は元の文字をそのまま使う）
		p := width.LookupRune(r)
		narrow := p.Narrow()
		if narrow == 0 {
			builder.WriteRune(r)
		} else {
			builder.WriteRune(narrow)
		}
	}
	t.Name = strings.ToLower(builder.String())
}
