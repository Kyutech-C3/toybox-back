package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	ENV                      string
	FRONTEND_URL             []string
	DB_DSN                   string
	POSTGRES_USER            string
	POSTGRES_PASSWORD        string
	POSTGRES_DB              string
	POSTGRES_HOST            string
	AUTHENTICATOR_URL        string
	S3_BUCKET                string
	S3_DIR                   string
	S3_BASE_URL              string
	REGION_NAME              string
	LEGACY_TOYBOX_BASE_URL   string
	LEGACY_TOYBOX_PROXY_HOST string
	WEBHOOK_URL              string
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
	AUTHENTICATOR_URL = os.Getenv("AUTHENTICATOR_URL")
	S3_BUCKET = os.Getenv("S3_BUCKET")
	S3_DIR = os.Getenv("S3_DIR")
	S3_BASE_URL = os.Getenv("S3_BASE_URL")
	REGION_NAME = os.Getenv("REGION_NAME")
	LEGACY_TOYBOX_BASE_URL = os.Getenv("LEGACY_TOYBOX_BASE_URL")
	LEGACY_TOYBOX_PROXY_HOST = os.Getenv("LEGACY_TOYBOX_PROXY_HOST")
	WEBHOOK_URL = os.Getenv("WEBHOOK_URL")
}
