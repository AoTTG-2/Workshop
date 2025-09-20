package workshop

import (
	"context"
	"time"
)

const (
	PostsLimitKey    = "posts_limit"
	CommentsLimitKey = "comments_limit"
)

type RateLimitInfo struct {
	Limit     uint64
	Current   uint64
	Remaining uint64
	ResetAt   time.Time
}

type LimitConfig struct {
	Limit  uint64
	Period time.Duration
}

type Limiter interface {
	RegisterGroup(groupKey string, cfg LimitConfig)
	Check(ctx context.Context, groupKey, userID string) (RateLimitInfo, error)
	TriggerIncrease(ctx context.Context, groupKey, userID string) error
}
