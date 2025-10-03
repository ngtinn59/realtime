package models

import (
	"time"

	"gorm.io/gorm"
)

// MessageType defines the type of message
type MessageType string

const (
	MessageTypeText MessageType = "text"
	MessageTypeFile MessageType = "file"
)

// PrivateMessage represents a one-to-one message
type PrivateMessage struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	SenderID   uint            `gorm:"not null;index" json:"sender_id"`
	Sender     User            `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReceiverID uint            `gorm:"not null;index" json:"receiver_id"`
	Receiver   User            `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	Content    string          `gorm:"type:text;not null" json:"content"`
	Type       MessageType     `gorm:"type:varchar(20);default:'text'" json:"type"`
	FileID     *uint           `gorm:"index" json:"file_id,omitempty"`
	File       *File           `gorm:"foreignKey:FileID" json:"file,omitempty"`
	IsRead     bool            `gorm:"default:false" json:"is_read"`
	ReadAt     *time.Time      `json:"read_at"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (PrivateMessage) TableName() string {
	return "private_messages"
}

// GroupMessage represents a message in a group chat
type GroupMessage struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GroupID   uint           `gorm:"not null;index" json:"group_id"`
	Group     Group          `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	SenderID  uint           `gorm:"not null;index" json:"sender_id"`
	Sender    User           `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Type      MessageType    `gorm:"type:varchar(20);default:'text'" json:"type"`
	FileID    *uint          `gorm:"index" json:"file_id,omitempty"`
	File      *File          `gorm:"foreignKey:FileID" json:"file,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (GroupMessage) TableName() string {
	return "group_messages"
}
