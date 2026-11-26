package router

import (
	"net/http"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/simesaba80/toybox-back/internal/infrastructure/config"
	"github.com/simesaba80/toybox-back/internal/infrastructure/external/jwks"
	"github.com/simesaba80/toybox-back/internal/infrastructure/external/proxy"
	"github.com/simesaba80/toybox-back/internal/interface/controller"
	"github.com/simesaba80/toybox-back/internal/interface/schema"
	"github.com/simesaba80/toybox-back/pkg/echovalidator"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Authenticator が発行するJWTを格納するCookie名
const authCookieName = "toybox_token"

type Router struct {
	echo               *echo.Echo
	UserController     *controller.UserController
	WorkController     *controller.WorkController
	CommentController  *controller.CommentController
	AssetController    *controller.AssetController
	FavoriteController *controller.FavoriteController
	TagController      *controller.TagController
	KeySet             *jwks.KeySet
}

func NewRouter(e *echo.Echo, uc *controller.UserController, wc *controller.WorkController, cc *controller.CommentController, assetc *controller.AssetController, fc *controller.FavoriteController, tagc *controller.TagController, keySet *jwks.KeySet) *Router {
	return &Router{
		echo:               e,
		UserController:     uc,
		WorkController:     wc,
		CommentController:  cc,
		AssetController:    assetc,
		FavoriteController: fc,
		TagController:      tagc,
		KeySet:             keySet,
	}
}

func (r *Router) Setup() *echo.Echo {
	r.echo.Validator = echovalidator.NewValidator()
	r.echo.Use(middleware.Logger())
	r.echo.Use(middleware.Recover())
	r.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.FRONTEND_URL,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	r.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	r.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	legacyProxy, legacyProxyErr := proxy.NewLegacyToyBoxProxy(
		config.LEGACY_TOYBOX_BASE_URL,
		config.LEGACY_TOYBOX_PROXY_HOST,
		controller.WriteLegacyProxyError,
	)
	legacyProxyController := controller.NewLegacyProxyController(legacyProxy, legacyProxyErr)
	r.echo.GET("/api/v1/works", legacyProxyController.GetWorksV1)
	r.echo.GET("/api/v2/works", legacyProxyController.GetWorksV2)
	r.echo.GET("/api/v1/blogs", legacyProxyController.GetBlogs)
	r.echo.GET("/api/v1/blogs/:blog_id", legacyProxyController.GetBlog)

	// User
	r.echo.GET("/users", r.UserController.GetAllUsers)
	r.echo.GET("/users/:id", r.UserController.GetUserByID)

	// Work
	optionalConfig := echojwt.Config{
		ParseTokenFunc: r.parseToken,
		TokenLookup:    "cookie:" + authCookieName,
		Skipper: func(c echo.Context) bool {
			_, err := c.Cookie(authCookieName)
			return err != nil
		},
	}
	o := r.echo.Group("/works", echojwt.WithConfig(optionalConfig), r.UserController.ResolveCurrentUser)

	o.GET("", r.WorkController.GetAllWorks)
	o.GET("/users/:user_id", r.WorkController.GetWorksByUserID)
	o.GET("/:work_id", r.WorkController.GetWorkByID)

	// Comment
	o.GET("/:work_id/comments", r.CommentController.GetCommentsByWorkID)
	o.POST("/:work_id/comments", r.CommentController.CreateComment)

	// Favorite
	r.echo.GET("/works/:work_id/favorite", r.FavoriteController.CountFavoritesByWorkID)

	// Tag (認証任意 - 一覧取得)
	r.echo.GET("/tags", r.TagController.GetAllTags, echojwt.WithConfig(optionalConfig), r.UserController.ResolveCurrentUser)

	config := echojwt.Config{
		ParseTokenFunc: r.parseToken,
		TokenLookup:    "cookie:" + authCookieName,
	}
	e := r.echo.Group("/auth", echojwt.WithConfig(config), r.UserController.ResolveCurrentUser)

	// User
	e.PUT("/users/me", r.UserController.RegisterCurrentUser)
	e.PATCH("/users", r.UserController.UpdateUser, middleware.BodyLimit("5MiB"))
	e.GET("/users/me", r.UserController.GetCurrentUser)

	// Work
	e.POST("/works", r.WorkController.CreateWork)
	e.PATCH("/works/:work_id", r.WorkController.UpdateWork)
	e.DELETE("/works/:work_id", r.WorkController.DeleteWork)

	// Asset
	e.POST("/works/asset", r.AssetController.UploadAsset, middleware.BodyLimit("2GiB"))

	// Favorite
	e.GET("/works/:work_id/favorite/is-favorite", r.FavoriteController.IsFavorite)
	e.POST("/works/:work_id/favorite", r.FavoriteController.CreateFavorite)
	e.DELETE("/works/:work_id/favorite", r.FavoriteController.DeleteFavorite)

	// Tag (認証必要 - 新規作成)
	e.POST("/tags", r.TagController.CreateTag)

	return r.echo
}

func (r *Router) parseToken(c echo.Context, auth string) (interface{}, error) {
	return r.KeySet.Parse(auth, new(schema.JWTCustomClaims))
}
