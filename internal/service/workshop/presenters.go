package workshop

import (
	"workshop/internal/repository/entity"
	"workshop/internal/types"
)

func presentPost(post *entity.Post) *Post {
	p := &Post{
		ID:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		PreviewURL:  post.PreviewURL,
		PostType:    post.PostType,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		ModerationData: PostModerationData{
			Status: "",
			Note:   "",
		},
		InteractionData: PostInteractionData{
			IsFavorite: post.MyFavorite != nil,
		},
		Rating:         post.Rating,
		CommentsCount:  post.CommentsCount,
		FavoritesCount: post.FavoritesCount,
	}

	for _, t := range post.Tags {
		p.Tags = append(p.Tags, t.Name)
	}

	for _, c := range post.Contents {
		p.Contents = append(p.Contents, PostContent{
			ID:          c.ID,
			ContentType: c.ContentType,
			ContentData: c.ContentData,
			IsLink:      c.IsLink,
		})
	}

	switch {
	case post.LastModeration == nil:
		p.ModerationData.Status = types.ModerationStatusPending
	case post.LastModeration.Action == types.ModerationActionTypeApprove:
		p.ModerationData.Status = types.ModerationStatusApproved
	case post.LastModeration.Action == types.ModeratorActionTypeDecline:
		p.ModerationData.Status = types.ModerationStatusDeclined
		p.ModerationData.Note = post.LastModeration.Note
	}

	switch {
	case post.MyVote == nil:
		p.InteractionData.Vote = types.RateTypeNone
	case post.MyVote.Vote == 1:
		p.InteractionData.Vote = types.RateTypeUpvoted
	case post.MyVote.Vote == -1:
		p.InteractionData.Vote = types.RateTypeDownvoted
	}

	return p
}

func presentModerationAction(action *entity.ModerationAction) *ModerationAction {
	a := &ModerationAction{
		ID:          action.ID,
		ModeratorID: action.ModeratorID,
		Action:      action.Action,
		Note:        action.Note,
		CreatedAt:   action.CreatedAt,
	}

	if action.Post != nil {
		a.Post = presentPost(action.Post)
	}
	return a
}

func presentComment(comment *entity.Comment) *Comment {
	c := &Comment{
		ID:        comment.ID,
		AuthorID:  comment.AuthorID,
		PostID:    comment.PostID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	return c
}
