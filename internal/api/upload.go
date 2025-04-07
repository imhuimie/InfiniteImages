package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emper0r/InfiniteImages/config"
	"github.com/emper0r/InfiniteImages/internal/storage"
	"github.com/emper0r/InfiniteImages/pkg/converter"
	"github.com/gin-gonic/gin"
)

// UploadResponse 表示上传响应
type UploadResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    []string `json:"data,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

// UploadHandler 处理图片上传
func UploadHandler(cfg *config.Config, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否是多文件上传
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, UploadResponse{
				Success: false,
				Message: "无法解析上传表单",
				Errors:  []string{err.Error()},
			})
			return
		}

		// 获取上传的文件
		files := form.File["images"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, UploadResponse{
				Success: false,
				Message: "未找到上传的图片",
			})
			return
		}

		// 检查文件数量是否超过限制
		if len(files) > cfg.MaxUploadCount {
			c.JSON(http.StatusBadRequest, UploadResponse{
				Success: false,
				Message: fmt.Sprintf("上传的图片数量超过限制（最大 %d 张）", cfg.MaxUploadCount),
			})
			return
		}

		// 处理每个文件
		var urls []string
		var errors []string

		for _, file := range files {
			// 检查文件大小
			if file.Size > cfg.MaxUploadSize {
				errors = append(errors, fmt.Sprintf("文件 %s 大小超过限制（最大 %d 字节）", file.Filename, cfg.MaxUploadSize))
				continue
			}

			// 检查文件类型
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if !isAllowedImageType(ext) {
				errors = append(errors, fmt.Sprintf("文件 %s 类型不支持", file.Filename))
				continue
			}

			// 打开文件
			src, err := file.Open()
			if err != nil {
				errors = append(errors, fmt.Sprintf("无法打开文件 %s: %v", file.Filename, err))
				continue
			}
			defer src.Close()

			// 创建临时文件
			tmpFile, err := os.CreateTemp("", "upload-*"+ext)
			if err != nil {
				errors = append(errors, fmt.Sprintf("无法创建临时文件: %v", err))
				continue
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			// 复制文件内容到临时文件
			if _, err = io.Copy(tmpFile, src); err != nil {
				errors = append(errors, fmt.Sprintf("无法复制文件内容: %v", err))
				continue
			}

			// 重置文件指针
			if _, err = tmpFile.Seek(0, 0); err != nil {
				errors = append(errors, fmt.Sprintf("无法重置文件指针: %v", err))
				continue
			}

			// 转换图片
			options := converter.DefaultOptions()
			options.Quality = cfg.ImageQuality
			options.CompressionLevel = cfg.CompressionEffort
			options.Lossless = cfg.ForceLossless

			data, info, err := converter.ConvertFromReader(tmpFile, options)
			if err != nil {
				errors = append(errors, fmt.Sprintf("无法转换图片 %s: %v", file.Filename, err))
				continue
			}

			// 保存原始图片
			originalReader, err := os.Open(tmpFile.Name())
			if err != nil {
				errors = append(errors, fmt.Sprintf("无法打开原始图片: %v", err))
				continue
			}
			defer originalReader.Close()

			// 保存原始图片
			id, err := store.Save(originalReader, file.Filename, storage.Original, storage.ImageOrientation(info.Orientation))
			if err != nil {
				errors = append(errors, fmt.Sprintf("无法保存原始图片: %v", err))
				continue
			}

			// 保存转换后的图片
			webpReader := strings.NewReader(string(data))
			_, err = store.Save(webpReader, file.Filename, storage.WebP, storage.ImageOrientation(info.Orientation))
			if err != nil {
				errors = append(errors, fmt.Sprintf("无法保存 WebP 图片: %v", err))
				continue
			}

			// 保存图片信息
			imageInfo := &storage.ImageInfo{
				ID:          id,
				Filename:    file.Filename,
				StoragePath: filepath.Join("original", info.Orientation),
				Size:        file.Size,
				Width:       info.Width,
				Height:      info.Height,
				Format:      string(info.Format),
				Orientation: info.Orientation,
				Tags:        []string{},
				CreatedAt:   time.Now(),
				ExpiresAt:   time.Time{},
				HasExpiry:   false,
			}

			if err := store.SaveInfo(imageInfo); err != nil {
				errors = append(errors, fmt.Sprintf("无法保存图片信息: %v", err))
				continue
			}

			// 获取图片URL
			url := store.GetURL(id, storage.Original, storage.ImageOrientation(info.Orientation))
			urls = append(urls, url)
		}

		// 返回响应
		if len(errors) == 0 {
			c.JSON(http.StatusOK, UploadResponse{
				Success: true,
				Message: fmt.Sprintf("成功上传 %d 张图片", len(urls)),
				Data:    urls,
			})
		} else if len(urls) > 0 {
			c.JSON(http.StatusOK, UploadResponse{
				Success: true,
				Message: fmt.Sprintf("部分图片上传成功（%d/%d）", len(urls), len(files)),
				Data:    urls,
				Errors:  errors,
			})
		} else {
			c.JSON(http.StatusBadRequest, UploadResponse{
				Success: false,
				Message: "所有图片上传失败",
				Errors:  errors,
			})
		}
	}
}

// 检查文件类型是否允许
func isAllowedImageType(ext string) bool {
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".avif": true,
	}
	return allowedTypes[ext]
}
