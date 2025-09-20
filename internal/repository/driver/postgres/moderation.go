package postgres

import (
	"context"
	"errors"
	"fmt"
	"workshop/internal/repository/common"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"

	"gorm.io/gorm"
)

type ModerationRepository struct {
	db *gorm.DB
}

func NewModerationRepository(db *gorm.DB) *ModerationRepository {
	return &ModerationRepository{db: db}
}

func (r *ModerationRepository) Create(ctx context.Context, moderationAction *entity.ModerationAction) error {
	if moderationAction.ModeratorID == "" {
		return fmt.Errorf("%w: moderator PostID is required", repoErrors.ErrInvalidData)
	}
	if moderationAction.PostID == 0 {
		return fmt.Errorf("%w: post PostID is required", repoErrors.ErrInvalidData)
	}
	if moderationAction.Action == "" {
		return fmt.Errorf("%w: moderation action is required", repoErrors.ErrInvalidData)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(moderationAction).Error; err != nil {
			switch {
			case errors.Is(err, gorm.ErrForeignKeyViolated):
				fallthrough
			case errors.Is(err, gorm.ErrRecordNotFound):
				return errors.Join(repoErrors.ErrNotFound, err)
			default:
				return fmt.Errorf("failed to create moderation action: %w", err)
			}
		}

		if err := tx.Model(&entity.Post{
			ID: moderationAction.PostID,
		}).
			Update("last_moderation_id", moderationAction.ID).Error; err != nil {
			switch {
			case errors.Is(err, gorm.ErrForeignKeyViolated):
				fallthrough
			case errors.Is(err, gorm.ErrRecordNotFound):
				return errors.Join(repoErrors.ErrNotFound, err)
			default:
				return fmt.Errorf("failed to update post last_moderation_id: %w", err)
			}
		}

		return nil
	})
}

func (r *ModerationRepository) GetList(ctx context.Context, filter common.GetModerationListFilter) ([]*entity.ModerationAction, error) {
	tx := r.db.WithContext(ctx).Model(&entity.ModerationAction{})

	if filter.PostID > 0 {
		tx = tx.Where("post_id = ?", filter.PostID)
	}

	if len(filter.ModeratorID) > 0 {
		tx = tx.Where("moderator_id = ?", filter.ModeratorID)
	}

	if len(filter.Action) > 0 {
		tx = tx.Where("action = ?", filter.Action)
	}

	tx = tx.Limit(filter.Limit).Offset(filter.Offset)

	if filter.CreatedAtOrder == common.OrderDesc {
		tx = tx.Order("created_at DESC")
	} else {
		tx = tx.Order("created_at ASC")
	}

	if filter.IncludePost {
		tx = tx.Preload("Post")
	}

	res := make([]*entity.ModerationAction, 0)
	if err := tx.Find(&res).Error; err != nil {
		return nil, fmt.Errorf("failed to find moderation actions: %w", err)
	}

	return res, nil
}
