package models

import (
	"time"

	"gorm.io/gorm"
)

// CallStatus defines the status of a video call
type CallStatus string

const (
	CallStatusInitiating CallStatus = "initiating"
	CallStatusRinging    CallStatus = "ringing"
	CallStatusConnected  CallStatus = "connected"
	CallStatusEnded      CallStatus = "ended"
	CallStatusRejected   CallStatus = "rejected"
	CallStatusMissed     CallStatus = "missed"
)

// CallType defines the type of call
type CallType string

const (
	CallTypePrivate CallType = "private"
	CallTypeGroup   CallType = "group"
)

// VideoCall represents a video call session
type VideoCall struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	InitiatorID uint       `gorm:"not null;index" json:"initiator_id"`
	Initiator   User       `gorm:"foreignKey:InitiatorID" json:"initiator,omitempty"`
	Type        CallType   `gorm:"type:varchar(20);not null" json:"type"`
	Status      CallStatus `gorm:"type:varchar(20);not null;default:'initiating'" json:"status"`

	// For private calls
	ReceiverID *uint `gorm:"index" json:"receiver_id,omitempty"`
	Receiver   *User `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`

	// For group calls
	GroupID *uint  `gorm:"index" json:"group_id,omitempty"`
	Group   *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`

	// Call timing
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Duration  *int       `json:"duration,omitempty"` // Duration in seconds

	// WebRTC signaling data (stored as JSON)
	OfferSDP  string `gorm:"type:text" json:"offer_sdp,omitempty"`
	AnswerSDP string `gorm:"type:text" json:"answer_sdp,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (VideoCall) TableName() string {
	return "video_calls"
}

// CallParticipant represents a participant in a video call
type CallParticipant struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CallID    uint           `gorm:"not null;index" json:"call_id"`
	Call      VideoCall      `gorm:"foreignKey:CallID" json:"call,omitempty"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	JoinedAt  *time.Time     `json:"joined_at,omitempty"`
	LeftAt    *time.Time     `json:"left_at,omitempty"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (CallParticipant) TableName() string {
	return "call_participants"
}

// ICECandidate represents WebRTC ICE candidates exchanged during signaling
type ICECandidate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CallID    uint      `gorm:"not null;index" json:"call_id"`
	Call      VideoCall `gorm:"foreignKey:CallID" json:"call,omitempty"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Candidate string    `gorm:"type:text;not null" json:"candidate"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name
func (ICECandidate) TableName() string {
	return "ice_candidates"
}
