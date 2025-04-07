package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/emper0r/InfiniteImages/config"
	"github.com/emper0r/InfiniteImages/internal/storage"
	"github.com/gin-gonic/gin"
)

// ImageResponse 表示单个图片的响应
type ImageResponse struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Size         int64     `json:"size"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Format       string    `json:"format"`
	Orientation  string    `json:"orientation"`
	Tags         []string  `json:"tags"`
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt,omitempty"`
	HasExpiry    bool      `json:"hasExpiry"`
}

// ListImagesResponse 表示图片列表响应
type ListImagesResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	Data    []ImageResponse `json:"data"`
}

// DeleteImageResponse 表示删除图片响应
type DeleteImageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

// ListImagesHandler 处理图片列表请求
func ListImagesHandler(cfg *config.Config, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		tag := c.Query("tag")

		// 验证分页参数
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}

		// 获取图片列表
		images, err := store.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "获取图片列表失败",
				"error":   err.Error(),
			})
			return
		}

		// 过滤标签（如果指定）
		var filteredImages []storage.ImageInfo
		if tag != "" {
			for _, img := range images {
				for _, t := range img.Tags {
					if t == tag {
						filteredImages = append(filteredImages, img)
						break
					}
				}
			}
		} else {
			filteredImages = images
		}

		// 计算分页
		total := len(filteredImages)
		start := (page - 1) * limit
		end := start + limit
		if start >= total {
			start = 0
			page = 1
		}
		if end > total {
			end = total
		}

		// 提取当前页的图片
		var pageImages []storage.ImageInfo
		if start < total {
			pageImages = filteredImages[start:end]
		}

		// 构建响应
		var responseData []ImageResponse
		for _, img := range pageImages {
			// 获取图片URL
			url := store.GetURL(img.ID, storage.Original, storage.ImageOrientation(img.Orientation))
			thumbnailURL := store.GetURL(img.ID, storage.WebP, storage.ImageOrientation(img.Orientation))

			responseData = append(responseData, ImageResponse{
				ID:           img.ID,
				Filename:     img.Filename,
				URL:          url,
				ThumbnailURL: thumbnailURL,
				Size:         img.Size,
				Width:        img.Width,
				Height:       img.Height,
				Format:       img.Format,
				Orientation:  img.Orientation,
				Tags:         img.Tags,
				CreatedAt:    img.CreatedAt,
				ExpiresAt:    img.ExpiresAt,
				HasExpiry:    img.HasExpiry,
			})
		}

		c.JSON(http.StatusOK, ListImagesResponse{
			Success: true,
			Message: "获取图片列表成功",
			Total:   total,
			Page:    page,
			Limit:   limit,
			Data:    responseData,
		})
	}
}

// GetImageHandler 处理获取单个图片信息请求
func GetImageHandler(cfg *config.Config, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取图片ID
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "缺少图片ID",
			})
			return
		}

		// 获取图片信息
		img, err := store.GetInfo(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "图片不存在",
				"error":   err.Error(),
			})
			return
		}

		// 获取图片URL
		url := store.GetURL(img.ID, storage.Original, storage.ImageOrientation(img.Orientation))
		thumbnailURL := store.GetURL(img.ID, storage.WebP, storage.ImageOrientation(img.Orientation))

		// 构建响应
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "获取图片信息成功",
			"data": ImageResponse{
				ID:           img.ID,
				Filename:     img.Filename,
				URL:          url,
				ThumbnailURL: thumbnailURL,
				Size:         img.Size,
				Width:        img.Width,
				Height:       img.Height,
				Format:       img.Format,
				Orientation:  img.Orientation,
				Tags:         img.Tags,
				CreatedAt:    img.CreatedAt,
				ExpiresAt:    img.ExpiresAt,
				HasExpiry:    img.HasExpiry,
			},
		})
	}
}

// DeleteImageHandler 处理删除图片请求
func DeleteImageHandler(cfg *config.Config, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取图片ID
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, DeleteImageResponse{
				Success: false,
				Message: "缺少图片ID",
			})
			return
		}

		// 获取图片信息
		img, err := store.GetInfo(id)
		if err != nil {
			c.JSON(http.StatusNotFound, DeleteImageResponse{
				Success: false,
				Message: "图片不存在",
			})
			return
		}

		// 删除所有格式的图片
		orientation := storage.ImageOrientation(img.Orientation)
		if err := store.Delete(id, storage.Original, orientation); err != nil {
			c.JSON(http.StatusInternalServerError, DeleteImageResponse{
				Success: false,
				Message: "删除原始图片失败: " + err.Error(),
			})
			return
		}

		// 尝试删除 WebP 格式（如果存在）
		_ = store.Delete(id, storage.WebP, orientation)

		// 尝试删除 AVIF 格式（如果存在）
		_ = store.Delete(id, storage.AVIF, orientation)

		// 删除图片信息
		if err := store.DeleteInfo(id); err != nil {
			c.JSON(http.StatusInternalServerError, DeleteImageResponse{
				Success: false,
				Message: "删除图片信息失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, DeleteImageResponse{
			Success: true,
			Message: "删除图片成功",
			ID:      id,
		})
	}
}

// UpdateTagsHandler 处理更新图片标签请求
func UpdateTagsHandler(cfg *config.Config, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取图片ID
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "缺少图片ID",
			})
			return
		}

		// 解析请求体
		var req struct {
			Tags []string `json:"tags"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的请求体",
				"error":   err.Error(),
			})
			return
		}

		// 获取图片信息
		img, err := store.GetInfo(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "图片不存在",
				"error":   err.Error(),
			})
			return
		}

		// 更新标签
		img.Tags = req.Tags

		// 保存图片信息
		if err := store.SaveInfo(img); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "更新标签失败",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "更新标签成功",
			"data": gin.H{
				"id":   id,
				"tags": req.Tags,
			},
		})
	}
}
