package entity

import (
	"time"
	"workshop/internal/types"
)

type ModerationAction struct {
	ID          types.ModerationActionID   `json:"id"`
	PostID      types.PostID               `json:"post_id"`
	Post        *Post                      `json:"post"`
	ModeratorID types.UserID               `json:"moderator_id"`
	Action      types.ModerationActionType `json:"action"`
	Note        string                     `json:"note"`
	CreatedAt   time.Time                  `json:"created_at"`
}
