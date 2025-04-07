package middleware

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/emper0r/InfiniteImages/config"
	"github.com/gin-gonic/gin"
)

// responseWriter 是一个自定义的响应写入器，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写 Write 方法，同时写入原始响应和缓冲区
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 重写 WriteString 方法，同时写入原始响应和缓冲区
func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Logger 中间件用于记录请求和响应信息
func Logger(cfg *config.Config) gin.HandlerFunc {
	// 确保日志目录存在
	if cfg.LogFile != "" {
		logDir := filepath.Dir(cfg.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("无法创建日志目录 %s: %v\n", logDir, err)
		}
	}

	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建自定义响应写入器
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = w

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 获取响应状态码
		statusCode := c.Writer.Status()

		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 获取请求方法和路径
		method := c.Request.Method
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		// 获取用户代理
		userAgent := c.Request.UserAgent()

		// 构建日志消息
		logMessage := fmt.Sprintf("[%s] %s | %3d | %13v | %15s | %s | %s | %s\n",
			endTime.Format("2006/01/02 - 15:04:05"),
			method,
			statusCode,
			latency,
			clientIP,
			path,
			userAgent,
			c.Errors.String(),
		)

		// 根据日志级别记录详细信息
		if cfg.LogLevel == "debug" {
			// 在调试模式下记录请求和响应体
			if len(requestBody) > 0 {
				logMessage += fmt.Sprintf("Request Body: %s\n", string(requestBody))
			}

			if w.body.Len() > 0 {
				// 限制响应体的长度，避免日志过大
				responseBody := w.body.String()
				if len(responseBody) > 1024 {
					responseBody = responseBody[:1024] + "... (truncated)"
				}
				logMessage += fmt.Sprintf("Response Body: %s\n", responseBody)
			}
		}

		// 写入日志
		if cfg.LogFile != "" {
			// 写入文件
			f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				defer f.Close()
				f.WriteString(logMessage)
			} else {
				// 如果无法写入文件，则输出到控制台
				fmt.Print(logMessage)
			}
		} else {
			// 输出到控制台
			fmt.Print(logMessage)
		}
	}
}
