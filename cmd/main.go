package main

// @title Toybox API
// @version 1.0
// @description This is the API server for the Toybox application.
// @host localhost:8080
// @BasePath /
// @SecurityDefinitions.apikey CookieAuth
// @in cookie
// @name toybox_token
// @description JWT issued by Authenticator

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/simesaba80/toybox-back/docs"
	"github.com/simesaba80/toybox-back/internal/di"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
)

func main() {
	config.LoadEnv()

	// Wireで依存性を注入してアプリケーションを初期化
	app, cleanup, err := di.InitializeApp()
	if err != nil {
		log.Fatal("Failed to initialize app:", err)
	}
	defer cleanup()

	e := app.Start()

	// OSからの割り込みシグナルをリッスンするコンテキストを作成します。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Errorf("shutting down the server: %v", err)
			stop()
		}
	}()

	// シグナルを受信するまでブロックします
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	log.Println("Process finished.")
}
