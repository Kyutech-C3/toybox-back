#!/bin/bash

# 旧ToyboxDBをDockerで再現し、Goによるスクリプトで新ToyboxへDBデータを移行するスクリプト
set -e

echo "Starting database migration process..."

# 既存の.envファイルを読み込み
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi



# 既存のネットワークを作成（存在しない場合）
docker network create app-network 2>/dev/null || true

# 既存のサービスが起動していることを確認
echo "Checking if target database is running..."
if ! docker ps | grep -q "db"; then
    echo "Target database (db) is not running. Please start it first with:"
    echo "  docker-compose up -d db"
    exit 1
fi

# 移行環境を起動
echo "Starting migration containers..."
cd docker/
# 移行などは一時的なコンテナで行う
# 旧DB読み込みはrestore_backup_db.shで行う
docker compose -f compose.migration.yaml up --build --abort-on-container-exit


# クリーンアップ
echo "Cleaning up migration containers..."
docker compose -f compose.migration.yaml down --volumes


cd ../

echo "Move data completed successfully!"