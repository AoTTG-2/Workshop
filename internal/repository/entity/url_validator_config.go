package entity

import (
	"github.com/lib/pq"
	"time"
)

type URLValidatorConfig struct {
	Type       string         `json:"type"`
	Protocols  pq.StringArray `gorm:"type:text[]" json:"protocols"`
	Domains    pq.StringArray `gorm:"type:text[]" json:"domains"`
	Extensions pq.StringArray `gorm:"type:text[]" json:"extensions"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
