package utils

import (
	"errors"
	"fmt"
	"goshopadmin/config"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 图片上传常量
const (
	// DefaultMaxFileSize 默认最大文件大小（2MB）
	DefaultMaxFileSize int64 = 2 * 1024 * 1024
	// MaxURLLength 最大URL长度（对应MySQL varchar(255)）
	MaxURLLength int = 255
	// DefaultStoragePath 默认存储路径
	DefaultStoragePath string = "./uploads"
)

// 允许的图片文件类型
var AllowedImageTypes = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

// 错误信息常量
const (
	ErrFileSizeExceeded       = "文件大小超过限制"
	ErrUnsupportedFileType    = "不支持的文件类型"
	ErrCreateDirFailed        = "创建存储目录失败"
	ErrOpenFileFailed         = "打开上传文件失败"
	ErrCreateFileFailed       = "创建目标文件失败"
	ErrWriteFileFailed        = "写入文件失败"
	ErrImageURLLengthExceeded = "生成的图片URL过长"
	ErrRenameFileFailed       = "重命名文件失败"
	ErrInvalidImageURL        = "无效的图片URL"
)

// ImageUploadConfig 图片上传配置
type ImageUploadConfig struct {
	MaxSize      int64    // 最大文件大小（字节）
	AllowedTypes []string // 允许的文件类型
	StoragePath  string   // 存储路径
}

// DefaultImageUploadConfig 默认图片上传配置
var DefaultImageUploadConfig = ImageUploadConfig{
	MaxSize:      DefaultMaxFileSize,
	AllowedTypes: AllowedImageTypes,
	StoragePath:  DefaultStoragePath,
}

// UploadImage 上传图片
func UploadImage(file *multipart.FileHeader, merchantID int) (string, error) {
	return UploadImageWithConfig(file, merchantID, DefaultImageUploadConfig)
}

// UploadImageWithConfig 使用自定义配置上传图片
func UploadImageWithConfig(file *multipart.FileHeader, merchantID int, imageUploadConfig ImageUploadConfig) (string, error) {
	// 验证文件大小
	if file.Size > imageUploadConfig.MaxSize {
		return "", errors.New(ErrFileSizeExceeded)
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := false
	for _, allowedExt := range imageUploadConfig.AllowedTypes {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", errors.New(ErrUnsupportedFileType)
	}

	// 生成唯一文件名
	fileName := generateUniqueFileName(file.Filename, merchantID)

	// 确保存储目录存在
	merchantPath := filepath.Join(imageUploadConfig.StoragePath, fmt.Sprintf("%d", merchantID))
	if err := os.MkdirAll(merchantPath, 0755); err != nil {
		return "", errors.New(ErrCreateDirFailed)
	}

	// 保存文件
	filePath := filepath.Join(merchantPath, fileName)
	src, err := file.Open()
	if err != nil {
		return "", errors.New(ErrOpenFileFailed)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", errors.New(ErrCreateFileFailed)
	}
	defer dst.Close()

	// 读取文件内容并写入目标文件
	buffer := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := src.Read(buffer)
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		if _, err := dst.Write(buffer[:n]); err != nil {
			return "", errors.New(ErrWriteFileFailed)
		}
	}

	// 生成可访问的URL
	imagePath := fmt.Sprintf("/uploads/%d/%s", merchantID, fileName)
	imageURL := fmt.Sprintf("%s%s", config.AppConfig.Domain, imagePath)

	// 确保URL长度不超过255个字符（按字符数，对应 MySQL varchar(255)）
	if !IsRuneLenMax(imageURL, MaxURLLength) {
		// 生成更短的文件名
		shortFileName := fmt.Sprintf("%d_%d%s", merchantID, time.Now().UnixNano()/int64(time.Millisecond), ext)
		shortImagePath := fmt.Sprintf("/uploads/%d/%s", merchantID, shortFileName)
		imageURL = fmt.Sprintf("%s%s", config.AppConfig.Domain, shortImagePath)

		// 如果仍然超过长度，返回错误
		if !IsRuneLenMax(imageURL, MaxURLLength) {
			return "", errors.New(ErrImageURLLengthExceeded)
		}

		// 更新文件路径并重新保存
		shortFilePath := filepath.Join(merchantPath, shortFileName)
		if err := os.Rename(filePath, shortFilePath); err != nil {
			return "", errors.New(ErrRenameFileFailed)
		}
	}

	return imageURL, nil
}

// generateUniqueFileName 生成唯一文件名
func generateUniqueFileName(originalName string, merchantID int) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s_%d_%d%s", name, merchantID, timestamp, ext)
}

// DeleteImage 删除图片
func DeleteImage(imageURL string) error {
	// 从URL中提取文件路径
	var relativePath string
	if strings.HasPrefix(imageURL, "/uploads/") {
		relativePath = imageURL
	} else {
		// 处理完整URL，提取相对路径部分
		if strings.Contains(imageURL, "/uploads/") {
			relativePath = "/" + strings.Split(imageURL, "/uploads/")[1]
			relativePath = "/uploads" + relativePath
		} else {
			return errors.New(ErrInvalidImageURL)
		}
	}

	// 转换为本地文件路径
	filePath := "." + relativePath

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，视为删除成功
	}

	// 删除文件
	return os.Remove(filePath)
}
