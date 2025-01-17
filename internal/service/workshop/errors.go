package workshop

import "errors"

var (
	ErrPostNotFound        = errors.New("post not found")
	ErrPostNotOwned        = errors.New("post not owned")
	ErrPostAlreadyFavorite = errors.New("post already favorite")
	ErrPostNotFavorite     = errors.New("post not favorite")
	ErrPostNotRated        = errors.New("post not rated")

	ErrCommentNotFound = errors.New("comment not found")
	ErrCommentNotOwned = errors.New("comment not owned")

	ErrLimitExceeded = errors.New("limit exceeded")
)
