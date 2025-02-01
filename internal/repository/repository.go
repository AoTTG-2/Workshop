package repository

import (
	"context"
	"time"
	"workshop/internal/repository/entity"
	"workshop/internal/types"
)

type Appender[V any] interface {
	Append(V)
}

type BaseFilter struct {
	Limit  int
	Offset int
}

type Order int

const (
	OrderAsc Order = iota + 1
	OrderDesc
)

type FavoritesFilter struct {
	BaseFilter
	UserID         types.UserID
	CreatedAtOrder Order
}

type VotesFilter struct {
	BaseFilter
	PostID         *int
	VoterID        *types.UserID
	CreatedAtOrder Order
}

type CommentsFilter struct {
	BaseFilter
	PostID         *int
	UserID         *types.UserID
	CreatedAtOrder Order
}

type GetPostsFilter struct {
	BaseFilter
	Query               string
	AuthorID            types.UserID
	ForUserID           types.UserID
	PostType            types.PostType
	Tags                []string
	CreatedAtOrder      Order
	UpdatedAtOrder      Order
	FavoritesCountOrder Order
	RatingOrder         Order
	CommentsCountOrder  Order
	OnlyFavorites       bool
	RatingFilter        types.RateType
	OnlyApproved        bool
	IncludePostContents bool
	IncludeTags         bool
	ShowDeclined        bool
}

type GetPostFilter struct {
	PostID              types.PostID
	ForUserID           types.UserID
	IncludePostContents bool
	IncludeTags         bool
	ShowDeclined        bool
}

type GetModerationActionsFilter struct {
	BaseFilter
	ModeratorID    types.UserID
	PostID         types.PostID
	Action         types.ModerationActionType
	CreatedAtOrder Order
	IncludePost    bool
}

type GetCommentsFilter struct {
	BaseFilter
	AuthorID       types.UserID
	PostID         types.PostID
	CreatedAtOrder Order
}

type Repository interface {
	Migrate(_ context.Context) error
	Drop(_ context.Context) error
	Truncate(ctx context.Context, tables []string) error
	Close()
	CreatePostWithContentsAndTags(ctx context.Context, p *entity.Post) error
	UpdatePost(ctx context.Context, p *entity.Post) error
	DeletePost(ctx context.Context, postID types.PostID, hard bool) error
	PurgeSoftDeletedPosts(ctx context.Context, olderThan time.Time) (int, error)
	GetPost(ctx context.Context, filter GetPostFilter) (*entity.Post, error)
	GetPosts(ctx context.Context, filter GetPostsFilter) ([]*entity.Post, error)
	CreateModerationAction(ctx context.Context, moderationAction *entity.ModerationAction) error
	GetModerationActions(ctx context.Context, filter GetModerationActionsFilter) ([]*entity.ModerationAction, error)
	AddPostToFavorites(ctx context.Context, favorite *entity.Favorite) error
	RemovePostFromFavoritesByPostAndUser(ctx context.Context, favorite *entity.Favorite) error
	RatePost(ctx context.Context, vote *entity.Vote) error
	RemovePostRateByPostAndUser(ctx context.Context, vote *entity.Vote) error
	AddComment(ctx context.Context, comment *entity.Comment) error
	UpdateComment(ctx context.Context, comment *entity.Comment) error
	DeleteComment(ctx context.Context, comment *entity.Comment) error
	GetCommentByID(ctx context.Context, id types.CommentID) (*entity.Comment, error)
	GetComments(ctx context.Context, filter GetCommentsFilter) ([]*entity.Comment, error)
	GetURLValidatorConfig(ctx context.Context, validatorType string) (*entity.URLValidatorConfig, error)
	GetAllURLValidatorConfigs(ctx context.Context, out Appender[*entity.URLValidatorConfig]) error
}
