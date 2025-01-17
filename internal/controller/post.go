package controller

import (
	"workshop/internal/types"
)

type PostsSortType string

const (
	PostsSortTypePopularity      PostsSortType = "popularity"
	PostsSortTypeBestRated       PostsSortType = "best_rated"
	PostsSortTypeNewest          PostsSortType = "newest"
	PostsSortTypeRecentlyUpdated PostsSortType = "recently_updated"
	PostSortTypeMostDiscussed    PostsSortType = "most_discussed"
)

type SortOrder string

const (
	SortOrderAscending  SortOrder = "asc"
	SortOrderDescending SortOrder = "desc"
)

type GetPostsRequest struct {
	Page         uint           `json:"-" query:"page" validate:"required,gte=1"`
	Limit        uint           `json:"-" query:"limit" validate:"required,gt=0,lte=100"`
	SearchQuery  string         `json:"-" query:"search_query" validate:"max=255"`
	AuthorID     types.UserID   `json:"-" query:"author_id"`
	OnlyApproved bool           `json:"-" query:"only_approved"`
	ShowDeclined bool           `json:"-" query:"show_declined"`
	Type         types.PostType `json:"-" query:"type"` // TODO: VALIDATION
	Tags         []string       `json:"-" query:"tags" validate:"omitempty,unique,min=1,max=10"`

	ForUserID     types.UserID   `json:"-" query:"for_user_id"`
	OnlyFavorites bool           `json:"-" query:"only_favorites"`
	RatingFilter  types.RateType `json:"-" query:"rating_filter" validate:"omitempty,oneof=voted upvoted downvoted"`

	SortType  PostsSortType `json:"-" query:"sort_type" validate:"omitempty,oneof=popularity best_rated newest recently_updated"`
	SortOrder SortOrder     `json:"-" query:"sort_order" validate:"omitempty,oneof=asc desc"`
}

type GetPostRequest struct {
	ID        types.PostID `json:"-" param:"postID" validate:"required"`
	ForUserID types.UserID `json:"-" query:"for_user_id"`
}

type CreatePostRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255" extensions:"x-order=0"`
	Description string `json:"description" validate:"required,min=1,max=4096" extensions:"x-order=1"`
	PreviewURL  string `json:"preview_url" validate:"omitempty,url" extensions:"x-order=2"`
	// TODO: Discuss Types, validation
	Type types.PostType `json:"type" validate:"required" extensions:"x-order=3"`
	// TODO: Discuss Limit
	Tags     []string `json:"tags" validate:"omitempty,unique,max=10" extensions:"x-order=4"`
	Contents []struct {
		Data string `json:"data" validate:"required" extensions:"x-order=0"`
		// TODO: VALIDATION
		Type   types.ContentType `json:"type" validate:"required" extensions:"x-order=1"`
		IsLink bool              `json:"is_link" extensions:"x-order=2"`
	} `json:"contents" validate:"required,min=1" extensions:"x-order=5"`
	UserID types.UserID `json:"-" extensions:"x-order=6"`
}

type UpdatePostRequest struct {
	PostID      types.PostID `json:"-" param:"postID" validate:"required" extensions:"x-order=0"`
	Title       string       `json:"title" validate:"required,min=1,max=255" extensions:"x-order=1"`
	Description string       `json:"description" validate:"required,min=1,max=4096" extensions:"x-order=2"`
	PreviewURL  string       `json:"preview_url" validate:"omitempty,url" extensions:"x-order=3"`
	UserID      types.UserID `json:"-" extensions:"x-order=4"`
}

type DeletePostRequest struct {
	PostID types.PostID `json:"-" param:"postID" validate:"required"`
	UserID types.UserID `json:"-"`
}
