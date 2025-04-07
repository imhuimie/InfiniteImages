package storage

import (
	"errors"
	"io"
	"time"
)

// ImageInfo 存储图片的元数据信息
type ImageInfo struct {
	ID          string    `json:"id"`          // 图片唯一标识符
	Filename    string    `json:"filename"`    // 原始文件名
	StoragePath string    `json:"storagePath"` // 存储路径
	Size        int64     `json:"size"`        // 文件大小（字节）
	Width       int       `json:"width"`       // 图片宽度
	Height      int       `json:"height"`      // 图片高度
	Format      string    `json:"format"`      // 图片格式
	Orientation string    `json:"orientation"` // 图片方向（横向/纵向）
	Tags        []string  `json:"tags"`        // 标签
	CreatedAt   time.Time `json:"createdAt"`   // 创建时间
	ExpiresAt   time.Time `json:"expiresAt"`   // 过期时间（如果有）
	HasExpiry   bool      `json:"hasExpiry"`   // 是否有过期时间
}

// ImageFormat 表示图片格式
type ImageFormat string

const (
	// Original 原始格式
	Original ImageFormat = "original"
	// WebP 格式
	WebP ImageFormat = "webp"
	// AVIF 格式
	AVIF ImageFormat = "avif"
)

// ImageOrientation 表示图片方向
type ImageOrientation string

const (
	// Landscape 横向
	Landscape ImageOrientation = "landscape"
	// Portrait 纵向
	Portrait ImageOrientation = "portrait"
)

// Storage 定义存储接口
type Storage interface {
	// Save 保存图片
	Save(reader io.Reader, filename string, format ImageFormat, orientation ImageOrientation) (string, error)

	// Delete 删除图片
	Delete(id string, format ImageFormat, orientation ImageOrientation) error

	// Get 获取图片
	Get(id string, format ImageFormat, orientation ImageOrientation) (io.ReadCloser, error)

	// GetURL 获取图片URL
	GetURL(id string, format ImageFormat, orientation ImageOrientation) string

	// List 列出所有图片
	List() ([]ImageInfo, error)

	// GetInfo 获取图片信息
	GetInfo(id string) (*ImageInfo, error)

	// SaveInfo 保存图片信息
	SaveInfo(info *ImageInfo) error

	// DeleteInfo 删除图片信息
	DeleteInfo(id string) error

	// CleanExpired 清理过期图片
	CleanExpired() (int, error)
}

// StorageFactory 创建存储实例的工厂函数类型
type StorageFactory func() (Storage, error)

// 存储类型注册表
var storageFactories = make(map[string]StorageFactory)

// RegisterStorage 注册存储类型
func RegisterStorage(name string, factory StorageFactory) {
	storageFactories[name] = factory
}

// 错误定义
var (
	// ErrUnsupportedStorageType 表示不支持的存储类型
	ErrUnsupportedStorageType = errors.New("不支持的存储类型")
	// ErrFileNotFound 表示文件未找到
	ErrFileNotFound = errors.New("文件未找到")
	// ErrInvalidImageFormat 表示无效的图片格式
	ErrInvalidImageFormat = errors.New("无效的图片格式")
	// ErrInvalidImageOrientation 表示无效的图片方向
	ErrInvalidImageOrientation = errors.New("无效的图片方向")
)

// NewStorage 创建指定类型的存储实例
func NewStorage(storageType string) (Storage, error) {
	factory, exists := storageFactories[storageType]
	if !exists {
		return nil, ErrUnsupportedStorageType
	}
	return factory()
}
