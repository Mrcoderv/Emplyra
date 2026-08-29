package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}

type SoftDeleteModel struct {
	BaseModel
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
