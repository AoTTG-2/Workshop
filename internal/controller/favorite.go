package controller

import "workshop/internal/types"

type FavoritePostRequest struct {
	PostID types.PostID `json:"-" param:"postID" validate:"required"`
	UserID types.UserID `json:"-"`
}

type UnfavoritePostRequest struct {
	PostID types.PostID `json:"-" param:"postID" validate:"required"`
	UserID types.UserID `json:"-"`
}
