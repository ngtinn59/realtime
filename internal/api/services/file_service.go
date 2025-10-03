package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"web-api/internal/pkg/database"
	"web-api/internal/pkg/models"

	"github.com/google/uuid"
)

type FileService struct{}

var FileServ = &FileService{}

const (
	MaxFileSize = 10 * 1024 * 1024 // 10MB
	UploadDir   = "./uploads"
)

// UploadFile handles file upload
func (s *FileService) UploadFile(userID uint, fileHeader *multipart.FileHeader) (*models.File, error) {
	// Validate file size
	if fileHeader.Size > MaxFileSize {
		return nil, errors.New("file size exceeds maximum limit of 10MB")
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create uploads directory if not exists
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return nil, err
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	
	// Create subdirectory based on date
	dateDir := time.Now().Format("2006-01-02")
	fullDir := filepath.Join(UploadDir, dateDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(fullDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	// Create file record in database
	db := database.GetDB()
	
	fileRecord := models.File{
		UploaderID:   userID,
		Filename:     filename,
		OriginalName: fileHeader.Filename,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		Size:         fileHeader.Size,
		Path:         filePath,
		URL:          fmt.Sprintf("/uploads/%s/%s", dateDir, filename),
	}

	if err := db.Create(&fileRecord).Error; err != nil {
		// Delete uploaded file if database insert fails
		os.Remove(filePath)
		return nil, err
	}

	return &fileRecord, nil
}

// GetFileByID retrieves file information by ID
func (s *FileService) GetFileByID(fileID uint) (*models.File, error) {
	db := database.GetDB()

	var file models.File
	if err := db.Preload("Uploader").First(&file, fileID).Error; err != nil {
		return nil, err
	}

	return &file, nil
}

// DeleteFile deletes a file
func (s *FileService) DeleteFile(fileID, userID uint) error {
	db := database.GetDB()

	var file models.File
	if err := db.First(&file, fileID).Error; err != nil {
		return err
	}

	// Only uploader can delete the file
	if file.UploaderID != userID {
		return errors.New("unauthorized to delete this file")
	}

	// Delete physical file
	if err := os.Remove(file.Path); err != nil {
		// Log error but continue to delete database record
		fmt.Printf("Warning: failed to delete physical file: %v\n", err)
	}

	// Delete database record
	return db.Delete(&file).Error
}

// GetUserFiles retrieves all files uploaded by a user
func (s *FileService) GetUserFiles(userID uint, limit, offset int) ([]models.File, error) {
	db := database.GetDB()

	var files []models.File
	if err := db.Where("uploader_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error; err != nil {
		return nil, err
	}

	return files, nil
}

// ValidateFileType validates if file type is allowed
func (s *FileService) ValidateFileType(mimeType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/pdf": true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/plain": true,
	}

	return allowedTypes[mimeType]
}
