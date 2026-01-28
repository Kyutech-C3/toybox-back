//go:generate wire
//go:build wireinject
// +build wireinject

package di

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"

	"github.com/simesaba80/toybox-back/internal/domain/repository"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/asset"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/comment"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/favorite"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/tag"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/token"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/user"
	"github.com/simesaba80/toybox-back/internal/infrastructure/database/work"
	customejwt "github.com/simesaba80/toybox-back/internal/infrastructure/external/custome-jwt"
	"github.com/simesaba80/toybox-back/internal/infrastructure/external/oauth"
	"github.com/simesaba80/toybox-back/internal/infrastructure/router"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/usecase"
	"github.com/simesaba80/toybox-back/pkg/db"
	"github.com/simesaba80/toybox-back/pkg/s3_client"
)

var RepositorySet = wire.NewSet(
	user.NewUserRepository,
	wire.Bind(new(repository.UserRepository), new(*user.UserRepository)),
	work.NewWorkRepository,
	wire.Bind(new(repository.WorkRepository), new(*work.WorkRepository)),
	comment.NewCommentRepository,
	wire.Bind(new(repository.CommentRepository), new(*comment.CommentRepository)),
	oauth.NewDiscordRepository,
	wire.Bind(new(repository.DiscordRepository), new(*oauth.DiscordRepository)),
	token.NewTokenRepository,
	wire.Bind(new(repository.TokenRepository), new(*token.TokenRepository)),
	asset.NewAssetRepository,
	wire.Bind(new(repository.AssetRepository), new(*asset.AssetRepository)),
	favorite.NewFavoriteRepository,
	wire.Bind(new(repository.FavoriteRepository), new(*favorite.FavoriteRepository)),
	tag.NewTagRepository,
	wire.Bind(new(repository.TagRepository), new(*tag.TagRepository)),
)

var UseCaseSet = wire.NewSet(
	ProvideUserUseCase,
	ProvideWorkUseCase,
	ProvideCommentUseCase,
	ProvideAuthUseCase,
	ProvideTokenProvider,
	ProvideAssetUseCase,
	ProvideFavoriteUseCase,
)

var ControllerSet = wire.NewSet(
	controller.NewUserController,
	controller.NewWorkController,
	controller.NewCommentController,
	controller.NewAuthController,
	controller.NewAssetController,
	controller.NewFavoriteController,
)

var InfrastructureSet = wire.NewSet(
	ProvideDatabase,
	ProvideS3Client,
	router.NewRouter,
	ProvideEcho,
)

// ProviderSet は依存関係を定義します
var ProviderSet = wire.NewSet(
	RepositorySet,
	UseCaseSet,
	ControllerSet,
	InfrastructureSet,
	NewApp, // App構造体のコンストラクタを追加
)

// ProvideDatabase はデータベース接続を提供します
func ProvideDatabase() *bun.DB {
	db.Init()
	return db.DB
}

func ProvideS3Client() *s3.Client {
	s3_client.Init()
	return s3_client.Client
}

// ProvideUserUseCase はUserUseCaseを提供します
func ProvideUserUseCase(repo repository.UserRepository) usecase.IUserUseCase {
	return usecase.NewUserUseCase(repo)
}

// ProvideWorkUseCase はWorkUseCaseを提供します
func ProvideWorkUseCase(workRepo repository.WorkRepository, tagRepo repository.TagRepository, assetRepo repository.AssetRepository, userRepo repository.UserRepository) usecase.IWorkUseCase {
	return usecase.NewWorkUseCase(workRepo, tagRepo, assetRepo, userRepo)
}

// ProvideCommentUseCase はCommentUseCaseを提供します
func ProvideCommentUseCase(commentRepo repository.CommentRepository, workRepo repository.WorkRepository) usecase.ICommentUsecase {
	return usecase.NewCommentUsecase(commentRepo, workRepo, 30*time.Second)
}

// ProvideDiscordUseCase はDiscordUseCaseを提供します
func ProvideAuthUseCase(authRepo repository.DiscordRepository, userRepo repository.UserRepository, tokenProvider usecase.TokenProvider, tokenRepo repository.TokenRepository, assetRepo repository.AssetRepository) usecase.IAuthUsecase {
	return usecase.NewAuthUsecase(authRepo, userRepo, tokenProvider, tokenRepo, assetRepo)
}

// ProvideTokenProvider はTokenProviderを提供します
func ProvideTokenProvider() usecase.TokenProvider {
	return tokenProviderFunc(customejwt.GenerateToken)
}

type tokenProviderFunc func(userID uuid.UUID) (string, error)

func (f tokenProviderFunc) GenerateToken(userID uuid.UUID) (string, error) {
	return f(userID)
}

// ProvideAssetUseCase はAssetUseCaseを提供します
func ProvideAssetUseCase(assetRepo repository.AssetRepository) usecase.IAssetUseCase {
	return usecase.NewAssetUseCase(assetRepo)
}

// ProvideFavoriteUseCase はFavoriteUseCaseを提供します
func ProvideFavoriteUseCase(favoriteRepo repository.FavoriteRepository) usecase.IFavoriteUsecase {
	return usecase.NewFavoriteUsecase(favoriteRepo)
}

// ProvideEcho はEchoインスタンスを提供します
func ProvideEcho() *echo.Echo {
	return echo.New()
}

// NewApp はAppインスタンスを作成します
func NewApp(router *router.Router, database *bun.DB, s3Client *s3.Client) *App {
	return &App{
		Router:   router,
		Database: database,
		S3Client: s3Client,
	}
}

// InitializeApp はアプリケーションを初期化します
func InitializeApp() (*App, func(), error) {
	wire.Build(ProviderSet)
	return nil, nil, nil
}

type App struct {
	Router   *router.Router
	Database *bun.DB
	S3Client *s3.Client
}

// Start アプリケーションの開始
func (app *App) Start() *echo.Echo {
	return app.Router.Setup()
}

// Cleanup アプリケーションのクリーンアップ
func (app *App) Cleanup() {
	if app.Database != nil {
		app.Database.Close()
	}
}
