package entity

import (
	"gorm.io/gorm"
	"time"
	"workshop/internal/types"
)

type Post struct {
	ID               types.PostID              `json:"id"`
	AuthorID         types.UserID              `json:"author_id"`
	Title            string                    `json:"title"`
	Description      string                    `json:"description"`
	PreviewURL       string                    `json:"preview_url"`
	PostType         types.PostType            `json:"post_type"`
	Tags             []*Tag                    `json:"tags,omitempty" gorm:"many2many:post_tags;"`
	Contents         []*PostContent            `json:"contents,omitempty"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	DeletedAt        gorm.DeletedAt            `json:"deleted_at,omitempty"`
	LastModerationID *types.ModerationActionID `json:"last_moderation_id,omitempty"`
	LastModeration   *ModerationAction         `json:"last_moderation,omitempty" gorm:"references:LastModerationID;foreignKey:ID"`
	Rating           int                       `json:"rating"`
	CommentsCount    int                       `json:"comments_count"`
	FavoritesCount   int                       `json:"favorites_count"`

	MyFavorite *Favorite `json:"my_favorite,omitempty"`
	MyVote     *Vote     `json:"my_vote,omitempty"    `
}
