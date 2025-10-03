package controllers

import (
	"net/http"
	"strconv"

	"web-api/internal/api/middlewares"
	"web-api/internal/api/services"

	"github.com/gin-gonic/gin"
)

type FileController struct{}

// UploadFile handles file upload
// @Summary Upload a file
// @Tags Files
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 201 {object} models.File
// @Router /api/files/upload [post]
func (ctrl *FileController) UploadFile(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file type
	if !services.FileServ.ValidateFileType(file.Header.Get("Content-Type")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	fileRecord, err := services.FileServ.UploadFile(userID, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, fileRecord)
}

// GetFile retrieves file information
// @Summary Get file by ID
// @Tags Files
// @Security BearerAuth
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} models.File
// @Router /api/files/:id [get]
func (ctrl *FileController) GetFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := services.FileServ.GetFileByID(uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, file)
}

// DeleteFile deletes a file
// @Summary Delete file
// @Tags Files
// @Security BearerAuth
// @Param id path int true "File ID"
// @Success 200
// @Router /api/files/:id [delete]
func (ctrl *FileController) DeleteFile(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	if err := services.FileServ.DeleteFile(uint(fileID), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// GetUserFiles retrieves all files uploaded by the user
// @Summary Get user files
// @Tags Files
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} models.File
// @Router /api/files [get]
func (ctrl *FileController) GetUserFiles(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	files, err := services.FileServ.GetUserFiles(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}
