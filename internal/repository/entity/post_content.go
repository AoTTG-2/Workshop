package entity

import (
	"time"
	"workshop/internal/types"
)

type PostContent struct {
	ID          types.PostContentID `json:"id"`
	PostID      types.PostID        `json:"post_id"`
	ContentType types.ContentType   `json:"content_type"`
	ContentData string              `json:"content_data"`
	IsLink      bool                `json:"is_link"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}
