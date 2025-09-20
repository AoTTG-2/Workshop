package postgres

import (
	"context"
	"errors"
	"fmt"
	"workshop/internal/repository/common"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"
	"workshop/internal/types"

	"gorm.io/gorm"
)

type CommentsRepository struct {
	db *gorm.DB
}

func NewCommentsRepository(db *gorm.DB) *CommentsRepository {
	return &CommentsRepository{db: db}
}

func (r *CommentsRepository) Create(ctx context.Context, comment *entity.Comment) error {
	if comment.PostID == 0 || comment.AuthorID == "" {
		return fmt.Errorf("%w: post PostID and user PostID is required to add a comment", repoErrors.ErrInvalidData)
	}

	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrForeignKeyViolated):
			fallthrough
		case errors.Is(err, gorm.ErrRecordNotFound):
			return repoErrors.ErrNotFound
		default:
			return fmt.Errorf("failed to add comment: %w", err)
		}
	}
	return nil
}

func (r *CommentsRepository) Update(ctx context.Context, comment *entity.Comment) error {
	res := r.db.WithContext(ctx).Model(comment).Select("Content").Updates(comment)
	if err := res.Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return repoErrors.ErrNotFound
		default:
			return fmt.Errorf("failed to update comment: %w", err)
		}
	}

	if res.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}
	return nil
}

func (r *CommentsRepository) Delete(ctx context.Context, comment *entity.Comment) error {
	res := r.db.WithContext(ctx).Delete(comment)
	if err := res.Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return repoErrors.ErrNotFound
		default:
			return fmt.Errorf("failed to delete comment: %w", err)
		}
	}

	if res.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}
	return nil
}

func (r *CommentsRepository) Get(ctx context.Context, id types.CommentID) (*entity.Comment, error) {
	comment := &entity.Comment{}
	if err := r.db.WithContext(ctx).First(comment, id).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, repoErrors.ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get comment by PostID: %w", err)
		}
	}

	return comment, nil
}

func (r *CommentsRepository) GetList(ctx context.Context, filter common.GetCommentsListFilter) ([]*entity.Comment, error) {
	tx := r.db.WithContext(ctx).Model(&entity.Comment{})

	if filter.AuthorID != "" {
		tx = tx.Where("author_id = ?", filter.AuthorID)
	}

	if filter.PostID > 0 {
		tx = tx.Where("post_id = ?", filter.PostID)
	}

	if filter.CreatedAtOrder == common.OrderDesc {
		tx = tx.Order("created_at DESC")
	} else {
		tx = tx.Order("created_at ASC")
	}

	tx = tx.Limit(filter.Limit).Offset(filter.Offset)

	res := make([]*entity.Comment, 0)
	if err := tx.Find(&res).Error; err != nil {
		return nil, fmt.Errorf("failed to find comments: %w", err)
	}

	return res, nil
}
