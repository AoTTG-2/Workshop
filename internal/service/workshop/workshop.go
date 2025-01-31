package workshop

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"time"
	"workshop/internal/controller"
	"workshop/internal/repository"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"
)

type Workshop struct {
	repo    repository.Repository
	limiter Limiter
}

func New(cfg *Config, repo repository.Repository, limiter Limiter) (*Workshop, error) {
	limiter.RegisterGroup(
		PostsLimitKey,
		cfg.PostsLimit,
	)
	limiter.RegisterGroup(
		CommentsLimitKey,
		cfg.CommentsLimit,
	)

	return &Workshop{
		repo:    repo,
		limiter: limiter,
	}, nil
}

func (w *Workshop) GetPosts(ctx context.Context, req *controller.GetPostsRequest) ([]*Post, error) {
	filter := repository.GetPostsFilter{
		BaseFilter: repository.BaseFilter{
			Limit:  int(req.Limit),
			Offset: int((req.Page - 1) * req.Limit),
		},
		Query:               req.SearchQuery,
		AuthorID:            req.AuthorID,
		PostType:            req.Type,
		Tags:                req.Tags,
		OnlyFavorites:       req.OnlyFavorites,
		RatingFilter:        req.RatingFilter,
		OnlyApproved:        req.OnlyApproved,
		IncludePostContents: false,
		IncludeTags:         true,
		ShowDeclined:        req.ShowDeclined,
	}

	sortOrder := repository.Order(0)
	switch req.SortOrder {
	case controller.SortOrderAscending:
		sortOrder = repository.OrderAsc
	case controller.SortOrderDescending:
		sortOrder = repository.OrderDesc
	}

	switch req.SortType {
	case controller.PostsSortTypePopularity:
		filter.FavoritesCountOrder = sortOrder
	case controller.PostsSortTypeBestRated:
		filter.RatingOrder = sortOrder
	case controller.PostsSortTypeNewest:
		filter.CreatedAtOrder = sortOrder
	case controller.PostsSortTypeRecentlyUpdated:
		filter.UpdatedAtOrder = sortOrder
	case controller.PostSortTypeMostDiscussed:
		filter.UpdatedAtOrder = sortOrder
	}

	posts, err := w.repo.GetPosts(ctx, filter)
	if err != nil {
		return nil, err
	}

	res := make([]*Post, len(posts))
	for i, post := range posts {
		res[i] = presentPost(post)
	}
	return res, nil
}

func (w *Workshop) GetPost(ctx context.Context, req *controller.GetPostRequest) (*Post, error) {
	filter := repository.GetPostFilter{
		PostID:              req.ID,
		ForUserID:           req.ForUserID,
		IncludePostContents: true,
		IncludeTags:         true,
		ShowDeclined:        false,
	}

	post, err := w.repo.GetPost(ctx, filter)
	if err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return nil, ErrPostNotFound
		default:
			return nil, err
		}
	}

	return presentPost(post), nil
}

func (w *Workshop) CreatePost(ctx context.Context, req *controller.CreatePostRequest) (*Post, error) {
	info, err := w.limiter.Check(ctx, PostsLimitKey, string(req.UserID))
	if err != nil {
		return nil, fmt.Errorf("check limit error: %w", err)
	}
	if info.Remaining == 0 {
		return nil, &RateLimitExceededError{
			Info: info,
		}
	}

	post := &entity.Post{
		AuthorID:         req.UserID,
		Title:            req.Title,
		Description:      req.Description,
		PreviewURL:       req.PreviewURL,
		PostType:         req.Type,
		Contents:         nil,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletedAt:        gorm.DeletedAt{},
		LastModerationID: nil,
		LastModeration:   nil,
		Rating:           0,
		CommentsCount:    0,
		FavoritesCount:   0,
	}

	for _, tag := range req.Tags {
		post.Tags = append(post.Tags, &entity.Tag{
			Name: tag,
		})
	}

	for _, content := range req.Contents {
		post.Contents = append(post.Contents, &entity.PostContent{
			ContentType: content.Type,
			ContentData: content.Data,
			IsLink:      content.IsLink,
		})
	}

	if err := w.repo.CreatePostWithContentsAndTags(ctx, post); err != nil {
		return nil, err
	}

	if err := w.limiter.TriggerIncrease(ctx, PostsLimitKey, string(req.UserID)); err != nil {
		log.Error().
			Str("user_id", string(req.UserID)).
			Err(err).
			Msg("Trigger increase limit error")
	}

	return presentPost(post), nil
}

func (w *Workshop) UpdatePost(ctx context.Context, req *controller.UpdatePostRequest) (*Post, error) {
	post, err := w.repo.GetPost(ctx, repository.GetPostFilter{PostID: req.PostID, ShowDeclined: true})
	if err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return nil, ErrPostNotFound
		default:
			return nil, err
		}
	}

	if post.AuthorID != req.UserID {
		return nil, ErrPostNotOwned
	}

	post.Title = req.Title
	post.Description = req.Description
	post.PreviewURL = req.PreviewURL

	if err := w.repo.UpdatePost(ctx, post); err != nil {
		return nil, err
	}

	return presentPost(post), nil
}

func (w *Workshop) DeletePost(ctx context.Context, req *controller.DeletePostRequest) error {
	post, err := w.repo.GetPost(ctx, repository.GetPostFilter{PostID: req.PostID, ShowDeclined: true})
	if err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return ErrPostNotFound
		default:
			return err
		}
	}

	if post.AuthorID != req.UserID {
		return ErrPostNotOwned
	}

	if err := w.repo.DeletePost(ctx, req.PostID, true); err != nil {
		return err
	}

	return nil
}

func (w *Workshop) FavoritePost(ctx context.Context, req *controller.FavoritePostRequest) error {
	if err := w.repo.AddPostToFavorites(ctx, &entity.Favorite{
		PostID: req.PostID,
		UserID: req.UserID,
	}); err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrAlreadyExists):
			return ErrPostAlreadyFavorite
		case errors.Is(err, repoErrors.ErrNotFound):
			return ErrPostNotFound
		default:
			return err
		}
	}

	return nil
}

func (w *Workshop) UnfavoritePost(ctx context.Context, req *controller.UnfavoritePostRequest) error {
	if err := w.repo.RemovePostFromFavoritesByPostAndUser(ctx, &entity.Favorite{
		PostID: req.PostID,
		UserID: req.UserID,
	}); err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return ErrPostNotFavorite
		default:
			return err
		}
	}

	return nil
}

func (w *Workshop) ModeratePost(ctx context.Context, req *controller.ModeratePostRequest) error {
	if err := w.repo.CreateModerationAction(ctx, &entity.ModerationAction{
		PostID:      req.PostID,
		ModeratorID: req.UserID,
		Action:      req.Action,
		Note:        req.Note,
	}); err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return ErrPostNotFound
		default:
			return err
		}
	}

	return nil
}

func (w *Workshop) GetModerationActions(ctx context.Context, req *controller.GetModerationActionsRequest) ([]*ModerationAction, error) {
	filter := repository.GetModerationActionsFilter{
		BaseFilter: repository.BaseFilter{
			Limit:  int(req.Limit),
			Offset: int((req.Page - 1) * req.Limit),
		},
		ModeratorID: req.ModeratorID,
		PostID:      req.PostID,
		Action:      req.Action,
		IncludePost: req.IncludePost,
	}

	if req.SortOrder == controller.SortOrderAscending {
		filter.CreatedAtOrder = repository.OrderAsc
	} else {
		filter.CreatedAtOrder = repository.OrderDesc
	}

	moderationActions, err := w.repo.GetModerationActions(ctx, filter)
	if err != nil {
		return nil, err
	}

	res := make([]*ModerationAction, len(moderationActions))
	for i, moderationAction := range moderationActions {
		res[i] = presentModerationAction(moderationAction)
	}

	return res, nil
}

func (w *Workshop) RatePost(ctx context.Context, req *controller.RatePostRequest) error {
	if req.Rating == controller.RateActionRetract {
		if err := w.repo.RemovePostRateByPostAndUser(ctx, &entity.Vote{
			PostID:  req.PostID,
			VoterID: req.UserID,
		}); err != nil {
			switch {
			case errors.Is(err, repoErrors.ErrNotFound):
				return ErrPostNotRated
			default:
				return err
			}
		}
	}

	vote := &entity.Vote{
		PostID:  req.PostID,
		VoterID: req.UserID,
	}
	switch req.Rating {
	case controller.RateActionUpvote:
		vote.Vote = 1
	case controller.RateActionDownvote:
		vote.Vote = -1
	}

	if err := w.repo.RatePost(ctx, vote); err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return ErrPostNotFound
		default:
			return err
		}
	}

	return nil
}

func (w *Workshop) AddComment(ctx context.Context, req *controller.AddCommentRequest) (*Comment, error) {
	info, err := w.limiter.Check(ctx, CommentsLimitKey, string(req.UserID))
	if err != nil {
		return nil, fmt.Errorf("check limit error: %w", err)
	}
	if info.Remaining == 0 {
		return nil, &RateLimitExceededError{
			Info: info,
		}
	}

	comment := &entity.Comment{
		PostID:   req.PostID,
		AuthorID: req.UserID,
		Content:  req.Content,
	}

	if err := w.repo.AddComment(ctx, comment); err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return nil, ErrPostNotFound
		default:
			return nil, err
		}
	}

	if err := w.limiter.TriggerIncrease(ctx, CommentsLimitKey, string(req.UserID)); err != nil {
		log.Error().
			Str("user_id", string(req.UserID)).
			Err(err).
			Msg("Trigger increase limit error")
	}

	return presentComment(comment), nil
}

func (w *Workshop) UpdateComment(ctx context.Context, req *controller.UpdateCommentRequest) (*Comment, error) {
	comment, err := w.repo.GetCommentByID(ctx, req.CommentID)
	if err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return nil, ErrCommentNotFound
		default:
			return nil, err
		}
	}

	if comment.AuthorID != req.UserID {
		return nil, ErrCommentNotOwned
	}

	comment.Content = req.Content

	if err := w.repo.UpdateComment(ctx, comment); err != nil {
		return nil, err
	}

	return presentComment(comment), nil
}

func (w *Workshop) DeleteComment(ctx context.Context, req *controller.DeleteCommentRequest) error {
	comment, err := w.repo.GetCommentByID(ctx, req.CommentID)
	if err != nil {
		switch {
		case errors.Is(err, repoErrors.ErrNotFound):
			return ErrCommentNotFound
		default:
			return err
		}
	}

	if comment.AuthorID != req.UserID {
		return ErrCommentNotOwned
	}

	if err := w.repo.DeleteComment(ctx, comment); err != nil {
		return err
	}

	return nil
}

func (w *Workshop) GetComments(ctx context.Context, req *controller.GetCommentsRequest) ([]*Comment, error) {
	filter := repository.GetCommentsFilter{
		BaseFilter: repository.BaseFilter{
			Limit:  int(req.Limit),
			Offset: int((req.Page - 1) * req.Limit),
		},
		AuthorID: req.AuthorID,
		PostID:   req.PostID,
	}

	if req.SortOrder == controller.SortOrderAscending {
		filter.CreatedAtOrder = repository.OrderAsc
	} else {
		filter.CreatedAtOrder = repository.OrderDesc
	}

	comments, err := w.repo.GetComments(ctx, filter)
	if err != nil {
		return nil, err
	}

	res := make([]*Comment, len(comments))
	for i, comment := range comments {
		res[i] = presentComment(comment)
	}
	return res, nil
}
