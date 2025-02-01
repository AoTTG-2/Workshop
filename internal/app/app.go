package app

import (
	"context"
	"fmt"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger"
	"net/http"
	"workshop/internal/config"
	"workshop/internal/controller/echo"
	"workshop/internal/repository"
	"workshop/internal/repository/entity"
	repoProvider "workshop/internal/repository/provider"
	"workshop/internal/router"
	"workshop/internal/router/middleware"
	"workshop/internal/router/validator"
	"workshop/internal/service/workshop"
	"workshop/internal/service/workshop/limiter"
	"workshop/pkg/appender"
	"workshop/pkg/urlwlv"
)

type Router interface {
	Run(port uint16) error
	Stop(ctx context.Context) error
}

type App struct {
	cfg                     *config.Config
	r                       Router
	redisClient             *redis.Client
	repo                    repository.Repository
	ws                      *workshop.Workshop
	limiter                 workshop.Limiter
	imageURLValidator       *urlwlv.Validator
	textURLValidator        *urlwlv.Validator
	assetBundleURLValidator *urlwlv.Validator
}

func New() (*App, error) {
	log.Info().Msgf("Initializing application")
	a := new(App)

	log.Info().Msgf("Initializing configuration")
	{
		cfg, err := config.New()
		if err != nil {
			return nil, err
		}
		a.cfg = cfg
		log.Info().Msgf("Configuration: %+v", a.cfg)
	}

	log.Info().Msgf("Initializing logger")
	{
		if err := a.buildLogger(); err != nil {
			return nil, err
		}
	}

	log.Info().Msgf("Initializing redis client")
	{
		a.redisClient = redis.NewClient(&redis.Options{
			Addr:     a.cfg.RedisConfig.Host,
			DB:       a.cfg.RedisConfig.DB,
			Username: a.cfg.RedisConfig.Username,
			Password: a.cfg.RedisConfig.Password,
		})
		if err := a.redisClient.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("failed to ping redis: %w", err)
		}
	}

	log.Info().Msgf("Initializing repository")
	{
		p := repoProvider.NewRepositoryProvider()

		repo, err := p.GetPostgresRepository(&repoProvider.PostgresConfiguration{
			Host:           a.cfg.Postgres.Host,
			Database:       a.cfg.Postgres.Database,
			Username:       a.cfg.Postgres.Username,
			Password:       a.cfg.Postgres.Password,
			MigrationsPath: a.cfg.Postgres.MigrationsPath,
			Params:         a.cfg.Postgres.Params,
		})
		if err != nil {
			return nil, err
		}

		if a.cfg.Postgres.DoMigrate {
			if err := repo.Migrate(context.Background()); err != nil {
				return nil, err
			}
		}

		a.repo = repo
	}

	log.Info().Msgf("Initializing URL validators")
	{
		cfgMap := appender.NewMapAppender(0, func(v *entity.URLValidatorConfig) string {
			return v.Type
		})
		if err := a.repo.GetAllURLValidatorConfigs(context.Background(), cfgMap); err != nil {
			return nil, fmt.Errorf("failed to get URL validator configs from repo: %w", err)
		}

		imageCfg, ok := cfgMap.Map()["image"]
		if !ok {
			return nil, fmt.Errorf("failed to find image URL validator config")
		}

		textCfg, ok := cfgMap.Map()["text"]
		if !ok {
			return nil, fmt.Errorf("failed to find text URL validator config")
		}

		assetBundleCfg, ok := cfgMap.Map()["asset_bundle"]
		if !ok {
			return nil, fmt.Errorf("failed to find asset bundle URL validator config")
		}

		a.imageURLValidator = urlwlv.NewValidator(imageCfg.Protocols, imageCfg.Domains, imageCfg.Extensions)
		a.textURLValidator = urlwlv.NewValidator(textCfg.Protocols, textCfg.Domains, textCfg.Extensions)
		a.assetBundleURLValidator = urlwlv.NewValidator(assetBundleCfg.Protocols, assetBundleCfg.Domains, assetBundleCfg.Extensions)
	}

	log.Info().Msg("Initializing rate limiter")
	{
		a.limiter = limiter.NewRedisLimiter(a.redisClient, "ws_limiter")
	}

	log.Info().Msgf("Initializing workshop service")
	{
		ws, err := workshop.New(
			&workshop.Config{
				PostsLimit: workshop.LimitConfig{
					Limit:  a.cfg.RateLimits.PostsLimit,
					Period: a.cfg.RateLimits.PostsPeriod,
				},
				CommentsLimit: workshop.LimitConfig{
					Limit:  a.cfg.RateLimits.CommentsLimit,
					Period: a.cfg.RateLimits.CommentsPeriod,
				},
			},
			a.repo,
			a.limiter,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create workshop service: %w", err)
		}
		a.ws = ws
	}

	log.Info().Msgf("Initializing router")
	{
		cfg := a.cfg.Router
		r := router.NewEchoRouter(cfg.Debug).
			WithDefaultNotFoundHandler().
			WithValidator(validator.NewEchoValidator())
		a.r = r

		if a.cfg.App.Local {
			r.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
				Skipper:                                  echoMiddleware.DefaultSkipper,
				AllowOrigins:                             a.cfg.App.LocalOrigins,
				AllowMethods:                             []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
				AllowCredentials:                         true,
				UnsafeWildcardOriginWithAllowCredentials: true,
			}))
		}

		r.Use(middleware.NewSessionsMiddleware(nil, middleware.SessionsMiddlewareConfig{
			QueryParam:       "token",
			Header:           "Authorization",
			ExtractionMethod: middleware.ExtractionMethodBoth,
			Debug:            cfg.AuthDebug,
		}))

		log.Info().Msgf("Registering routes")
		{
			r.GET("/swagger/*", echoSwagger.WrapHandler)
			api := r.Group("/api")
			{
				postHandler := echo.NewPostHandler(
					a.ws,
					a.imageURLValidator,
					a.textURLValidator,
					a.assetBundleURLValidator,
				)

				posts := api.Group("/posts")
				{
					posts.GET("", postHandler.GetList)
					posts.POST("", postHandler.Create)

					post := posts.Group("/:postID")
					{
						post.GET("", postHandler.GetOne)
						post.PUT("", postHandler.Update)
						post.DELETE("", postHandler.Delete)

						post.POST("/favorite", postHandler.FavoritePost)
						post.DELETE("/favorite", postHandler.UnfavoritePost)

						post.POST("/rate", postHandler.RatePost)

						post.POST("/moderate", postHandler.ModeratePost)
					}

					post.GET("/moderation-actions", postHandler.GetModerationActions)

					comments := api.Group("/comments")
					{
						comments.GET("", postHandler.GetComments)
						comments.POST("", postHandler.AddComment)
						comments.PUT("/:commentID", postHandler.UpdateComment)
						comments.DELETE("/:commentID", postHandler.DeleteComment)
					}
				}
			}
		}
	}

	return a, nil
}

func (a *App) Run() error {
	return a.r.Run(a.cfg.Router.ServerPort)
}

func (a *App) Stop(ctx context.Context) error {
	return a.r.Stop(ctx)
}

func (a *App) buildLogger() error {
	cfg := a.cfg.Logger
	zerolog.SetGlobalLevel(zerolog.Level(cfg.LogLevel))
	return nil
}
