package postgres

import (
	"context"
	"errors"
	"workshop/internal/repository/common"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type URLValidatorConfigRepository struct {
	db *gorm.DB
}

func NewURLValidatorConfigRepository(db *gorm.DB) *URLValidatorConfigRepository {
	return &URLValidatorConfigRepository{db: db}
}

func (r *URLValidatorConfigRepository) Get(ctx context.Context, validatorType string) (*entity.URLValidatorConfig, error) {
	config := &entity.URLValidatorConfig{}
	if err := r.db.WithContext(ctx).Where("type = ?", validatorType).First(config).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, repoErrors.ErrNotFound
		default:
			return nil, err
		}
	}
	return config, nil
}

func (r *URLValidatorConfigRepository) GetList(ctx context.Context, out common.Appender[*entity.URLValidatorConfig]) error {
	rows, err := r.db.WithContext(ctx).Model(&entity.URLValidatorConfig{}).Rows()
	if err != nil {
		return err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()

	for rows.Next() {
		cfg := &entity.URLValidatorConfig{}
		if err := r.db.ScanRows(rows, cfg); err != nil {
			return err
		}
		out.Append(cfg)
	}
	return nil
}
