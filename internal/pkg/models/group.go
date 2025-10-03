package models

import (
	"time"

	"gorm.io/gorm"
)

// Group represents a chat group
type Group struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Avatar      string         `gorm:"size:500" json:"avatar"`
	OwnerID     uint           `gorm:"not null;index" json:"owner_id"`
	Owner       User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members     []GroupMember  `gorm:"foreignKey:GroupID" json:"members,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (Group) TableName() string {
	return "groups"
}

// GroupMember represents a member of a group
type GroupMember struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GroupID   uint           `gorm:"not null;index" json:"group_id"`
	Group     Group          `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role      string         `gorm:"type:varchar(50);default:'member'" json:"role"` // admin, member
	JoinedAt  time.Time      `gorm:"autoCreateTime" json:"joined_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (GroupMember) TableName() string {
	return "group_members"
}

// GroupResponse is used for API responses
type GroupResponse struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Avatar      string     `json:"avatar"`
	OwnerID     uint       `json:"owner_id"`
	MemberCount int64      `json:"member_count"`
	CreatedAt   time.Time  `json:"created_at"`
}
