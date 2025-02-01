package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"time"
	"workshop/internal/repository"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"
	"workshop/internal/repository/migrator"
	"workshop/internal/types"
)

const (
	PostsTableName               = "posts"
	TagsTableName                = "tags"
	PostContentsTableName        = "post_contents"
	ModerationActionsTableName   = "moderation_actions"
	CommentsTableName            = "comments"
	VotesTableName               = "votes"
	PostTagsTagsTableName        = "post_tags"
	FavoritesTableName           = "favorites"
	URLValidatorConfigsTableName = "url_validator_configs"
)

type GORMDriver struct {
	db       *gorm.DB
	migrator *migrator.PostgresMigrator
}

func NewGORMDriver(
	connString string,
	migrationsPath string,
) (*GORMDriver, error) {
	d := &GORMDriver{}
	var err error

	d.db, err = gorm.Open(postgres.Open(connString), &gorm.Config{
		PrepareStmt:    true,
		TranslateError: true,
	})

	sqlDB, err := d.db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	d.db.Logger = d.db.Logger.LogMode(logger.Info)
	if migrationsPath != "" {
		if d.migrator, err = migrator.NewPostgresMigrator(
			fmt.Sprintf("file://%s", migrationsPath),
			connString,
		); err != nil {
			return nil, fmt.Errorf("failed to create migrator: %w", err)
		}
	}

	return d, nil
}

func (d *GORMDriver) Migrate(_ context.Context) error {
	if d.migrator == nil {
		return fmt.Errorf("migrator is not initialized")
	}
	return d.migrator.Up()
}

func (d *GORMDriver) Drop(_ context.Context) error {
	if d.migrator == nil {
		return fmt.Errorf("migrator is not initialized")
	}
	return d.migrator.Drop()
}

func (d *GORMDriver) Truncate(ctx context.Context, tables []string) error {
	for _, table := range tables {
		if err := d.db.WithContext(ctx).Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			return err
		}
	}
	return nil
}

func (d *GORMDriver) Close() {
	sqlDB, err := d.db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}

func (d *GORMDriver) CreatePostWithContentsAndTags(ctx context.Context, p *entity.Post) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (d *GORMDriver) UpdatePost(ctx context.Context, p *entity.Post) error {
	res := d.db.WithContext(ctx).Model(p).Select("Title", "Description", "PreviewURL").Updates(p)
	if res.Error != nil {
		return fmt.Errorf("failed to update post: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}

	return nil
}

func (d *GORMDriver) DeletePost(ctx context.Context, postID types.PostID, hard bool) error {
	if !hard {
		tx := d.db.WithContext(ctx).Delete(&entity.Post{ID: postID})
		if tx.Error != nil {
			return fmt.Errorf("failed to exec: %w", tx.Error)
		}
		if tx.RowsAffected == 0 {
			return repoErrors.ErrNotFound
		}
		return nil
	}

	tx := d.db.WithContext(ctx).Unscoped().Delete(&entity.Post{ID: postID})
	if tx.Error != nil {
		return fmt.Errorf("failed to exec: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}
	return nil
}

func (d *GORMDriver) RestorePost(ctx context.Context, postID types.PostID) error {
	ppp := &entity.Post{}
	_ = d.db.WithContext(ctx).Unscoped().Model(&entity.Post{}).Where("id", postID).First(ppp).Error
	res := d.db.WithContext(ctx).Unscoped().Model(&entity.Post{}).Where("id", postID).Update("DeletedAt", nil)
	if err := res.Error; err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}
	if res.RowsAffected == 0 {
		return repoErrors.ErrNotFound
	}
	return nil
}

func (d *GORMDriver) PurgeSoftDeletedPosts(ctx context.Context, olderThan time.Time) (int, error) {
	res := d.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL AND deleted_at <?", olderThan).Delete(&entity.Post{})
	if err := res.Error; err != nil {
		return 0, fmt.Errorf("failed to exec: %w", err)
	}
	return int(res.RowsAffected), nil
}

func (d *GORMDriver) GetPost(ctx context.Context, filter repository.GetPostFilter) (*entity.Post, error) {
	var post entity.Post

	tx := d.db.WithContext(ctx).
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

	if err := tx.First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repoErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find post: %w", err)
	}
	return &post, nil
}

func (d *GORMDriver) GetPosts(ctx context.Context, filter repository.GetPostsFilter) ([]*entity.Post, error) {
	tx := d.db.WithContext(ctx).Model(&entity.Post{})

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
		if filter.RatingOrder == repository.OrderDesc {
			tx = tx.Order("rating DESC")
		} else if filter.RatingOrder == repository.OrderAsc {
			tx = tx.Order("rating ASC")
		}
	} else if filter.CommentsCountOrder > 0 {
		if filter.CommentsCountOrder == repository.OrderDesc {
			tx = tx.Order("comments_count DESC")
		} else if filter.CommentsCountOrder == repository.OrderAsc {
			tx = tx.Order("comments_count ASC")
		}
	} else if filter.FavoritesCountOrder > 0 {
		if filter.FavoritesCountOrder == repository.OrderDesc {
			tx = tx.Order("favorites_count DESC")
		} else if filter.FavoritesCountOrder == repository.OrderAsc {
			tx = tx.Order("favorites_count ASC")
		}
	}

	if filter.UpdatedAtOrder == repository.OrderDesc {
		tx = tx.Order("updated_at DESC")
	} else if filter.UpdatedAtOrder == repository.OrderAsc {
		tx = tx.Order("updated_at ASC")
	}

	if filter.CreatedAtOrder == repository.OrderDesc {
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

func (d *GORMDriver) CreateModerationAction(ctx context.Context, moderationAction *entity.ModerationAction) error {
	if moderationAction.ModeratorID == "" {
		return fmt.Errorf("%w: moderator PostID is required", repoErrors.ErrInvalidData)
	}
	if moderationAction.PostID == 0 {
		return fmt.Errorf("%w: post PostID is required", repoErrors.ErrInvalidData)
	}
	if moderationAction.Action == "" {
		return fmt.Errorf("%w: moderation action is required", repoErrors.ErrInvalidData)
	}

	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (d *GORMDriver) GetModerationActions(ctx context.Context, filter repository.GetModerationActionsFilter) ([]*entity.ModerationAction, error) {
	tx := d.db.WithContext(ctx).Model(&entity.ModerationAction{})

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

	if filter.CreatedAtOrder == repository.OrderDesc {
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

func (d *GORMDriver) AddPostToFavorites(ctx context.Context, favorite *entity.Favorite) error {
	if err := d.db.WithContext(ctx).Create(favorite).Error; err != nil {
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

func (d *GORMDriver) RemovePostFromFavoritesByPostAndUser(ctx context.Context, favorite *entity.Favorite) error {
	res := d.db.WithContext(ctx).Where("post_id = ?", favorite.PostID).Where("user_id = ?", favorite.UserID).Delete(favorite)
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

func (d *GORMDriver) RatePost(ctx context.Context, vote *entity.Vote) error {
	if vote.PostID == 0 || vote.VoterID == "" || vote.Vote == 0 {
		return fmt.Errorf("%w: post PostID, user PostID, and vote value are required to rate a post", repoErrors.ErrInvalidData)
	}

	err := d.db.WithContext(ctx).
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

func (d *GORMDriver) RemovePostRateByPostAndUser(ctx context.Context, vote *entity.Vote) error {
	if vote.PostID == 0 || vote.VoterID == "" {
		return fmt.Errorf("%w: post PostID and user PostID is required to remove post rate", repoErrors.ErrInvalidData)
	}

	res := d.db.WithContext(ctx).Where("post_id = ?", vote.PostID).Where("voter_id = ?", vote.VoterID).Delete(vote)
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

func (d *GORMDriver) AddComment(ctx context.Context, comment *entity.Comment) error {
	if comment.PostID == 0 || comment.AuthorID == "" {
		return fmt.Errorf("%w: post PostID and user PostID is required to add a comment", repoErrors.ErrInvalidData)
	}

	if err := d.db.WithContext(ctx).Create(comment).Error; err != nil {
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

func (d *GORMDriver) UpdateComment(ctx context.Context, comment *entity.Comment) error {
	res := d.db.WithContext(ctx).Model(comment).Select("Content").Updates(comment)
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

func (d *GORMDriver) DeleteComment(ctx context.Context, comment *entity.Comment) error {
	res := d.db.WithContext(ctx).Delete(comment)
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

func (d *GORMDriver) GetCommentByID(ctx context.Context, id types.CommentID) (*entity.Comment, error) {
	comment := &entity.Comment{}
	if err := d.db.WithContext(ctx).First(comment, id).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, repoErrors.ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get comment by PostID: %w", err)
		}
	}

	return comment, nil
}

func (d *GORMDriver) GetComments(ctx context.Context, filter repository.GetCommentsFilter) ([]*entity.Comment, error) {
	tx := d.db.WithContext(ctx).Model(&entity.Comment{})

	if filter.AuthorID != "" {
		tx = tx.Where("author_id = ?", filter.AuthorID)
	}

	if filter.PostID > 0 {
		tx = tx.Where("post_id = ?", filter.PostID)
	}

	if filter.CreatedAtOrder == repository.OrderDesc {
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

func (d *GORMDriver) GetURLValidatorConfig(ctx context.Context, validatorType string) (*entity.URLValidatorConfig, error) {
	config := &entity.URLValidatorConfig{}
	if err := d.db.WithContext(ctx).Where("type = ?", validatorType).First(config).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, repoErrors.ErrNotFound
		default:
			return nil, err
		}
	}
	return config, nil
}

func (d *GORMDriver) GetAllURLValidatorConfigs(ctx context.Context, out repository.Appender[*entity.URLValidatorConfig]) error {
	rows, err := d.db.WithContext(ctx).Model(&entity.URLValidatorConfig{}).Rows()
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
		if err := d.db.ScanRows(rows, cfg); err != nil {
			return err
		}
		out.Append(cfg)
	}
	return nil
}
