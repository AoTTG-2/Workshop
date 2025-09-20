package repository

import (
	"context"
	"time"
	"workshop/internal/repository/common"
	"workshop/internal/repository/entity"
	"workshop/internal/types"
)

type Driver interface {
	Posts() Posts
	Votes() Votes
	Favorites() Favorites
	Comments() Comments
	Moderation() Moderation
	URLValidatorConfig() URLValidatorConfig

	Migrate(ctx context.Context) error
	Drop(ctx context.Context) error
	Truncate(ctx context.Context, tables []string) error
	Close()
}

type Posts interface {
	Create(context.Context, *entity.Post) error
	Update(context.Context, *entity.Post) error
	Delete(ctx context.Context, postID types.PostID, hard bool) error
	Restore(ctx context.Context, postID types.PostID) error
	PurgeSoftDeleted(ctx context.Context, olderThan time.Time) (int, error)
	Get(context.Context, common.GetPostFilter) (*entity.Post, error)
	GetList(context.Context, common.GetPostListFilter) ([]*entity.Post, error)
}

type Moderation interface {
	Create(ctx context.Context, moderationAction *entity.ModerationAction) error
	GetList(ctx context.Context, filter common.GetModerationListFilter) ([]*entity.ModerationAction, error)
}

type Favorites interface {
	Create(context.Context, *entity.Favorite) error
	Delete(context.Context, *entity.Favorite) error
}

type Comments interface {
	Create(context.Context, *entity.Comment) error
	Update(context.Context, *entity.Comment) error
	Delete(context.Context, *entity.Comment) error
	Get(context.Context, types.CommentID) (*entity.Comment, error)
	GetList(context.Context, common.GetCommentsListFilter) ([]*entity.Comment, error)
}

type URLValidatorConfig interface {
	Get(ctx context.Context, validatorType string) (*entity.URLValidatorConfig, error)
	GetList(ctx context.Context, out common.Appender[*entity.URLValidatorConfig]) error
}

type Votes interface {
	Create(context.Context, *entity.Vote) error
	Delete(context.Context, *entity.Vote) error
}
