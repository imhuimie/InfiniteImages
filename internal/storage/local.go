package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage 实现本地文件系统存储
type LocalStorage struct {
	basePath string // 基础存储路径
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(basePath string) (Storage, error) {
	// 确保存储目录存在
	if err := ensureDirectories(basePath); err != nil {
		return nil, fmt.Errorf("无法创建存储目录: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// 确保所有必要的目录都存在
func ensureDirectories(basePath string) error {
	// 创建基础目录
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return err
	}

	// 创建原始图片目录
	originalDir := filepath.Join(basePath, "original")
	if err := os.MkdirAll(filepath.Join(originalDir, string(Landscape)), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(originalDir, string(Portrait)), 0755); err != nil {
		return err
	}

	// 创建 WebP 格式目录
	webpDir := filepath.Join(basePath, string(WebP))
	if err := os.MkdirAll(filepath.Join(webpDir, string(Landscape)), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(webpDir, string(Portrait)), 0755); err != nil {
		return err
	}

	// 创建 AVIF 格式目录
	avifDir := filepath.Join(basePath, string(AVIF))
	if err := os.MkdirAll(filepath.Join(avifDir, string(Landscape)), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(avifDir, string(Portrait)), 0755); err != nil {
		return err
	}

	// 创建元数据目录
	if err := os.MkdirAll(filepath.Join(basePath, "metadata"), 0755); err != nil {
		return err
	}

	return nil
}

// Save 保存图片到本地文件系统
func (s *LocalStorage) Save(reader io.Reader, filename string, format ImageFormat, orientation ImageOrientation) (string, error) {
	// 生成唯一ID
	id := generateID()

	// 构建存储路径
	var dirPath string
	if format == Original {
		dirPath = filepath.Join(s.basePath, "original", string(orientation))
	} else {
		dirPath = filepath.Join(s.basePath, string(format), string(orientation))
	}

	// 确保目录存在
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("无法创建目录 %s: %w", dirPath, err)
	}

	// 构建文件路径
	ext := filepath.Ext(filename)
	if format != Original {
		ext = "." + strings.ToLower(string(format))
	}
	filePath := filepath.Join(dirPath, id+ext)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("无法创建文件 %s: %w", filePath, err)
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, reader)
	if err != nil {
		return "", fmt.Errorf("无法写入文件 %s: %w", filePath, err)
	}

	return id, nil
}

// Delete 从本地文件系统删除图片
func (s *LocalStorage) Delete(id string, format ImageFormat, orientation ImageOrientation) error {
	// 构建文件路径
	var filePath string
	if format == Original {
		// 查找原始文件的扩展名
		dirPath := filepath.Join(s.basePath, "original", string(orientation))
		matches, err := filepath.Glob(filepath.Join(dirPath, id+".*"))
		if err != nil {
			return fmt.Errorf("无法查找文件: %w", err)
		}
		if len(matches) == 0 {
			return ErrFileNotFound
		}
		filePath = matches[0]
	} else {
		ext := "." + strings.ToLower(string(format))
		filePath = filepath.Join(s.basePath, string(format), string(orientation), id+ext)
	}

	// 删除文件
	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("无法删除文件 %s: %w", filePath, err)
	}

	return nil
}

// Get 从本地文件系统获取图片
func (s *LocalStorage) Get(id string, format ImageFormat, orientation ImageOrientation) (io.ReadCloser, error) {
	// 构建文件路径
	var filePath string
	if format == Original {
		// 查找原始文件的扩展名
		dirPath := filepath.Join(s.basePath, "original", string(orientation))
		matches, err := filepath.Glob(filepath.Join(dirPath, id+".*"))
		if err != nil {
			return nil, fmt.Errorf("无法查找文件: %w", err)
		}
		if len(matches) == 0 {
			return nil, ErrFileNotFound
		}
		filePath = matches[0]
	} else {
		ext := "." + strings.ToLower(string(format))
		filePath = filepath.Join(s.basePath, string(format), string(orientation), id+ext)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("无法打开文件 %s: %w", filePath, err)
	}

	return file, nil
}

// GetURL 获取图片的URL
func (s *LocalStorage) GetURL(id string, format ImageFormat, orientation ImageOrientation) string {
	var path string
	if format == Original {
		path = fmt.Sprintf("/static/images/original/%s/%s", orientation, id)
	} else {
		path = fmt.Sprintf("/static/images/%s/%s/%s.%s", format, orientation, id, strings.ToLower(string(format)))
	}
	return path
}

// List 列出所有图片
func (s *LocalStorage) List() ([]ImageInfo, error) {
	metadataDir := filepath.Join(s.basePath, "metadata")
	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, fmt.Errorf("无法读取元数据目录: %w", err)
	}

	var images []ImageInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(file.Name(), ".json")
		info, err := s.GetInfo(id)
		if err != nil {
			continue
		}

		images = append(images, *info)
	}

	return images, nil
}

// GetInfo 获取图片信息
func (s *LocalStorage) GetInfo(id string) (*ImageInfo, error) {
	metadataPath := filepath.Join(s.basePath, "metadata", id+".json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("无法读取元数据文件: %w", err)
	}

	var info ImageInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("无法解析元数据: %w", err)
	}

	return &info, nil
}

// SaveInfo 保存图片信息
func (s *LocalStorage) SaveInfo(info *ImageInfo) error {
	metadataDir := filepath.Join(s.basePath, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("无法创建元数据目录: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("无法序列化元数据: %w", err)
	}

	metadataPath := filepath.Join(metadataDir, info.ID+".json")
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("无法写入元数据文件: %w", err)
	}

	return nil
}

// DeleteInfo 删除图片信息
func (s *LocalStorage) DeleteInfo(id string) error {
	metadataPath := filepath.Join(s.basePath, "metadata", id+".json")
	err := os.Remove(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("无法删除元数据文件: %w", err)
	}

	return nil
}

// CleanExpired 清理过期图片
func (s *LocalStorage) CleanExpired() (int, error) {
	metadataDir := filepath.Join(s.basePath, "metadata")
	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return 0, fmt.Errorf("无法读取元数据目录: %w", err)
	}

	count := 0
	now := time.Now()

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(file.Name(), ".json")
		info, err := s.GetInfo(id)
		if err != nil {
			continue
		}

		// 检查是否过期
		if info.HasExpiry && !info.ExpiresAt.IsZero() && info.ExpiresAt.Before(now) {
			// 删除所有格式的图片
			s.Delete(id, Original, ImageOrientation(info.Orientation))
			s.Delete(id, WebP, ImageOrientation(info.Orientation))
			s.Delete(id, AVIF, ImageOrientation(info.Orientation))

			// 删除元数据
			s.DeleteInfo(id)

			count++
		}
	}

	return count, nil
}

// 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// 注册本地存储
func init() {
	RegisterStorage("local", func() (Storage, error) {
		basePath := os.Getenv("LOCAL_STORAGE_PATH")
		if basePath == "" {
			basePath = "static/images"
		}
		return NewLocalStorage(basePath)
	})
}
