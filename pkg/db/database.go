package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"
)

var DB *bun.DB

func Init() {
	err := godotenv.Load()

	if err != nil {
		log.Printf("読み込み出来ませんでした: %v", err)
	}

	DB_DSN := os.Getenv("DB_DSN")
	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(DB_DSN),
	))
	DB = bun.NewDB(sqlDB, pgdialect.New())

	DB.RegisterModel((*dto.Tagging)(nil))
	DB.RegisterModel((*dto.Collaborator)(nil))

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}

	// クエリーフックを追加
	DB.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
	))
}
