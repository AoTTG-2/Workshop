package entity

import (
	"time"
	"workshop/internal/types"
)

type Comment struct {
	ID        types.CommentID `json:"id"`
	PostID    types.PostID    `json:"post_id"`
	AuthorID  types.UserID    `json:"author_id"`
	Content   string          `json:"content"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
