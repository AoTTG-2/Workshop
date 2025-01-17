package entity

import (
	"time"
	"workshop/internal/types"
)

type Favorite struct {
	ID        types.FavoriteID `json:"id"`
	PostID    types.PostID     `json:"post_id"`
	UserID    types.UserID     `json:"user_id"`
	CreatedAt time.Time        `json:"created_at"`
}
