package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	ENV                   string
	FRONTEND_URL          []string
	DB_DSN                string
	POSTGRES_USER         string
	POSTGRES_PASSWORD     string
	POSTGRES_DB           string
	POSTGRES_HOST         string
	DISCORD_CLIENT_ID     string
	DISCORD_CLIENT_SECRET string
	TOKEN_SECRET          string
	DISCORD_GUILD_IDS     []string
	REDIRECT_URL          string
	S3_BUCKET             string
	S3_DIR                string
	S3_BASE_URL           string
	REGION_NAME           string
)

// .envを呼び出します。
func LoadEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Printf("読み込み出来ませんでした: %v", err)
	}

	ENV = os.Getenv("ENV")
	FRONTEND_URL = strings.Split(os.Getenv("FRONTEND_URL"), ",")
	DB_DSN = os.Getenv("DB_DSN")
	POSTGRES_USER = os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD = os.Getenv("POSTGRES_PASSWORD")
	POSTGRES_DB = os.Getenv("POSTGRES_DB")
	POSTGRES_HOST = os.Getenv("POSTGRES_HOST")
	DISCORD_CLIENT_ID = os.Getenv("DISCORD_CLIENT_ID")
	DISCORD_CLIENT_SECRET = os.Getenv("DISCORD_CLIENT_SECRET")
	TOKEN_SECRET = os.Getenv("TOKEN_SECRET")
	DISCORD_GUILD_IDS = strings.Split(os.Getenv("DISCORD_GUILD_IDS"), ",")
	REDIRECT_URL = os.Getenv("REDIRECT_URL")
	S3_BUCKET = os.Getenv("S3_BUCKET")
	S3_DIR = os.Getenv("S3_DIR")
	S3_BASE_URL = os.Getenv("S3_BASE_URL")
	REGION_NAME = os.Getenv("REGION_NAME")
}
