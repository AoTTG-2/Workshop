package entity

import "workshop/internal/types"

type Tag struct {
	ID    types.TagID `json:"id"`
	Name  string      `json:"name"`
	Posts []*Post     `json:"posts,omitempty" gorm:"many2many:post_tags;"`
}

type PostTags struct {
	PostID types.PostID `json:"post_id"`
	TagID  types.TagID  `json:"tag_id"`
}
