package workshop

import (
	"time"
	"workshop/internal/types"
)

type PostContent struct {
	ID          types.PostContentID `json:"id" extensions:"x-order=0"`
	ContentType types.ContentType   `json:"content_type" extensions:"x-order=1"`
	ContentData string              `json:"content_data" extensions:"x-order=2"`
	IsLink      bool                `json:"is_link" extensions:"x-order=3"`
}

type PostModerationData struct {
	Status types.ModerationStatus `json:"status" extensions:"x-order=0"`
	Note   string                 `json:"note" extensions:"x-order=1"`
}

type PostInteractionData struct {
	IsFavorite bool           `json:"is_favorite" extensions:"x-order=0"`
	Vote       types.RateType `json:"vote" extensions:"x-order=1"`
}

type Post struct {
	ID              types.PostID        `json:"id" extensions:"x-order=0"`
	Title           string              `json:"title" extensions:"x-order=1"`
	Description     string              `json:"description" extensions:"x-order=2"`
	PreviewURL      string              `json:"preview_url" extensions:"x-order=3"`
	PostType        types.PostType      `json:"post_type" extensions:"x-order=4"`
	Tags            []string            `json:"tags" extensions:"x-order=5"`
	Contents        []PostContent       `json:"contents" extensions:"x-order=6"`
	CreatedAt       time.Time           `json:"created_at" extensions:"x-order=7"`
	UpdatedAt       time.Time           `json:"updated_at" extensions:"x-order=8"`
	ModerationData  PostModerationData  `json:"moderation_data" extensions:"x-order=9"`
	InteractionData PostInteractionData `json:"interaction_data" extensions:"x-order=10"`
	Rating          int                 `json:"rating" extensions:"x-order=11"`
	CommentsCount   int                 `json:"comments_count" extensions:"x-order=12"`
	FavoritesCount  int                 `json:"favorites_count" extensions:"x-order=13"`
}

type Comment struct {
	ID        types.CommentID `json:"id" extensions:"x-order=0"`
	PostID    types.PostID    `json:"post_id" extensions:"x-order=1"`
	AuthorID  types.UserID    `json:"author_id" extensions:"x-order=2"`
	Content   string          `json:"content" extensions:"x-order=3"`
	CreatedAt time.Time       `json:"created_at" extensions:"x-order=4"`
	UpdatedAt time.Time       `json:"updated_at" extensions:"x-order=5"`
}

type ModerationAction struct {
	ID          types.ModerationActionID   `json:"id" extensions:"x-order=0"`
	Post        *Post                      `json:"post" extensions:"x-order=1"`
	ModeratorID types.UserID               `json:"moderator_id" extensions:"x-order=2"`
	Action      types.ModerationActionType `json:"action" extensions:"x-order=3"`
	Note        string                     `json:"note" extensions:"x-order=4"`
	CreatedAt   time.Time                  `json:"created_at" extensions:"x-order=5"`
}
