package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"tennis-platform/backend/internal/config"
	"time"

	"github.com/google/uuid"
)

// UploadService 文件上傳服務
type UploadService struct {
	config *config.Config
}

// NewUploadService 創建新的文件上傳服務
func NewUploadService(cfg *config.Config) *UploadService {
	return &UploadService{
		config: cfg,
	}
}

// UploadResult 上傳結果
type UploadResult struct {
	FileName     string `json:"fileName"`
	OriginalName string `json:"originalName"`
	Size         int64  `json:"size"`
	URL          string `json:"url"`
	Path         string `json:"path"`
}

// UploadAvatar 上傳頭像
func (us *UploadService) UploadAvatar(file *multipart.FileHeader, userID string) (*UploadResult, error) {
	// 驗證文件類型
	if !us.isValidImageType(file.Filename) {
		return nil, errors.New("不支援的文件類型，僅支援 jpg, jpeg, png, gif")
	}

	// 驗證文件大小
	if file.Size > us.config.Upload.MaxFileSize {
		return nil, fmt.Errorf("文件大小超過限制，最大允許 %d MB", us.config.Upload.MaxFileSize/(1024*1024))
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("avatar_%s_%d%s", userID, time.Now().Unix(), ext)

	// 創建上傳目錄
	avatarDir := filepath.Join(us.config.Upload.UploadPath, "avatars")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return nil, fmt.Errorf("創建上傳目錄失敗: %v", err)
	}

	// 完整文件路徑
	filePath := filepath.Join(avatarDir, fileName)

	// 打開上傳的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打開上傳文件失敗: %v", err)
	}
	defer src.Close()

	// 創建目標文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("創建目標文件失敗: %v", err)
	}
	defer dst.Close()

	// 複製文件內容
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("保存文件失敗: %v", err)
	}

	// 生成訪問URL
	url := fmt.Sprintf("/uploads/avatars/%s", fileName)

	return &UploadResult{
		FileName:     fileName,
		OriginalName: file.Filename,
		Size:         file.Size,
		URL:          url,
		Path:         filePath,
	}, nil
}

// UploadFile 通用文件上傳
func (us *UploadService) UploadFile(file *multipart.FileHeader, subDir string) (*UploadResult, error) {
	// 驗證文件類型
	if !us.isValidFileType(file.Filename) {
		return nil, errors.New("不支援的文件類型")
	}

	// 驗證文件大小
	if file.Size > us.config.Upload.MaxFileSize {
		return nil, fmt.Errorf("文件大小超過限制，最大允許 %d MB", us.config.Upload.MaxFileSize/(1024*1024))
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	// 創建上傳目錄
	uploadDir := filepath.Join(us.config.Upload.UploadPath, subDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("創建上傳目錄失敗: %v", err)
	}

	// 完整文件路徑
	filePath := filepath.Join(uploadDir, fileName)

	// 打開上傳的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打開上傳文件失敗: %v", err)
	}
	defer src.Close()

	// 創建目標文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("創建目標文件失敗: %v", err)
	}
	defer dst.Close()

	// 複製文件內容
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("保存文件失敗: %v", err)
	}

	// 生成訪問URL
	url := fmt.Sprintf("/uploads/%s/%s", subDir, fileName)

	return &UploadResult{
		FileName:     fileName,
		OriginalName: file.Filename,
		Size:         file.Size,
		URL:          url,
		Path:         filePath,
	}, nil
}

// DeleteFile 刪除文件
func (us *UploadService) DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	// 確保文件路徑在上傳目錄內
	if !strings.HasPrefix(filePath, us.config.Upload.UploadPath) {
		return errors.New("無效的文件路徑")
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("刪除文件失敗: %v", err)
	}

	return nil
}

// isValidImageType 檢查是否為有效的圖片類型
func (us *UploadService) isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif"}

	for _, validExt := range imageExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// isValidFileType 檢查是否為有效的文件類型
func (us *UploadService) isValidFileType(filename string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	allowedExts := strings.Split(strings.ToLower(us.config.Upload.AllowedExts), ",")

	for _, allowedExt := range allowedExts {
		if ext == strings.TrimSpace(allowedExt) {
			return true
		}
	}
	return false
}
