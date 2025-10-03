package services

import (
	"errors"

	"web-api/internal/pkg/database"
	"web-api/internal/pkg/models"

	"gorm.io/gorm"
)

type GroupService struct{}

var Group = &GroupService{}

// CreateGroupRequest represents group creation request
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

// AddMemberRequest represents add member request
type AddMemberRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role"` // admin or member
}

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(ownerID uint, req CreateGroupRequest) (*models.Group, error) {
	db := database.GetDB()

	// Create group in a transaction
	var group models.Group
	err := db.Transaction(func(tx *gorm.DB) error {
		// Create group
		group = models.Group{
			Name:        req.Name,
			Description: req.Description,
			Avatar:      req.Avatar,
			OwnerID:     ownerID,
		}

		if err := tx.Create(&group).Error; err != nil {
			return err
		}

		// Add owner as admin member
		member := models.GroupMember{
			GroupID: group.ID,
			UserID:  ownerID,
			Role:    "admin",
		}

		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load owner info
	db.Preload("Owner").First(&group, group.ID)

	return &group, nil
}

// AddMember adds a user to a group
func (s *GroupService) AddMember(groupID, requestorID uint, req AddMemberRequest) error {
	db := database.GetDB()

	// Verify requestor is admin of the group
	var requestorMember models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, requestorID).First(&requestorMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("you are not a member of this group")
		}
		return err
	}

	if requestorMember.Role != "admin" {
		return errors.New("only admins can add members")
	}

	// Check if user already a member
	var existingMember models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, req.UserID).First(&existingMember).Error; err == nil {
		return errors.New("user is already a member of this group")
	}

	// Verify user exists
	var user models.User
	if err := db.First(&user, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Add member
	role := req.Role
	if role == "" {
		role = "member"
	}

	member := models.GroupMember{
		GroupID: groupID,
		UserID:  req.UserID,
		Role:    role,
	}

	return db.Create(&member).Error
}

// RemoveMember removes a user from a group
func (s *GroupService) RemoveMember(groupID, requestorID, userID uint) error {
	db := database.GetDB()

	// Verify requestor is admin
	var requestorMember models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, requestorID).First(&requestorMember).Error; err != nil {
		return errors.New("you are not authorized to remove members")
	}

	if requestorMember.Role != "admin" {
		return errors.New("only admins can remove members")
	}

	// Cannot remove group owner
	var group models.Group
	if err := db.First(&group, groupID).Error; err != nil {
		return err
	}

	if group.OwnerID == userID {
		return errors.New("cannot remove group owner")
	}

	// Remove member
	return db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

// GetGroupMembers retrieves all members of a group
func (s *GroupService) GetGroupMembers(groupID, userID uint) ([]models.GroupMember, error) {
	db := database.GetDB()

	// Verify user is a member
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("you are not a member of this group")
		}
		return nil, err
	}

	// Get all members
	var members []models.GroupMember
	if err := db.Where("group_id = ?", groupID).
		Preload("User").
		Find(&members).Error; err != nil {
		return nil, err
	}

	return members, nil
}

// GetUserGroups retrieves all groups a user is member of
func (s *GroupService) GetUserGroups(userID uint) ([]models.Group, error) {
	db := database.GetDB()

	var groups []models.Group
	if err := db.Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userID).
		Preload("Owner").
		Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

// GetGroupByID retrieves a group by ID
func (s *GroupService) GetGroupByID(groupID, userID uint) (*models.Group, error) {
	db := database.GetDB()

	// Verify user is a member
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("you are not a member of this group")
		}
		return nil, err
	}

	var group models.Group
	if err := db.Preload("Owner").Preload("Members.User").First(&group, groupID).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

// UpdateGroup updates group information
func (s *GroupService) UpdateGroup(groupID, userID uint, updates map[string]interface{}) error {
	db := database.GetDB()

	// Verify user is admin
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		return errors.New("you are not authorized to update this group")
	}

	if member.Role != "admin" {
		return errors.New("only admins can update group information")
	}

	return db.Model(&models.Group{}).Where("id = ?", groupID).Updates(updates).Error
}

// DeleteGroup deletes a group (owner only)
func (s *GroupService) DeleteGroup(groupID, userID uint) error {
	db := database.GetDB()

	// Verify user is owner
	var group models.Group
	if err := db.First(&group, groupID).Error; err != nil {
		return err
	}

	if group.OwnerID != userID {
		return errors.New("only group owner can delete the group")
	}

	// Delete group and related data in transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// Delete all messages
		if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupMessage{}).Error; err != nil {
			return err
		}

		// Delete all members
		if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}

		// Delete group
		if err := tx.Delete(&group).Error; err != nil {
			return err
		}

		return nil
	})
}
