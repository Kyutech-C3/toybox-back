//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/simesaba80/toybox-back/db/migrations"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/dto"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	iofs "github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

const (
	pgImage    = "postgres:17.6"
	pgUser     = "toybox"
	pgPassword = "toybox"
	pgDatabase = "toybox_test"

	pgPort = "5432/tcp"
)

var (
	setupOnce   sync.Once
	dbInstance  *bun.DB
	sqlInstance *sql.DB
	container   testcontainers.Container
	initErr     error
)

// SetupTestDB はテスト用PostgreSQLを起動し、bun.DB を返します。
func SetupTestDB(tb testing.TB) *bun.DB {
	tb.Helper()

	setupOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		var dsn string
		container, dsn, initErr = startPostgresContainer(ctx)
		if initErr != nil {
			return
		}

		if err := runMigrations(dsn); err != nil {
			initErr = err
			return
		}

		sqlInstance = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
		if err := sqlInstance.PingContext(ctx); err != nil {
			initErr = err
			return
		}

		dbInstance = bun.NewDB(sqlInstance, pgdialect.New())

		dbInstance.RegisterModel((*dto.Tagging)(nil))
		dbInstance.RegisterModel((*dto.Collaborator)(nil))
	})

	if initErr != nil {
		tb.Fatalf("テストDB初期化に失敗しました: %v", initErr)
	}

	ResetTables(tb)

	return dbInstance
}

func startPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        pgImage,
		ExposedPorts: []string{pgPort},
		Env: map[string]string{
			"POSTGRES_USER":     pgUser,
			"POSTGRES_PASSWORD": pgPassword,
			"POSTGRES_DB":       pgDatabase,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort(pgPort),
		).WithStartupTimeout(2 * time.Minute),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch container host: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, pgPort)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch container port: %w", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUser, pgPassword, host, mappedPort.Port(), pgDatabase)
	return c, dsn, nil
}

func runMigrations(dsn string) error {
	source, err := iofs.New(migrations.EmbedFiles, ".")
	if err != nil {
		return fmt.Errorf("migrations source: %w", err)
	}

	migrationDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	defer migrationDB.Close()

	driver, err := postgres.WithInstance(migrationDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

// ResetTables は主要テーブルをTRUNCATEし、テスト間の副作用を防ぎます。
func ResetTables(tb testing.TB) {
	tb.Helper()

	if dbInstance == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tables := []string{
		"thumbnail",
		"tagging",
		"collaborator",
		"tag",
		"urlinfo",
		"comment",
		"asset",
		"favorite",
		"work",
		`"user"`,
		"token",
	}

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))
	if _, err := dbInstance.NewRaw(query).Exec(ctx); err != nil {
		tb.Fatalf("テーブルのクリーンアップに失敗しました: %v", err)
	}
}

// Teardown はコンテナとDB接続を解放します。各パッケージの TestMain から呼び出してください。
func Teardown() {
	if dbInstance != nil {
		_ = dbInstance.Close()
		dbInstance = nil
	}

	if sqlInstance != nil {
		_ = sqlInstance.Close()
		sqlInstance = nil
	}

	if container != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = container.Terminate(ctx)
		container = nil
	}

	setupOnce = sync.Once{}
	initErr = nil
}
