package parser

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// DecodeQRFromBytes 从图片字节数据中解码二维码，返回解析到的链接列表
func DecodeQRFromBytes(data []byte) ([]string, error) {
	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 将图片转换为 gozxing 的二进制位图
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, fmt.Errorf("failed to create bitmap: %w", err)
	}

	// 创建 QR 码解码器
	qrReader := qrcode.NewQRCodeReader()

	// 解码
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return nil, fmt.Errorf("no QR code found: %w", err)
	}

	text := strings.TrimSpace(result.GetText())
	if text == "" {
		return nil, fmt.Errorf("QR code is empty")
	}

	// 可能包含多个链接（换行分隔）
	lines := strings.Split(text, "\n")
	var links []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			links = append(links, line)
		}
	}

	if len(links) == 0 {
		return nil, fmt.Errorf("no valid links found in QR code")
	}

	return links, nil
}
