package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"
	"workshop/internal/repository/common"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"
	"workshop/internal/types"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostsRepository struct {
	db *gorm.DB
}

func NewPostsRepository(db *gorm.DB) *PostsRepository {
	return &PostsRepository{db: db}
}

func (r *PostsRepository) Create(ctx context.Context, p *entity.Post) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, tag := range p.Tags {
			if tag.ID == 0 && tag.Name != "" {
				if err := tx.
					Where("name = ?", tag.Name).
					FirstOrCreate(&p.Tags[i]).
					Error; err != nil {
					return fmt.Errorf("failed to find or create tag: %w", err)
				}
			}
		}

		if err := tx.Create(p).Error; err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}

		return nil
	})
}

func (r *PostsRepository) Update(ctx context.Context, p *entity.Post) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.Post
		if err := tx.Where("id = ?", p.ID).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repoErrors.ErrNotFound
			}
			return err
		}

		for i, tag := range p.Tags {
			if tag.ID == 0 && tag.Name != "" {
				if err := tx.
					Where("name = ?", tag.Name).
					FirstOrCreate(&p.Tags[i]).Error; err != nil {
					return fmt.Errorf("failed to find or create tag: %w", err)
				}
			}
		}
		if err := tx.Model(&existing).Association("Tags").Replace(p.Tags); err != nil {
			return fmt.Errorf("failed to update tags: %w", err)
		}

		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"title":       p.Title,
			"description": p.Description,
			"preview_url": p.PreviewURL,
			"post_type":   p.PostType,
			"updated_at":  time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("failed to update post: %w", err)
		}

		var contentIDs []types.PostContentID
		for _, c := range p.Contents {
			if c.ID == 0 {
				c.PostID = existing.ID
				if err := tx.Create(&c).Error; err != nil {
					return fmt.Errorf("failed to create content: %w", err)
				}
				contentIDs = append(contentIDs, c.ID)
			} else {
				contentIDs = append(contentIDs, c.ID)
				res := tx.Model(&c).
					Where("id = ?", c.ID).
					Updates(map[string]interface{}{
						"content_type": c.ContentType,
						"content_data": c.ContentData,
						"is_link":      c.IsLink,
					})

				if res.Error != nil {
					return fmt.Errorf("failed to update content: %w", res.Error)
				}
				if res.RowsAffected == 0 {
					return fmt.Errorf("failed to update content: %w", repoErrors.ErrNotFound)
				}
			}
		}

		if err := tx.
			Where("post_id = ? AND id NOT IN (?)", existing.ID, contentIDs).
			Delete(&entity.PostContent{}).Error; err != nil {
			return fmt.Errorf("failed to delete old contents: %w", err)
		}

		return nil
	})
}

func (r *PostsRepository) Delete(ctx context.Context, postID types.PostID, hard bool) error {
	if !hard {
		tx := r.db.WithContext(ctx).Delete(&entity.Post{ID: postID})
		if tx.Error != nil {
			return fmt.Errorf("failed to exec: %w", tx.Error)
		}
		if tx.RowsAffected == 0 {
			return repoErrors.ErrNotFound
		}
		return nil
	}

	tx := r.db.WithContext(ctx).Unscoped().Delete(&entity.Post{ID: postID})
	if tx.Error != nil {
		return fmt.Errorf("failed to exec: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}
	return nil
}

func (r *PostsRepository) Restore(ctx context.Context, postID types.PostID) error {
	p := &entity.Post{}
	_ = r.db.WithContext(ctx).Unscoped().Model(&entity.Post{}).Where("id", postID).First(p).Error
	res := r.db.WithContext(ctx).Unscoped().Model(&entity.Post{}).Where("id", postID).Update("DeletedAt", nil)
	if err := res.Error; err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}
	if res.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}
	return nil
}

func (r *PostsRepository) PurgeSoftDeleted(ctx context.Context, olderThan time.Time) (int, error) {
	res := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL AND deleted_at <?", olderThan).Delete(&entity.Post{})
	if err := res.Error; err != nil {
		return 0, fmt.Errorf("failed to exec: %w", err)
	}
	return int(res.RowsAffected), nil
}

func (r *PostsRepository) Get(ctx context.Context, filter common.GetPostFilter) (*entity.Post, error) {
	var post entity.Post

	tx := r.db.WithContext(ctx).
		Model(&entity.Post{ID: filter.PostID})

	if !filter.ShowDeclined {
		tx = tx.Joins("LEFT JOIN moderation_actions ma ON ma.id = posts.last_moderation_id").
			Where("ma IS NULL OR ma.action <> ?", types.ModeratorActionTypeDecline)
	}

	if filter.ForUserID != "" {
		tx = tx.Preload("MyFavorite", "user_id = ?", filter.ForUserID).
			Preload("MyVote", "voter_id = ?", filter.ForUserID)
	}

	if filter.IncludeTags {
		tx = tx.Preload("Tags")
	}

	if filter.IncludePostContents {
		tx = tx.Preload("Contents")
	}

	tx = tx.Preload("LastModeration")

	if err := tx.First(&post, filter.PostID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repoErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find post: %w", err)
	}
	return &post, nil
}

func (r *PostsRepository) GetList(ctx context.Context, filter common.GetPostListFilter) ([]*entity.Post, error) {
	tx := r.db.WithContext(ctx).Model(&entity.Post{})

	if filter.OnlyApproved {
		tx = tx.Joins("JOIN moderation_actions ma ON ma.id = posts.last_moderation_id").
			Where("ma.action = ?", types.ModerationActionTypeApprove)
	} else if !filter.ShowDeclined {
		tx = tx.Joins("LEFT JOIN moderation_actions ma ON ma.id = posts.last_moderation_id").
			Where("ma IS NULL OR ma.action <> ?", types.ModeratorActionTypeDecline)
	}

	if len(filter.Tags) > 0 {
		tx = tx.Joins("JOIN post_tags ON posts.id = post_tags.post_id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.name IN ?", filter.Tags)
	}

	if filter.AuthorID != "" {
		tx = tx.Where("author_id = ?", filter.AuthorID)
	}

	if filter.Query != "" {
		tx = tx.Where("search_vector @@ plainto_tsquery('simple', ?)", filter.Query).
			Order(clause.Expr{
				SQL:                "ts_rank(search_vector, plainto_tsquery('simple', ?)) DESC",
				Vars:               []interface{}{filter.Query},
				WithoutParentheses: true,
			})
	}

	if filter.PostType != "" {
		tx = tx.Where("post_type = ?", filter.PostType)
	}

	if filter.ForUserID != "" {
		if filter.OnlyFavorites {
			tx = tx.Joins("JOIN favorites ON posts.id = favorites.post_id").Where("favorites.user_id = ?", filter.ForUserID)
		}

		if filter.RatingFilter == types.RateTypeVoted {
			tx = tx.Joins("JOIN votes ON posts.id = votes.post_id").
				Where("votes.voter_id = ?", filter.ForUserID)
		} else if filter.RatingFilter == types.RateTypeUpvoted {
			tx = tx.Joins("JOIN votes ON posts.id = votes.post_id").
				Where("votes.vote = 1 AND votes.voter_id = ?", filter.ForUserID)
		} else if filter.RatingFilter == types.RateTypeDownvoted {
			tx = tx.Joins("JOIN votes ON posts.id = votes.post_id").
				Where("votes.vote = -1 AND votes.voter_id = ?", filter.ForUserID)
		}

		tx = tx.Preload("MyFavorite", "user_id = ?", filter.ForUserID).
			Preload("MyVote", "voter_id = ?", filter.ForUserID)
	}

	tx = tx.Limit(filter.Limit).Offset(filter.Offset)

	if filter.RatingOrder > 0 {
		if filter.RatingOrder == common.OrderDesc {
			tx = tx.Order("rating DESC")
		} else if filter.RatingOrder == common.OrderAsc {
			tx = tx.Order("rating ASC")
		}
	} else if filter.CommentsCountOrder > 0 {
		if filter.CommentsCountOrder == common.OrderDesc {
			tx = tx.Order("comments_count DESC")
		} else if filter.CommentsCountOrder == common.OrderAsc {
			tx = tx.Order("comments_count ASC")
		}
	} else if filter.FavoritesCountOrder > 0 {
		if filter.FavoritesCountOrder == common.OrderDesc {
			tx = tx.Order("favorites_count DESC")
		} else if filter.FavoritesCountOrder == common.OrderAsc {
			tx = tx.Order("favorites_count ASC")
		}
	}

	if filter.UpdatedAtOrder == common.OrderDesc {
		tx = tx.Order("updated_at DESC")
	} else if filter.UpdatedAtOrder == common.OrderAsc {
		tx = tx.Order("updated_at ASC")
	}

	if filter.CreatedAtOrder == common.OrderDesc {
		tx = tx.Order("created_at DESC")
	} else {
		tx = tx.Order("created_at ASC")
	}

	if filter.IncludeTags {
		tx = tx.Preload("Tags")
	}

	if filter.IncludePostContents {
		tx = tx.Preload("Contents")
	}

	tx = tx.Preload("LastModeration")

	res := make([]*entity.Post, 0)
	if err := tx.Find(&res).Error; err != nil {
		return nil, fmt.Errorf("failed to find posts: %w", err)
	}

	return res, nil
}
