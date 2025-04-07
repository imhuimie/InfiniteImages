package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Config 存储应用程序配置
type Config struct {
	// 服务器配置
	ServerAddr string
	DebugMode  bool

	// API 密钥配置
	APIKey string

	// 存储配置
	StorageType      string
	LocalStoragePath string

	// S3 存储配置
	S3Endpoint   string
	S3Region     string
	S3AccessKey  string
	S3SecretKey  string
	S3Bucket     string
	CustomDomain string

	// FTP 存储配置
	FTPHost     string
	FTPPort     int
	FTPUsername string
	FTPPassword string
	FTPPath     string

	// 图像处理配置
	MaxUploadSize     int64
	MaxUploadCount    int
	ImageQuality      int
	WorkerThreads     int
	CompressionEffort int
	ForceLossless     bool

	// 安全配置
	AllowedOrigins []string
	JWTSecret      string
	IPWhitelist    []string
	IPBlacklist    []string

	// 功能开关
	EnableWatermark     bool
	EnableEXIFStrip     bool
	EnableAutoClean     bool
	EnableNSFWDetection bool

	// 水印配置
	WatermarkType      string
	WatermarkText      string
	WatermarkFont      string
	WatermarkSize      int
	WatermarkColor     string
	WatermarkOpacity   int
	WatermarkPosition  string
	WatermarkImagePath string

	// 缓存配置
	EnableCache bool
	CacheTTL    int

	// 日志配置
	LogLevel string
	LogFile  string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	cfg := &Config{
		// 服务器配置
		ServerAddr: getEnv("SERVER_ADDR", "0.0.0.0:8080"),
		DebugMode:  getEnvBool("DEBUG_MODE", false),

		// API 密钥配置
		APIKey: getEnv("API_KEY", ""),

		// 存储配置
		StorageType:      getEnv("STORAGE_TYPE", "local"),
		LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "static/images"),

		// S3 存储配置
		S3Endpoint:   getEnv("S3_ENDPOINT", ""),
		S3Region:     getEnv("S3_REGION", ""),
		S3AccessKey:  getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:  getEnv("S3_SECRET_KEY", ""),
		S3Bucket:     getEnv("S3_BUCKET", ""),
		CustomDomain: getEnv("CUSTOM_DOMAIN", ""),

		// FTP 存储配置
		FTPHost:     getEnv("FTP_HOST", ""),
		FTPPort:     getEnvInt("FTP_PORT", 21),
		FTPUsername: getEnv("FTP_USERNAME", ""),
		FTPPassword: getEnv("FTP_PASSWORD", ""),
		FTPPath:     getEnv("FTP_PATH", ""),

		// 图像处理配置
		MaxUploadSize:     getEnvInt64("MAX_UPLOAD_SIZE", 10*1024*1024), // 默认 10MB
		MaxUploadCount:    getEnvInt("MAX_UPLOAD_COUNT", 20),
		ImageQuality:      getEnvInt("IMAGE_QUALITY", 80),
		WorkerThreads:     getEnvInt("WORKER_THREADS", 4),
		CompressionEffort: getEnvInt("COMPRESSION_EFFORT", 6),
		ForceLossless:     getEnvBool("FORCE_LOSSLESS", false),

		// 安全配置
		AllowedOrigins: getEnvStringSlice("ALLOWED_ORIGINS", []string{"*"}),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		IPWhitelist:    getEnvStringSlice("IP_WHITELIST", []string{}),
		IPBlacklist:    getEnvStringSlice("IP_BLACKLIST", []string{}),

		// 功能开关
		EnableWatermark:     getEnvBool("ENABLE_WATERMARK", false),
		EnableEXIFStrip:     getEnvBool("ENABLE_EXIF_STRIP", true),
		EnableAutoClean:     getEnvBool("ENABLE_AUTO_CLEAN", true),
		EnableNSFWDetection: getEnvBool("ENABLE_NSFW_DETECTION", false),

		// 水印配置
		WatermarkType:      getEnv("WATERMARK_TYPE", "text"),
		WatermarkText:      getEnv("WATERMARK_TEXT", "InfiniteImages"),
		WatermarkFont:      getEnv("WATERMARK_FONT", "Arial"),
		WatermarkSize:      getEnvInt("WATERMARK_SIZE", 24),
		WatermarkColor:     getEnv("WATERMARK_COLOR", "#ffffff"),
		WatermarkOpacity:   getEnvInt("WATERMARK_OPACITY", 50),
		WatermarkPosition:  getEnv("WATERMARK_POSITION", "bottom-right"),
		WatermarkImagePath: getEnv("WATERMARK_IMAGE_PATH", ""),

		// 缓存配置
		EnableCache: getEnvBool("ENABLE_CACHE", true),
		CacheTTL:    getEnvInt("CACHE_TTL", 3600),

		// 日志配置
		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFile:  getEnv("LOG_FILE", "logs/app.log"),
	}

	// 验证配置
	cfg.validate()

	return cfg
}

// validate 验证配置的有效性
func (c *Config) validate() {
	// 验证存储类型
	if c.StorageType != "local" && c.StorageType != "s3" && c.StorageType != "ftp" {
		log.Printf("警告: 无效的存储类型 '%s'，使用默认值 'local'", c.StorageType)
		c.StorageType = "local"
	}

	// 验证 S3 配置
	if c.StorageType == "s3" {
		if c.S3Endpoint == "" || c.S3AccessKey == "" || c.S3SecretKey == "" || c.S3Bucket == "" {
			log.Fatal("错误: 使用 S3 存储时必须提供 S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY 和 S3_BUCKET")
		}
	}

	// 验证 FTP 配置
	if c.StorageType == "ftp" {
		if c.FTPHost == "" || c.FTPUsername == "" || c.FTPPassword == "" {
			log.Fatal("错误: 使用 FTP 存储时必须提供 FTP_HOST, FTP_USERNAME 和 FTP_PASSWORD")
		}
	}

	// 验证图像质量
	if c.ImageQuality < 1 || c.ImageQuality > 100 {
		log.Printf("警告: 无效的图像质量 %d，使用默认值 80", c.ImageQuality)
		c.ImageQuality = 80
	}

	// 验证压缩级别
	if c.CompressionEffort < 1 || c.CompressionEffort > 10 {
		log.Printf("警告: 无效的压缩级别 %d，使用默认值 6", c.CompressionEffort)
		c.CompressionEffort = 6
	}

	// 验证水印配置
	if c.EnableWatermark {
		if c.WatermarkType != "text" && c.WatermarkType != "image" {
			log.Printf("警告: 无效的水印类型 '%s'，使用默认值 'text'", c.WatermarkType)
			c.WatermarkType = "text"
		}

		if c.WatermarkType == "image" && c.WatermarkImagePath == "" {
			log.Printf("警告: 使用图片水印时必须提供 WATERMARK_IMAGE_PATH")
			c.EnableWatermark = false
		}
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool 获取布尔类型的环境变量，如果不存在则返回默认值
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("警告: 无法解析布尔值 '%s'，使用默认值 %v: %v", key, defaultValue, err)
		return defaultValue
	}
	return b
}

// getEnvInt 获取整数类型的环境变量，如果不存在则返回默认值
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("警告: 无法解析整数 '%s'，使用默认值 %d: %v", key, defaultValue, err)
		return defaultValue
	}
	return i
}

// getEnvInt64 获取 int64 类型的环境变量，如果不存在则返回默认值
func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Printf("警告: 无法解析 int64 '%s'，使用默认值 %d: %v", key, defaultValue, err)
		return defaultValue
	}
	return i
}

// getEnvStringSlice 获取字符串切片类型的环境变量，如果不存在则返回默认值
func getEnvStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}
