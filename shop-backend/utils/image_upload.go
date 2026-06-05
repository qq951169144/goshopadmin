package utils

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 头像上传常量
const (
	// AvatarMaxFileSize 头像最大文件大小（2MB）
	AvatarMaxFileSize int64 = 2 * 1024 * 1024
	// AvatarStoragePath 头像存储路径
	AvatarStoragePath string = "./uploads/avatars"
)

// 允许的图片文件类型
var allowedAvatarTypes = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

// 头像上传错误信息
const (
	ErrAvatarSizeExceeded     = "头像文件大小超过限制（最大2MB）"
	ErrAvatarTypeUnsupported  = "不支持的头像文件类型"
	ErrAvatarCreateDirFailed  = "创建头像存储目录失败"
	ErrAvatarOpenFileFailed   = "打开上传文件失败"
	ErrAvatarCreateFileFailed = "创建目标文件失败"
	ErrAvatarWriteFileFailed  = "写入文件失败"
	ErrInvalidImageURL        = "无效的图片URL"
)

// AvatarUploadConfig 头像上传配置
type AvatarUploadConfig struct {
	MaxSize      int64    // 最大文件大小（字节）
	AllowedTypes []string // 允许的文件类型
	StoragePath  string   // 存储路径
	Domain       string   // 域名
}

// DefaultAvatarUploadConfig 默认头像上传配置
var DefaultAvatarUploadConfig = AvatarUploadConfig{
	MaxSize:      AvatarMaxFileSize,
	AllowedTypes: allowedAvatarTypes,
	StoragePath:  AvatarStoragePath,
	Domain:       "http://localhost:8081",
}

// UploadAvatar 上传头像图片
func UploadAvatar(file *multipart.FileHeader, customerID int) (string, error) {
	return UploadAvatarWithConfig(file, customerID, DefaultAvatarUploadConfig)
}

// UploadAvatarWithConfig 使用自定义配置上传头像
func UploadAvatarWithConfig(file *multipart.FileHeader, customerID int, config AvatarUploadConfig) (string, error) {
	// 验证文件大小
	if file.Size > config.MaxSize {
		return "", errors.New(ErrAvatarSizeExceeded)
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := false
	for _, allowedExt := range config.AllowedTypes {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", errors.New(ErrAvatarTypeUnsupported)
	}

	// 生成唯一文件名
	fileName := generateAvatarFileName(customerID, ext)

	// 确保存储目录存在
	customerPath := filepath.Join(config.StoragePath, fmt.Sprintf("%d", customerID))
	if err := os.MkdirAll(customerPath, 0755); err != nil {
		return "", errors.New(ErrAvatarCreateDirFailed)
	}

	// 保存文件
	filePath := filepath.Join(customerPath, fileName)
	src, err := file.Open()
	if err != nil {
		return "", errors.New(ErrAvatarOpenFileFailed)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", errors.New(ErrAvatarCreateFileFailed)
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
			return "", errors.New(ErrAvatarWriteFileFailed)
		}
	}

	// 生成可访问的URL
	imagePath := fmt.Sprintf("/uploads/avatars/%d/%s", customerID, fileName)
	imageURL := fmt.Sprintf("%s%s", config.Domain, imagePath)

	return imageURL, nil
}

// generateAvatarFileName 生成头像唯一文件名
func generateAvatarFileName(customerID int, ext string) string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("avatar_%d_%d%s", customerID, timestamp, ext)
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
