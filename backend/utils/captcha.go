package utils

import (
	"bytes"
	"crypto/rand"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/big"
)

// Captcha 验证码结构
type Captcha struct {
	ID     string
	Answer int
	Image  []byte
}

// GenerateCaptcha 生成验证码
func GenerateCaptcha(width, height, blockSize int) (*Captcha, error) {
	// 创建背景图片
	bg := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景色
	bgColor := color.RGBA{240, 240, 240, 255}
	draw.Draw(bg, bg.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 生成随机背景图案
	drawBackgroundPattern(bg, width, height)

	// 生成随机滑块位置
	maxX := width - blockSize - 10
	answer, err := rand.Int(rand.Reader, big.NewInt(int64(maxX)))
	if err != nil {
		return nil, err
	}
	sliderX := int(answer.Int64()) + 10
	sliderY := height/2 - blockSize/2

	// 定义挖空区域的颜色（与背景形成对比）
	holeColor := color.RGBA{255, 255, 255, 255}

	// 挖空背景（创建目标区域）
	drawHole(bg, sliderX, sliderY, blockSize, holeColor)

	// 创建滑块图片
	slider := image.NewRGBA(image.Rect(0, 0, blockSize, blockSize))
	// 填充滑块背景为透明
	for x := 0; x < blockSize; x++ {
		for y := 0; y < blockSize; y++ {
			slider.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// 绘制滑块内容（与挖空区域匹配）
	drawSliderContent(slider, blockSize)

	// 合并背景和滑块
	combined := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(combined, combined.Bounds(), bg, image.Point{}, draw.Src)
	draw.Draw(combined, image.Rect(sliderX, sliderY, sliderX+blockSize, sliderY+blockSize), slider, image.Point{}, draw.Over)

	// 转换为字节流
	var buf bytes.Buffer
	if err := png.Encode(&buf, combined); err != nil {
		return nil, err
	}

	// 生成验证码ID
	id := generateID(16)

	return &Captcha{
		ID:     id,
		Answer: sliderX,
		Image:  buf.Bytes(),
	}, nil
}

// drawPath 绘制滑块路径
func drawPath(img *image.RGBA, x, y, size int, c color.RGBA) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			distance := math.Sqrt(float64(i*i + j*j))
			if distance <= float64(size/2) {
				x1 := x + i
				y1 := y + j
				if x1 < img.Bounds().Max.X && y1 < img.Bounds().Max.Y {
					img.Set(x1, y1, c)
				}
			}
		}
	}
}

// drawBlock 绘制滑块内容
func drawBlock(img *image.RGBA, size int, c color.RGBA) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			distance := math.Sqrt(float64(i*i + j*j))
			if distance <= float64(size/2) {
				img.Set(i, j, c)
			}
		}
	}
}

// drawBackgroundPattern 绘制有意义的背景图案
func drawBackgroundPattern(img *image.RGBA, width, height int) {
	// 生成随机颜色
	getRandomColor := func() color.RGBA {
		r, _ := rand.Int(rand.Reader, big.NewInt(100))
		g, _ := rand.Int(rand.Reader, big.NewInt(100))
		b, _ := rand.Int(rand.Reader, big.NewInt(100))
		return color.RGBA{uint8(155 + r.Int64()), uint8(155 + g.Int64()), uint8(155 + b.Int64()), 255}
	}

	// 绘制随机几何图形
	shapeType, _ := rand.Int(rand.Reader, big.NewInt(3))
	switch shapeType.Int64() {
	case 0:
		// 绘制随机线条
		for i := 0; i < 10; i++ {
			x1, _ := rand.Int(rand.Reader, big.NewInt(int64(width)))
			y1, _ := rand.Int(rand.Reader, big.NewInt(int64(height)))
			x2, _ := rand.Int(rand.Reader, big.NewInt(int64(width)))
			y2, _ := rand.Int(rand.Reader, big.NewInt(int64(height)))
			lineColor := getRandomColor()
			drawLine(img, int(x1.Int64()), int(y1.Int64()), int(x2.Int64()), int(y2.Int64()), lineColor)
		}
	case 1:
		// 绘制随机圆点
		for i := 0; i < 20; i++ {
			x, _ := rand.Int(rand.Reader, big.NewInt(int64(width)))
			y, _ := rand.Int(rand.Reader, big.NewInt(int64(height)))
			radius, _ := rand.Int(rand.Reader, big.NewInt(5))
			dotColor := getRandomColor()
			drawCircle(img, int(x.Int64()), int(y.Int64()), int(radius.Int64()), dotColor)
		}
	case 2:
		// 绘制随机矩形
		for i := 0; i < 5; i++ {
			x, _ := rand.Int(rand.Reader, big.NewInt(int64(width-20)))
			y, _ := rand.Int(rand.Reader, big.NewInt(int64(height-20)))
			w, _ := rand.Int(rand.Reader, big.NewInt(30))
			h, _ := rand.Int(rand.Reader, big.NewInt(30))
			rectColor := getRandomColor()
			drawRectangle(img, int(x.Int64()), int(y.Int64()), int(w.Int64()), int(h.Int64()), rectColor)
		}
	}
}

// drawLine 绘制直线
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	sy := 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		if x1 >= 0 && x1 < img.Bounds().Max.X && y1 >= 0 && y1 < img.Bounds().Max.Y {
			img.Set(x1, y1, c)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// drawCircle 绘制圆形
func drawCircle(img *image.RGBA, x, y, radius int, c color.RGBA) {
	for i := -radius; i <= radius; i++ {
		for j := -radius; j <= radius; j++ {
			if i*i+j*j <= radius*radius {
				x1 := x + i
				y1 := y + j
				if x1 >= 0 && x1 < img.Bounds().Max.X && y1 >= 0 && y1 < img.Bounds().Max.Y {
					img.Set(x1, y1, c)
				}
			}
		}
	}
}

// drawRectangle 绘制矩形
func drawRectangle(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			x1 := x + i
			y1 := y + j
			if x1 >= 0 && x1 < img.Bounds().Max.X && y1 >= 0 && y1 < img.Bounds().Max.Y {
				img.Set(x1, y1, c)
			}
		}
	}
}

// abs 返回整数的绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawHole 挖空背景，创建目标区域
func drawHole(img *image.RGBA, x, y, size int, c color.RGBA) {
	// 绘制一个圆形挖空区域
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			distance := math.Sqrt(float64((i-size/2)*(i-size/2) + (j-size/2)*(j-size/2)))
			if distance <= float64(size/2) {
				x1 := x + i
				y1 := y + j
				if x1 < img.Bounds().Max.X && y1 < img.Bounds().Max.Y {
					img.Set(x1, y1, c)
				}
			}
		}
	}
}

// drawSliderContent 绘制滑块内容
func drawSliderContent(img *image.RGBA, size int) {
	// 生成随机颜色
	r, _ := rand.Int(rand.Reader, big.NewInt(100))
	g, _ := rand.Int(rand.Reader, big.NewInt(100))
	b, _ := rand.Int(rand.Reader, big.NewInt(100))
	sliderColor := color.RGBA{uint8(55 + r.Int64()), uint8(55 + g.Int64()), uint8(55 + b.Int64()), 255}

	// 绘制圆形滑块
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			distance := math.Sqrt(float64((i-size/2)*(i-size/2) + (j-size/2)*(j-size/2)))
			if distance <= float64(size/2) {
				img.Set(i, j, sliderColor)
			}
		}
	}

	// 添加边框，使滑块更明显
	borderColor := color.RGBA{0, 0, 0, 255}
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			distance := math.Sqrt(float64((i-size/2)*(i-size/2) + (j-size/2)*(j-size/2)))
			if distance >= float64(size/2-2) && distance <= float64(size/2) {
				img.Set(i, j, borderColor)
			}
		}
	}
}

// generateID 生成随机ID
func generateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			continue
		}
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

// VerifyCaptcha 验证验证码
func VerifyCaptcha(userAnswer, correctAnswer int, tolerance int) bool {
	return math.Abs(float64(userAnswer-correctAnswer)) <= float64(tolerance)
}
