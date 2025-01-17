package controller

import "workshop/internal/types"

type RateAction string

const (
	RateActionUpvote   RateAction = "upvote"
	RateActionDownvote RateAction = "downvote"
	RateActionRetract  RateAction = "retract"
)

type RatePostRequest struct {
	PostID types.PostID `json:"-" param:"postID" validate:"required"`
	Rating RateAction   `json:"rating" validate:"required,oneof=upvote downvote retract"`
	UserID types.UserID `json:"-"`
}
