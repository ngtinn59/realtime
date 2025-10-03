package entities

import "time"

type BaseModel struct {
	ID        uint64    `gorm:"column:id;primary_key;auto_increment;" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updatedAt"`
}
