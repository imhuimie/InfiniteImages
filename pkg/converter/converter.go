package converter

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
)

// ImageFormat 表示图片格式
type ImageFormat string

const (
	// JPEG 格式
	JPEG ImageFormat = "jpeg"
	// PNG 格式
	PNG ImageFormat = "png"
	// GIF 格式
	GIF ImageFormat = "gif"
	// WebP 格式
	WebP ImageFormat = "webp"
	// AVIF 格式
	AVIF ImageFormat = "avif"
)

// ImageInfo 存储图片的基本信息
type ImageInfo struct {
	Width       int         // 图片宽度
	Height      int         // 图片高度
	Format      ImageFormat // 图片格式
	Orientation string      // 图片方向（横向/纵向）
}

// ConvertOptions 定义图片转换选项
type ConvertOptions struct {
	Format           ImageFormat // 目标格式
	Quality          int         // 图片质量（1-100）
	CompressionLevel int         // 压缩级别（1-10）
	Lossless         bool        // 是否无损压缩
	Width            int         // 目标宽度（0表示保持原始宽度）
	Height           int         // 目标高度（0表示保持原始高度）
}

// DefaultOptions 返回默认的转换选项
func DefaultOptions() ConvertOptions {
	return ConvertOptions{
		Format:           WebP,
		Quality:          80,
		CompressionLevel: 6,
		Lossless:         false,
		Width:            0,
		Height:           0,
	}
}

// DetectFormat 检测图片格式
func DetectFormat(reader io.Reader) (ImageFormat, image.Image, error) {
	// 读取图片数据
	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)

	// 解码图片
	img, format, err := image.Decode(tee)
	if err != nil {
		return "", nil, fmt.Errorf("无法解码图片: %w", err)
	}

	// 根据格式返回对应的 ImageFormat
	switch format {
	case "jpeg":
		return JPEG, img, nil
	case "png":
		return PNG, img, nil
	case "gif":
		return GIF, img, nil
	case "webp":
		return WebP, img, nil
	default:
		return "", nil, fmt.Errorf("不支持的图片格式: %s", format)
	}
}

// GetImageInfo 获取图片信息
func GetImageInfo(img image.Image, format ImageFormat) ImageInfo {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 确定图片方向
	orientation := "landscape"
	if height > width {
		orientation = "portrait"
	}

	return ImageInfo{
		Width:       width,
		Height:      height,
		Format:      format,
		Orientation: orientation,
	}
}

// Convert 将图片转换为指定格式
func Convert(img image.Image, options ConvertOptions) ([]byte, error) {
	// 调整图片大小（如果需要）
	if options.Width > 0 && options.Height > 0 {
		img = resize(img, options.Width, options.Height)
	}

	// 根据目标格式进行转换
	var buf bytes.Buffer
	var err error

	switch options.Format {
	case JPEG:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: options.Quality})
	case PNG:
		err = png.Encode(&buf, img)
	case GIF:
		err = gif.Encode(&buf, img, nil)
	case WebP:
		err = webp.Encode(&buf, img, &webp.Options{
			Lossless: options.Lossless,
			Quality:  float32(options.Quality),
		})
	case AVIF:
		// AVIF 需要使用外部命令行工具
		return convertToAVIF(img, options)
	default:
		return nil, fmt.Errorf("不支持的目标格式: %s", options.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("转换图片失败: %w", err)
	}

	return buf.Bytes(), nil
}

// ConvertFromReader 从 Reader 读取图片并转换为指定格式
func ConvertFromReader(reader io.Reader, options ConvertOptions) ([]byte, ImageInfo, error) {
	// 检测图片格式
	format, img, err := DetectFormat(reader)
	if err != nil {
		return nil, ImageInfo{}, err
	}

	// 获取图片信息
	info := GetImageInfo(img, format)

	// 转换图片
	data, err := Convert(img, options)
	if err != nil {
		return nil, info, err
	}

	// 更新图片信息中的格式
	info.Format = options.Format

	return data, info, nil
}

// 调整图片大小
func resize(img image.Image, width, height int) image.Image {
	// 这里简单实现，实际项目中应该使用更高质量的调整算法
	// 例如使用 github.com/nfnt/resize 或 github.com/disintegration/imaging
	// 这里仅作为示例
	return img
}

// 将图片转换为 AVIF 格式（使用外部命令行工具）
func convertToAVIF(img image.Image, options ConvertOptions) ([]byte, error) {
	// 创建临时文件
	tmpDir, err := os.MkdirTemp("", "infiniteimages")
	if err != nil {
		return nil, fmt.Errorf("无法创建临时目录: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 将图片保存为 PNG
	pngPath := filepath.Join(tmpDir, "input.png")
	pngFile, err := os.Create(pngPath)
	if err != nil {
		return nil, fmt.Errorf("无法创建临时 PNG 文件: %w", err)
	}

	if err := png.Encode(pngFile, img); err != nil {
		pngFile.Close()
		return nil, fmt.Errorf("无法编码 PNG 文件: %w", err)
	}
	pngFile.Close()

	// 输出 AVIF 路径
	avifPath := filepath.Join(tmpDir, "output.avif")

	// 构建命令
	args := []string{
		pngPath,
		"-o", avifPath,
		"-s", fmt.Sprintf("%d", options.CompressionLevel),
	}

	if options.Lossless {
		args = append(args, "--lossless")
	} else {
		args = append(args, "-q", fmt.Sprintf("%d", options.Quality))
	}

	// 执行命令
	cmd := exec.Command("avifenc", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("avifenc 命令执行失败: %w, stderr: %s", err, stderr.String())
	}

	// 读取生成的 AVIF 文件
	avifData, err := os.ReadFile(avifPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取生成的 AVIF 文件: %w", err)
	}

	return avifData, nil
}

// FormatFromExtension 从文件扩展名获取图片格式
func FormatFromExtension(filename string) ImageFormat {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return JPEG
	case ".png":
		return PNG
	case ".gif":
		return GIF
	case ".webp":
		return WebP
	case ".avif":
		return AVIF
	default:
		return ""
	}
}

// ExtensionFromFormat 从图片格式获取文件扩展名
func ExtensionFromFormat(format ImageFormat) string {
	switch format {
	case JPEG:
		return ".jpg"
	case PNG:
		return ".png"
	case GIF:
		return ".gif"
	case WebP:
		return ".webp"
	case AVIF:
		return ".avif"
	default:
		return ""
	}
}
