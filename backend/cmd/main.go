package main

import (
	"log"
	"vaultseed-backend/internal/database"
	"vaultseed-backend/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	r := gin.Default()

	// CORS 配置
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	r.Use(cors.New(config))

	// API 路由
	api := r.Group("/api")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.LoginHandler)
			auth.POST("/register-public-key", handlers.RegisterPublicKeyHandler)
			auth.GET("/nonce", handlers.GetNonceHandler)
		}

		// 内容相关
		content := api.Group("/content")
		{
			content.POST("/create", handlers.CreateContentHandler)
			content.GET("/list", handlers.ListContentHandler)
			content.POST("/decrypt", handlers.DecryptContentHandler)
			content.GET("/:id", handlers.GetContentDetailHandler)
		}

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	// 启动服务器
	log.Println("VaultSeed backend server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
