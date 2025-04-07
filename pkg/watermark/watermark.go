package watermark

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Position 表示水印位置
type Position string

const (
	// TopLeft 左上角
	TopLeft Position = "top-left"
	// TopRight 右上角
	TopRight Position = "top-right"
	// BottomLeft 左下角
	BottomLeft Position = "bottom-left"
	// BottomRight 右下角
	BottomRight Position = "bottom-right"
	// Center 中心
	Center Position = "center"
)

// WatermarkType 表示水印类型
type WatermarkType string

const (
	// Text 文字水印
	Text WatermarkType = "text"
	// Image 图片水印
	Image WatermarkType = "image"
)

// Options 水印选项
type Options struct {
	Type      WatermarkType // 水印类型
	Text      string        // 水印文字
	FontPath  string        // 字体路径
	FontSize  float64       // 字体大小
	Color     color.Color   // 水印颜色
	Opacity   uint8         // 水印透明度（0-255）
	Position  Position      // 水印位置
	ImagePath string        // 水印图片路径
	Margin    int           // 边距
}

// DefaultOptions 返回默认的水印选项
func DefaultOptions() Options {
	return Options{
		Type:     Text,
		Text:     "InfiniteImages",
		FontPath: "",
		FontSize: 24,
		Color:    color.RGBA{255, 255, 255, 255},
		Opacity:  128,
		Position: BottomRight,
		Margin:   10,
	}
}

// AddWatermark 在图片上添加水印
func AddWatermark(img image.Image, options Options) (image.Image, error) {
	// 创建一个新的 RGBA 图片
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// 根据水印类型添加水印
	switch options.Type {
	case Text:
		return addTextWatermark(rgba, options)
	case Image:
		return addImageWatermark(rgba, options)
	default:
		return nil, fmt.Errorf("不支持的水印类型: %s", options.Type)
	}
}

// AddWatermarkFromReader 从 Reader 读取图片并添加水印
func AddWatermarkFromReader(reader io.Reader, options Options) ([]byte, error) {
	// 解码图片
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("无法解码图片: %w", err)
	}

	// 添加水印
	watermarked, err := AddWatermark(img, options)
	if err != nil {
		return nil, err
	}

	// 编码为 PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, watermarked); err != nil {
		return nil, fmt.Errorf("无法编码图片: %w", err)
	}

	return buf.Bytes(), nil
}

// 添加文字水印
func addTextWatermark(img *image.RGBA, options Options) (image.Image, error) {
	// 加载字体
	var fontData []byte
	var err error

	if options.FontPath != "" {
		fontData, err = os.ReadFile(options.FontPath)
		if err != nil {
			return nil, fmt.Errorf("无法读取字体文件: %w", err)
		}
	} else {
		// 使用默认字体
		fontData = defaultFontData
	}

	f, err := freetype.ParseFont(fontData)
	if err != nil {
		return nil, fmt.Errorf("无法解析字体: %w", err)
	}

	// 创建绘图上下文
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(options.FontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(options.Color))
	c.SetHinting(font.HintingFull)

	// 计算文字大小
	metrics := truetype.NewFace(f, &truetype.Options{
		Size: options.FontSize,
		DPI:  72,
	}).Metrics()
	textHeight := metrics.Height.Ceil()
	textWidth := fixed.Int26_6(len(options.Text) * int(options.FontSize) * 64).Ceil()

	// 计算文字位置
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	var x, y int

	switch options.Position {
	case TopLeft:
		x = options.Margin
		y = options.Margin + textHeight
	case TopRight:
		x = width - textWidth - options.Margin
		y = options.Margin + textHeight
	case BottomLeft:
		x = options.Margin
		y = height - options.Margin
	case BottomRight:
		x = width - textWidth - options.Margin
		y = height - options.Margin
	case Center:
		x = (width - textWidth) / 2
		y = (height + textHeight) / 2
	default:
		return nil, fmt.Errorf("不支持的水印位置: %s", options.Position)
	}

	// 绘制文字
	pt := freetype.Pt(x, y)
	_, err = c.DrawString(options.Text, pt)
	if err != nil {
		return nil, fmt.Errorf("无法绘制文字: %w", err)
	}

	return img, nil
}

// 添加图片水印
func addImageWatermark(img *image.RGBA, options Options) (image.Image, error) {
	// 加载水印图片
	watermarkFile, err := os.Open(options.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开水印图片: %w", err)
	}
	defer watermarkFile.Close()

	watermarkImg, _, err := image.Decode(watermarkFile)
	if err != nil {
		return nil, fmt.Errorf("无法解码水印图片: %w", err)
	}

	// 计算水印位置
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	watermarkBounds := watermarkImg.Bounds()
	watermarkWidth := watermarkBounds.Dx()
	watermarkHeight := watermarkBounds.Dy()
	var x, y int

	switch options.Position {
	case TopLeft:
		x = options.Margin
		y = options.Margin
	case TopRight:
		x = width - watermarkWidth - options.Margin
		y = options.Margin
	case BottomLeft:
		x = options.Margin
		y = height - watermarkHeight - options.Margin
	case BottomRight:
		x = width - watermarkWidth - options.Margin
		y = height - watermarkHeight - options.Margin
	case Center:
		x = (width - watermarkWidth) / 2
		y = (height - watermarkHeight) / 2
	default:
		return nil, fmt.Errorf("不支持的水印位置: %s", options.Position)
	}

	// 创建水印图片的透明版本
	mask := image.NewRGBA(watermarkBounds)
	for py := 0; py < watermarkHeight; py++ {
		for px := 0; px < watermarkWidth; px++ {
			r, g, b, a := watermarkImg.At(px+watermarkBounds.Min.X, py+watermarkBounds.Min.Y).RGBA()
			mask.Set(px, py, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8((a >> 8) * uint32(options.Opacity) / 255),
			})
		}
	}

	// 绘制水印
	draw.Draw(img, image.Rect(x, y, x+watermarkWidth, y+watermarkHeight), mask, watermarkBounds.Min, draw.Over)

	return img, nil
}

// 默认字体数据（嵌入的简单字体）
// 实际项目中应该使用真实的字体文件
var defaultFontData = []byte{
	// 这里应该是字体数据，但为了简化示例，使用一个空的字节数组
	// 实际项目中应该嵌入一个真实的字体文件
}
