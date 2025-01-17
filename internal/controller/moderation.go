package controller

import "workshop/internal/types"

type ModeratePostRequest struct {
	PostID types.PostID               `json:"-" param:"postID" validate:"required"`
	Action types.ModerationActionType `json:"action" validate:"required,oneof=approve decline" extensions:"x-order=0"`
	Note   string                     `json:"note" extensions:"x-order=1"`
	UserID types.UserID               `json:"-"`
}

type GetModerationActionsRequest struct {
	PostID      types.PostID               `json:"-" query:"postID"`
	ModeratorID types.UserID               `json:"-" query:"moderator_id"`
	Action      types.ModerationActionType `json:"-" query:"action"`
	SortOrder   SortOrder                  `json:"-" query:"sort_order"`
	IncludePost bool                       `json:"-" query:"include_post"`
	Limit       uint                       `json:"-" query:"limit"`
	Page        uint                       `json:"-" query:"page"`
}
