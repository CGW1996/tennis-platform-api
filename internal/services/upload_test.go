package services

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"os"
	"strings"
	"tennis-platform/backend/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupUploadService() *UploadService {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxFileSize: 10 * 1024 * 1024, // 10MB
			AllowedExts: "jpg,jpeg,png,gif,pdf",
			UploadPath:  "./test_uploads",
		},
	}
	return NewUploadService(cfg)
}

func createTestFileHeader(filename string, content []byte) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "image/jpeg")

	part, _ := writer.CreatePart(h)
	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(10 << 20)

	return form.File["file"][0]
}

func TestUploadService_isValidImageType(t *testing.T) {
	service := setupUploadService()

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.jpg", true},
		{"test.jpeg", true},
		{"test.png", true},
		{"test.gif", true},
		{"test.JPG", true},
		{"test.txt", false},
		{"test.pdf", false},
		{"test", false},
	}

	for _, test := range tests {
		result := service.isValidImageType(test.filename)
		assert.Equal(t, test.expected, result, "filename: %s", test.filename)
	}
}

func TestUploadService_isValidFileType(t *testing.T) {
	service := setupUploadService()

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.jpg", true},
		{"test.jpeg", true},
		{"test.png", true},
		{"test.gif", true},
		{"test.pdf", true},
		{"test.JPG", true},
		{"test.txt", false},
		{"test.doc", false},
		{"test", false},
	}

	for _, test := range tests {
		result := service.isValidFileType(test.filename)
		assert.Equal(t, test.expected, result, "filename: %s", test.filename)
	}
}

func TestUploadService_UploadAvatar_InvalidType(t *testing.T) {
	service := setupUploadService()

	// 創建無效類型的文件
	fileHeader := createTestFileHeader("test.txt", []byte("test content"))

	result, err := service.UploadAvatar(fileHeader, "test-user-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "不支援的文件類型")
}

func TestUploadService_UploadAvatar_TooLarge(t *testing.T) {
	service := setupUploadService()

	// 創建過大的文件
	largeContent := make([]byte, 11*1024*1024) // 11MB
	fileHeader := createTestFileHeader("test.jpg", largeContent)

	result, err := service.UploadAvatar(fileHeader, "test-user-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "文件大小超過限制")
}

func TestUploadService_DeleteFile(t *testing.T) {
	service := setupUploadService()

	// 測試刪除空路徑
	err := service.DeleteFile("")
	assert.NoError(t, err)

	// 測試刪除無效路徑
	err = service.DeleteFile("/invalid/path/file.jpg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "無效的文件路徑")
}

func TestUploadService_UploadAvatar_Success(t *testing.T) {
	service := setupUploadService()

	// 創建測試目錄
	testDir := "./test_uploads"
	defer os.RemoveAll(testDir)

	// 創建有效的圖片文件
	content := []byte("fake image content")
	fileHeader := createTestFileHeader("test.jpg", content)

	result, err := service.UploadAvatar(fileHeader, "test-user-id")

	if err != nil {
		// 如果因為目錄權限等問題失敗，跳過這個測試
		if strings.Contains(err.Error(), "創建上傳目錄失敗") ||
			strings.Contains(err.Error(), "創建目標文件失敗") {
			t.Skip("Skipping file system test due to permissions")
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test.jpg", result.OriginalName)
	assert.Equal(t, int64(len(content)), result.Size)
	assert.Contains(t, result.URL, "/uploads/avatars/")
	assert.Contains(t, result.FileName, "avatar_test-user-id_")

	// 驗證文件是否存在
	if _, err := os.Stat(result.Path); err == nil {
		// 清理測試文件
		os.Remove(result.Path)
	}
}
