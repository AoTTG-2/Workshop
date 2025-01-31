package workshop

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrPostNotFound        = errors.New("post not found")
	ErrPostNotOwned        = errors.New("post not owned")
	ErrPostAlreadyFavorite = errors.New("post already favorite")
	ErrPostNotFavorite     = errors.New("post not favorite")
	ErrPostNotRated        = errors.New("post not rated")

	ErrCommentNotFound = errors.New("comment not found")
	ErrCommentNotOwned = errors.New("comment not owned")
)

type RateLimitExceededError struct {
	Info RateLimitInfo
}

func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf(
		"rate limit exceeded (limit=%d, current=%d, remaining=%d, resetAt=%s)",
		e.Info.Limit, e.Info.Current, e.Info.Remaining, e.Info.ResetAt.Format(time.RFC3339),
	)
}
