package common

import "workshop/internal/types"

type BaseFilter struct {
	Limit  int
	Offset int
}

type Order int

const (
	OrderAsc Order = iota + 1
	OrderDesc
)

type GetPostListFilter struct {
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

type GetModerationListFilter struct {
	BaseFilter
	ModeratorID    types.UserID
	PostID         types.PostID
	Action         types.ModerationActionType
	CreatedAtOrder Order
	IncludePost    bool
}

type GetCommentsListFilter struct {
	BaseFilter
	AuthorID       types.UserID
	PostID         types.PostID
	CreatedAtOrder Order
}
