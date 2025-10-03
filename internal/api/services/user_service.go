package services

import (
	"errors"
	"fmt"
	"time"

	"web-api/internal/pkg/database"
	"web-api/internal/pkg/models"
	"web-api/internal/pkg/redis"
	"web-api/internal/pkg/utils"

	"gorm.io/gorm"
)

type UserService struct{}

var User = &UserService{}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string               `json:"token"`
	User  models.UserResponse  `json:"user"`
}

// Register creates a new user account
func (s *UserService) Register(req RegisterRequest) (*AuthResponse, error) {
	db := database.GetDB()

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email or username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
		IsOnline: false,
	}

	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// Login authenticates a user
func (s *UserService) Login(req LoginRequest) (*AuthResponse, error) {
	db := database.GetDB()

	// Find user by email
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Check password
	if !utils.CheckPassword(user.Password, req.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Update last seen
	now := time.Now()
	user.LastSeen = &now
	db.Save(&user)

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// GetOnlineUsers returns list of online users
func (s *UserService) GetOnlineUsers() ([]models.UserResponse, error) {
	db := database.GetDB()

	// Get online user IDs from Redis
	onlineUserIDs, err := redis.GetOnlineUsers()
	if err != nil {
		return nil, err
	}

	if len(onlineUserIDs) == 0 {
		return []models.UserResponse{}, nil
	}

	// Convert string IDs to uint
	var userIDs []uint
	for _, idStr := range onlineUserIDs {
		var id uint
		if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil {
			userIDs = append(userIDs, id)
		}
	}

	// Fetch users from database
	var users []models.User
	if err := db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
		responses[i].IsOnline = true
	}

	return responses, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	db := database.GetDB()

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUserStatus updates user online status
func (s *UserService) UpdateUserStatus(userID uint, isOnline bool) error {
	db := database.GetDB()

	updates := map[string]interface{}{
		"is_online": isOnline,
	}

	if !isOnline {
		now := time.Now()
		updates["last_seen"] = now
	}

	return db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

// SearchUsers searches for users by username or email
func (s *UserService) SearchUsers(query string, limit int) ([]models.UserResponse, error) {
	db := database.GetDB()

	var users []models.User
	if err := db.Where("username LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, err
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, nil
}
