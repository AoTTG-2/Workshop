package controller

import "workshop/internal/types"

type GetCommentsRequest struct {
	PostID    types.PostID `json:"-" query:"postID" validate:"required"`
	AuthorID  types.UserID `json:"-" query:"author_id"`
	Limit     uint         `json:"-" query:"limit"`
	Page      uint         `json:"-" query:"page"`
	SortOrder SortOrder    `json:"-" query:"sort_order"`
}

type AddCommentRequest struct {
	PostID  types.PostID `json:"postID" validate:"required"`
	Content string       `json:"content" validate:"required,min=1,max=4096"`
	UserID  types.UserID `json:"-"`
}

type UpdateCommentRequest struct {
	CommentID types.CommentID `json:"-" param:"commentID" validate:"required"`
	Content   string          `json:"content" validate:"required,min=1,max=4096"`
	UserID    types.UserID    `json:"-"`
}

type DeleteCommentRequest struct {
	CommentID types.CommentID `json:"-" param:"commentID" validate:"required"`
	UserID    types.UserID    `json:"-"`
}
