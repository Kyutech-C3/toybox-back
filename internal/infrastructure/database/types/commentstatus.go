package types

import (
	"database/sql/driver"
	"fmt"
)

type CommentStatus string

const (
	CommentStatusActive  CommentStatus = "active"
	CommentStatusDeleted CommentStatus = "deleted"
)

// Value はdatabase/sql/driver.Valuerインターフェースを実装
func (s CommentStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan はsql.Scannerインターフェースを実装
func (s *CommentStatus) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*s = CommentStatus(string(v))
		return nil
	default:
		return fmt.Errorf("cannot scan %T into CommentStatus", value)
	}
}