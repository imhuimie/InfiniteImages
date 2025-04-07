package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emper0r/InfiniteImages/config"
	"github.com/emper0r/InfiniteImages/internal/api"
	"github.com/emper0r/InfiniteImages/internal/middleware"
	"github.com/emper0r/InfiniteImages/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 解析命令行参数
	configFile := flag.String("config", ".env", "配置文件路径")
	flag.Parse()

	// 加载环境变量
	err := godotenv.Load(*configFile)
	if err != nil {
		log.Printf("警告: 无法加载 %s 文件: %v", *configFile, err)
	}

	// 加载配置
	cfg := config.LoadConfig()

	// 设置运行模式
	if !cfg.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建存储实例
	store, err := storage.NewStorage(cfg.StorageType)
	if err != nil {
		log.Fatalf("无法创建存储实例: %v", err)
	}

	// 创建 Gin 引擎
	r := gin.New()

	// 使用自定义中间件
	r.Use(middleware.Logger(cfg))
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.IPFilter(cfg))
	r.Use(gin.Recovery())

	// 设置信任代理
	r.SetTrustedProxies(nil)

	// 设置静态文件服务
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// 设置 API 路由
	setupRoutes(r, cfg, store)

	// 获取服务器地址
	addr := cfg.ServerAddr

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 优雅关闭服务器
	go func() {
		// 监听中断信号
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("正在关闭服务器...")

		// 设置关闭超时时间
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("服务器强制关闭:", err)
		}
		log.Println("服务器已关闭")
	}()

	// 启动服务器
	log.Printf("服务器正在运行于 %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// setupRoutes 设置 API 路由
func setupRoutes(r *gin.Engine, cfg *config.Config, store storage.Storage) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 公共 API 路由组
	publicAPI := r.Group("/api")
	{
		// 验证 API 密钥
		publicAPI.POST("/validate-api-key", func(c *gin.Context) {
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" {
				apiKey = c.PostForm("api_key")
			}

			if apiKey == cfg.APIKey {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "API 密钥有效",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "API 密钥无效",
				})
			}
		})

		// 随机图片
		publicAPI.GET("/random", func(c *gin.Context) {
			// 获取所有图片
			images, err := store.List()
			if err != nil || len(images) == 0 {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "没有可用的图片",
				})
				return
			}

			// 随机选择一张图片
			randomIndex := time.Now().UnixNano() % int64(len(images))
			img := images[randomIndex]

			// 获取图片URL
			url := store.GetURL(img.ID, storage.Original, storage.ImageOrientation(img.Orientation))

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "获取随机图片成功",
				"data": gin.H{
					"id":       img.ID,
					"url":      url,
					"filename": img.Filename,
				},
			})
		})
	}

	// 受保护的 API 路由组
	protectedAPI := r.Group("/api")
	protectedAPI.Use(middleware.APIKeyAuth(cfg))
	{
		// 上传图片
		protectedAPI.POST("/upload", api.UploadHandler(cfg, store))

		// 获取图片列表
		protectedAPI.GET("/images", api.ListImagesHandler(cfg, store))

		// 获取单个图片信息
		protectedAPI.GET("/images/:id", api.GetImageHandler(cfg, store))

		// 删除图片
		protectedAPI.DELETE("/images/:id", api.DeleteImageHandler(cfg, store))

		// 更新图片标签
		protectedAPI.PUT("/images/:id/tags", api.UpdateTagsHandler(cfg, store))

		// 获取系统配置
		protectedAPI.GET("/config", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "获取系统配置成功",
				"data": gin.H{
					"storageType":       cfg.StorageType,
					"maxUploadSize":     cfg.MaxUploadSize,
					"maxUploadCount":    cfg.MaxUploadCount,
					"enableWatermark":   cfg.EnableWatermark,
					"enableEXIFStrip":   cfg.EnableEXIFStrip,
					"enableAutoClean":   cfg.EnableAutoClean,
					"imageQuality":      cfg.ImageQuality,
					"compressionEffort": cfg.CompressionEffort,
				},
			})
		})

		// 手动触发清理过期图片
		protectedAPI.POST("/trigger-cleanup", func(c *gin.Context) {
			count, err := store.CleanExpired()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "清理过期图片失败",
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": fmt.Sprintf("成功清理 %d 张过期图片", count),
				"data": gin.H{
					"count": count,
				},
			})
		})

		// 获取所有标签
		protectedAPI.GET("/tags", func(c *gin.Context) {
			// 获取所有图片
			images, err := store.List()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "获取图片列表失败",
					"error":   err.Error(),
				})
				return
			}

			// 收集所有标签
			tagsMap := make(map[string]int)
			for _, img := range images {
				for _, tag := range img.Tags {
					tagsMap[tag]++
				}
			}

			// 构建标签列表
			var tags []gin.H
			for tag, count := range tagsMap {
				tags = append(tags, gin.H{
					"name":  tag,
					"count": count,
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "获取标签列表成功",
				"data":    tags,
			})
		})
	}

	// 前端路由
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "InfiniteImages 服务正在运行")
	})
}
