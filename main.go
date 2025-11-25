package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"
	"zero-music/config"
	"zero-music/database"
	"zero-music/handlers"
	"zero-music/logger"
	"zero-music/middleware"
	"zero-music/repository"
	"zero-music/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Params 定义命令行参数
type Params struct {
	ConfigPath string
	LogFile    string
}

// parseFlags 解析命令行参数
func parseFlags() *Params {
	configPath := flag.String("config", "config.json", "指定配置文件的路径。")
	logFile := flag.String("log", "app.log", "指定日志文件的路径。")
	flag.Parse()

	return &Params{
		ConfigPath: *configPath,
		LogFile:    *logFile,
	}
}

// ProvideParams 提供命令行参数
func ProvideParams() *Params {
	return parseFlags()
}

// ProvideConfig 提供配置实例
func ProvideConfig(params *Params) (*config.Config, error) {
	cfg, err := config.Load(params.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	return cfg, nil
}

// ProvideScanner 提供音乐扫描器实例
func ProvideScanner(cfg *config.Config) services.Scanner {
	return services.NewMusicScanner(
		cfg.Music.Directory,
		cfg.Music.SupportedFormats,
		cfg.Music.CacheTTLMinutes,
	)
}

// ProvideDBManager 提供数据库管理器实例
func ProvideDBManager(lc fx.Lifecycle, cfg *config.Config) (*database.DBManager, error) {
	dbCfg := &database.DBConfig{
		Driver: cfg.Database.Driver,
		DSN:    cfg.Database.Path,
	}

	provider := database.NewSQLiteProvider()
	dbManager := database.NewDBManager(provider, dbCfg)

	// 连接数据库
	if err := dbManager.Connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 注册生命周期钩子
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("正在关闭数据库连接...")
			if err := dbManager.Close(); err != nil {
				logger.Errorf("关闭数据库连接时出错: %v", err)
			}
			return nil
		},
	})

	return dbManager, nil
}

// ProvideDB 提供数据库连接实例
func ProvideDB(dbManager *database.DBManager) database.DB {
	return dbManager.GetDB()
}

// ProvideJWTManager 提供JWT管理器实例
func ProvideJWTManager(cfg *config.Config) *middleware.JWTManager {
	return middleware.NewJWTManager(cfg.Auth.JWTSecret)
}

// ProvideUserRepository 提供用户仓储实例
func ProvideUserRepository(db database.DB) repository.UserRepository {
	return repository.NewSQLiteUserRepository(db)
}

// ProvideFavoriteRepository 提供收藏仓储实例
func ProvideFavoriteRepository(db database.DB) repository.FavoriteRepository {
	return repository.NewSQLiteFavoriteRepository(db)
}

// ProvidePlayStatsRepository 提供播放统计仓储实例
func ProvidePlayStatsRepository(db database.DB) repository.PlayStatsRepository {
	return repository.NewSQLitePlayStatsRepository(db)
}

// ProvidePlaylistRepository 提供播放列表仓储实例
func ProvidePlaylistRepository(db database.DB) repository.PlaylistRepository {
	return repository.NewSQLitePlaylistRepository(db)
}

// ProvidePlaylistHandler 提供播放列表处理器
func ProvidePlaylistHandler(scanner services.Scanner) *handlers.PlaylistHandler {
	return handlers.NewPlaylistHandler(scanner)
}

// ProvideStreamHandler 提供流处理器
func ProvideStreamHandler(scanner services.Scanner, cfg *config.Config) *handlers.StreamHandler {
	return handlers.NewStreamHandler(scanner, cfg)
}

// ProvideSystemHandler 提供系统处理器
func ProvideSystemHandler(cfg *config.Config) *handlers.SystemHandler {
	return handlers.NewSystemHandler(cfg)
}

// ProvideAuthHandler 提供认证处理器
func ProvideAuthHandler(cfg *config.Config, userRepo repository.UserRepository, jwtManager *middleware.JWTManager) *handlers.AuthHandler {
	expiration := time.Duration(cfg.Auth.JWTExpireHours) * time.Hour
	return handlers.NewAuthHandler(expiration, userRepo, jwtManager)
}

// ProvideUserHandler 提供用户处理器
func ProvideUserHandler(
	scanner services.Scanner,
	favoriteRepo repository.FavoriteRepository,
	playStats repository.PlayStatsRepository,
	playlistRepo repository.PlaylistRepository,
) *handlers.UserHandler {
	return handlers.NewUserHandler(scanner, favoriteRepo, playStats, playlistRepo)
}

// ProvideSearchHandler 提供搜索处理器
func ProvideSearchHandler(scanner services.Scanner) *handlers.SearchHandler {
	return handlers.NewSearchHandler(scanner)
}

// ProvideRouter 提供 Gin 路由器
func ProvideRouter(
	cfg *config.Config,
	playlistHandler *handlers.PlaylistHandler,
	streamHandler *handlers.StreamHandler,
	systemHandler *handlers.SystemHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	searchHandler *handlers.SearchHandler,
	jwtManager *middleware.JWTManager,
) *gin.Engine {
	router := gin.Default()

	// 添加请求 ID 中间件
	router.Use(middleware.RequestID())

	// 健康检查端点
	router.GET("/health", systemHandler.HealthCheck)

	// API 根端点
	router.GET("/", systemHandler.APIIndex)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 认证路由（公开）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 播放列表路由（公开，可选认证）
		v1.GET("/songs", playlistHandler.GetAllSongs)
		v1.GET("/song/:id", playlistHandler.GetSongByID)

		// 音频流路由（公开，可选认证）
		v1.GET("/stream/:id", streamHandler.StreamAudio)

		// 搜索和浏览路由（公开）
		v1.GET("/search", searchHandler.Search)
		v1.GET("/artists", searchHandler.GetArtists)
		v1.GET("/artists/:name", searchHandler.GetArtistSongs)
		v1.GET("/albums", searchHandler.GetAlbums)
		v1.GET("/albums/:name", searchHandler.GetAlbumSongs)

		// 需要认证的用户路由
		user := v1.Group("/user")
		user.Use(middleware.JWTAuth(jwtManager))
		{
			// 用户信息
			user.GET("/profile", authHandler.GetProfile)
			user.PUT("/profile", authHandler.UpdateProfile)
			user.PUT("/password", authHandler.ChangePassword)
			user.POST("/refresh-token", authHandler.RefreshToken)

			// 收藏
			user.GET("/favorites", userHandler.GetFavorites)
			user.POST("/favorites/:id", userHandler.AddFavorite)
			user.DELETE("/favorites/:id", userHandler.RemoveFavorite)
			user.GET("/favorites/:id/check", userHandler.CheckFavorite)

			// 播放历史和统计
			user.POST("/play", userHandler.RecordPlay)
			user.GET("/history", userHandler.GetPlayHistory)
			user.GET("/stats", userHandler.GetUserStats)
			user.GET("/play-stats", userHandler.GetPlayStats)

			// 用户播放列表
			user.GET("/playlists", userHandler.GetPlaylists)
			user.POST("/playlists", userHandler.CreatePlaylist)
			user.GET("/playlists/:id", userHandler.GetPlaylist)
			user.PUT("/playlists/:id", userHandler.UpdatePlaylist)
			user.DELETE("/playlists/:id", userHandler.DeletePlaylist)
			user.POST("/playlists/:id/songs", userHandler.AddSongToPlaylist)
			user.DELETE("/playlists/:id/songs/:songId", userHandler.RemoveSongFromPlaylist)
			user.PUT("/playlists/:id/reorder", userHandler.ReorderPlaylistSongs)
		}
	}

	return router
}

// ProvideHTTPServer 提供 HTTP 服务器
func ProvideHTTPServer(cfg *config.Config, router *gin.Engine) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	readTimeout := time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second
	writeTimeout := time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second
	idleTimeout := time.Duration(cfg.Server.IdleTimeoutSeconds) * time.Second
	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// initLogger 初始化日志系统
func initLogger(lc fx.Lifecycle, params *Params) error {
	logCloser, err := logger.Init(params.LogFile)
	if err != nil {
		logger.Warnf("日志文件初始化警告: %v", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("zero music服务器正在启动...")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if logCloser != nil {
				logger.Info("正在关闭日志输出...")
				if err := logCloser.Close(); err != nil {
					logger.Errorf("关闭日志输出时出错: %v", err)
				}
			}
			return nil
		},
	})

	return nil
}

// startHTTPServer 启动 HTTP 服务器
func startHTTPServer(lc fx.Lifecycle, srv *http.Server, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Zero Music 服务器启动中...")
			logger.Infof("服务地址: http://localhost:%d", cfg.Server.Port)
			logger.Infof("音乐目录: %s", cfg.Music.Directory)

			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Errorf("服务器启动失败: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("正在关闭服务器...")
			shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeoutSeconds) * time.Second
			shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Errorf("服务器强制关闭: %v", err)
				return err
			}
			logger.Info("服务器已优雅关闭")
			return nil
		},
	})
}

func main() {
	app := fx.New(
		// 提供依赖
		fx.Provide(
			ProvideParams,
			ProvideConfig,
			ProvideDBManager,
			ProvideDB,
			ProvideScanner,
			ProvideJWTManager,
			// Repository 层
			ProvideUserRepository,
			ProvideFavoriteRepository,
			ProvidePlayStatsRepository,
			ProvidePlaylistRepository,
			// Handler 层
			ProvidePlaylistHandler,
			ProvideStreamHandler,
			ProvideSystemHandler,
			ProvideAuthHandler,
			ProvideUserHandler,
			ProvideSearchHandler,
			ProvideRouter,
			ProvideHTTPServer,
		),
		// 调用初始化函数
		fx.Invoke(
			initLogger,
			startHTTPServer,
		),
	)

	app.Run()
}
