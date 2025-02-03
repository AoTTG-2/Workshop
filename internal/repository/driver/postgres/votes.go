package postgres

import (
	"context"
	"errors"
	"fmt"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VotesRepository struct {
	db *gorm.DB
}

func NewVotesRepository(db *gorm.DB) *VotesRepository {
	return &VotesRepository{db: db}
}

func (r *VotesRepository) Create(ctx context.Context, vote *entity.Vote) error {
	if vote.PostID == 0 || vote.VoterID == "" || vote.Vote == 0 {
		return fmt.Errorf("%w: post PostID, user PostID, and vote value are required to rate a post", repoErrors.ErrInvalidData)
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}, {Name: "voter_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"vote"}),
		}).
		Create(vote).Error

	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrForeignKeyViolated):
			fallthrough
		case errors.Is(err, gorm.ErrRecordNotFound):
			return repoErrors.ErrNotFound
		default:
			return fmt.Errorf("failed to rate post: %w", err)
		}
	}

	return nil
}

func (r *VotesRepository) Delete(ctx context.Context, vote *entity.Vote) error {
	if vote.PostID == 0 || vote.VoterID == "" {
		return fmt.Errorf("%w: post PostID and user PostID is required to remove post rate", repoErrors.ErrInvalidData)
	}

	res := r.db.WithContext(ctx).Where("post_id = ?", vote.PostID).Where("voter_id = ?", vote.VoterID).Delete(vote)
	if err := res.Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return repoErrors.ErrNotFound
		default:
			return fmt.Errorf("failed to remove post rate: %w", err)
		}
	}
	if res.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}
	return nil
}
