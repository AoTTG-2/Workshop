package postgres

import (
	"context"
	"errors"
	"fmt"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"

	"gorm.io/gorm"
)

type FavoritesRepository struct {
	db *gorm.DB
}

func NewFavoritesRepository(db *gorm.DB) *FavoritesRepository {
	return &FavoritesRepository{db: db}
}

func (r *FavoritesRepository) Create(ctx context.Context, favorite *entity.Favorite) error {
	if err := r.db.WithContext(ctx).Create(favorite).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return repoErrors.ErrAlreadyExists
		case errors.Is(err, gorm.ErrForeignKeyViolated):
			fallthrough
		case errors.Is(err, gorm.ErrRecordNotFound):
			return repoErrors.ErrNotFound
		default:
			return fmt.Errorf("failed to add post to favorites: %w", err)
		}
	}
	return nil
}

func (r *FavoritesRepository) Delete(ctx context.Context, favorite *entity.Favorite) error {
	res := r.db.WithContext(ctx).Where("post_id = ?", favorite.PostID).Where("user_id = ?", favorite.UserID).Delete(favorite)
	if err := res.Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return repoErrors.ErrNotFound
		default:
			return fmt.Errorf("failed to remove post from favorites: %w", err)
		}
	}
	if res.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}
	return nil
}
