package router

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/pkg/echovalidator"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Router struct {
	echo               *echo.Echo
	UserController     *controller.UserController
	WorkController     *controller.WorkController
	CommentController  *controller.CommentController
	AuthController     *controller.AuthController
	AssetController    *controller.AssetController
	FavoriteController *controller.FavoriteController
	TagController      *controller.TagController
}

func NewRouter(e *echo.Echo, uc *controller.UserController, wc *controller.WorkController, cc *controller.CommentController, authc *controller.AuthController, assetc *controller.AssetController, fc *controller.FavoriteController, tagc *controller.TagController) *Router {
	return &Router{
		echo:               e,
		UserController:     uc,
		WorkController:     wc,
		CommentController:  cc,
		AuthController:     authc,
		AssetController:    assetc,
		FavoriteController: fc,
		TagController:      tagc,
	}
}

func (r *Router) Setup() *echo.Echo {
	r.echo.Validator = echovalidator.NewValidator()
	r.echo.Use(middleware.Logger())
	r.echo.Use(middleware.Recover())
	r.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.FRONTEND_URL,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	r.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	r.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Auth
	r.echo.GET("/auth/discord", r.AuthController.GetDiscordAuthURL)
	r.echo.GET("/auth/discord/callback", r.AuthController.AuthenticateUser)
	r.echo.POST("/auth/refresh", r.AuthController.RegenerateToken)
	r.echo.POST("/auth/logout", r.AuthController.Logout)

	// User
	r.echo.GET("/users", r.UserController.GetAllUsers)
	r.echo.GET("/users/:id", r.UserController.GetUserByID)

	// Work
	optionalConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(schema.JWTCustomClaims)
		},
		SigningKey: []byte(config.TOKEN_SECRET),
		Skipper: func(c echo.Context) bool {
			authHeader := c.Request().Header.Get("Authorization")
			return authHeader == ""
		},
	}
	o := r.echo.Group("/works", echojwt.WithConfig(optionalConfig))

	o.GET("", r.WorkController.GetAllWorks)
	o.GET("/users/:user_id", r.WorkController.GetWorksByUserID)

	r.echo.GET("/works/:work_id", r.WorkController.GetWorkByID)

	// Comment
	r.echo.GET("/works/:work_id/comments", r.CommentController.GetCommentsByWorkID)
	r.echo.POST("/works/:work_id/comments", r.CommentController.CreateComment)

	// Favorite
	r.echo.GET("/works/:work_id/favorite", r.FavoriteController.CountFavoritesByWorkID)

	// Tag (認証不要 - 一覧取得)
	r.echo.GET("/tags", r.TagController.GetAllTags)

	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(schema.JWTCustomClaims)
		},
		SigningKey: []byte(config.TOKEN_SECRET),
	}
	e := r.echo.Group("/auth", echojwt.WithConfig(config))
	e.Use(echojwt.WithConfig(config))

	// User
	e.PUT("/users", r.UserController.UpdateUser)
	e.GET("/users/me", r.UserController.GetIconAndURLByUserID)

	// Work
	e.POST("/works", r.WorkController.CreateWork)
	e.PATCH("/works/:work_id", r.WorkController.UpdateWork)
	e.DELETE("/works/:work_id", r.WorkController.DeleteWork)

	// Asset
	e.POST("/works/asset", r.AssetController.UploadAsset)

	// Favorite
	e.GET("/works/:work_id/favorite/is-favorite", r.FavoriteController.IsFavorite)
	e.POST("/works/:work_id/favorite", r.FavoriteController.CreateFavorite)
	e.DELETE("/works/:work_id/favorite", r.FavoriteController.DeleteFavorite)

	// Tag (認証必要 - 新規作成)
	e.POST("/tags", r.TagController.CreateTag)

	return r.echo
}
