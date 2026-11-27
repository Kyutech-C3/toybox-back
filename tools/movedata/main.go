package main

import (
	"context"
	"fmt"
	"os"

	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
	"github.com/simesaba80/toybox-back/pkg/db"
	"github.com/simesaba80/toybox-back/tools/movedata/internal/connect"
	"github.com/simesaba80/toybox-back/tools/movedata/migration"
	"github.com/uptrace/bun"
)

func main() {
	// This is a placeholder main function.
	// The actual migration logic would be implemented here.
	config.LoadEnv()
	fmt.Println("Environment variables loaded.")
	//移行後のDB接続
	db.Init()
	fmt.Println("Connected to the database1.")
	//移行前のDB接続
	connect.Connect()
	fmt.Println("Connected to the database2.")

	ctx := context.Background()

	// Start a transaction on the target DB
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to begin transaction: %v\n", err)
		os.Exit(1)
	}

	defer tx.Rollback()

	type migrationFunc struct {
		name string
		f    func(context.Context, bun.IDB, bun.IDB) error
	}

	// Tagは旧DBから移行する際に全角半角等の修正を行うため別でマイグレーションを行う
	migrationFuncs := []migrationFunc{
		{name: "Users", f: migration.MigrateUsers},
		{name: "Works", f: migration.MigrateWorks},
		{name: "Assets", f: migration.MigrateAssets},
		{name: "Comments", f: migration.MigrateComments},
		{name: "URLInfos", f: migration.MigrateURLInfos},
		{name: "Favorites", f: migration.MigrateFavorites},
		{name: "Thumbnails", f: migration.MigrateThumbnails},
	}

	for _, mig := range migrationFuncs {
		if err := mig.f(ctx, connect.DB, tx); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: migration failed on %s: %v\n", mig.name, err)
			tx.Rollback() // Explicitly rollback before exit
			os.Exit(1)
		}
	}

	// Tag関係のマイグレーション
	fmt.Println("Running migration: Tags...")
	tagIdMap, err := migration.MigrateTags(ctx, connect.DB, tx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: migration failed on Tags: %v\n", err)
		tx.Rollback()
		os.Exit(1)
	}

	fmt.Println("Running migration: Taggings...")
	if err := migration.MigrateTaggings(ctx, connect.DB, tx, tagIdMap); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: migration failed on Taggings: %v\n", err)
		tx.Rollback()
		os.Exit(1)
	}

	// If all migrations succeed, commit the transaction
	if err := tx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to commit transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migration script executed.")
}
