package middleware

import (
	"net/http"
	"strings"

	"github.com/emper0r/InfiniteImages/config"
	"github.com/gin-gonic/gin"
)

// APIKeyAuth 中间件用于验证 API 密钥
func APIKeyAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果 API 密钥未设置，则跳过验证
		if cfg.APIKey == "" {
			c.Next()
			return
		}

		// 从请求头获取 API 密钥
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// 如果请求头中没有 API 密钥，则尝试从查询参数获取
			apiKey = c.Query("api_key")
		}

		// 验证 API 密钥
		if apiKey != cfg.APIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的 API 密钥",
			})
			return
		}

		c.Next()
	}
}

// IPFilter 中间件用于 IP 黑白名单过滤
func IPFilter(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 检查白名单
		if len(cfg.IPWhitelist) > 0 {
			allowed := false
			for _, ip := range cfg.IPWhitelist {
				if ip == "*" || ip == clientIP || strings.HasPrefix(clientIP, ip+".") {
					allowed = true
					break
				}
			}

			if !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "IP 地址不在白名单中",
				})
				return
			}
		}

		// 检查黑名单
		if len(cfg.IPBlacklist) > 0 {
			for _, ip := range cfg.IPBlacklist {
				if ip == clientIP || strings.HasPrefix(clientIP, ip+".") {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"success": false,
						"message": "IP 地址在黑名单中",
					})
					return
				}
			}
		}

		c.Next()
	}
}

// CORS 中间件用于处理跨域请求
func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置允许的来源
		origin := c.GetHeader("Origin")
		if origin != "" {
			// 检查是否允许该来源
			allowed := false
			for _, allowedOrigin := range cfg.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}

			if !allowed {
				c.Header("Access-Control-Allow-Origin", cfg.AllowedOrigins[0])
			}
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// 设置允许的方法
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// 设置允许的头部
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")

		// 设置是否允许携带凭证
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
