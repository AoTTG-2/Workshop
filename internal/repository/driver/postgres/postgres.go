package postgres

import (
	"context"
	"fmt"
	"workshop/internal/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PGXDriver struct {
	db       *gorm.DB
	migrator *Migrator

	posts              repository.Posts
	moderation         repository.Moderation
	favorites          repository.Favorites
	comments           repository.Comments
	urlValidatorConfig repository.URLValidatorConfig
	votes              repository.Votes
}

func NewDriver(
	_ context.Context,
	connString string,
	migrationsPath string,
	logLevel logger.LogLevel,
) (*PGXDriver, error) {
	d := &PGXDriver{}

	if err := d.initDB(connString, logLevel); err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}

	if migrationsPath != "" {
		if err := d.initMigrator(connString, migrationsPath); err != nil {
			return nil, fmt.Errorf("failed to init migrator: %w", err)
		}
	}

	d.initRepos()

	return d, nil
}

func (d *PGXDriver) Posts() repository.Posts {
	return d.posts
}

func (d *PGXDriver) Votes() repository.Votes {
	return d.votes
}

func (d *PGXDriver) Favorites() repository.Favorites {
	return d.favorites
}

func (d *PGXDriver) Comments() repository.Comments {
	return d.comments
}

func (d *PGXDriver) Moderation() repository.Moderation {
	return d.moderation
}

func (d *PGXDriver) URLValidatorConfig() repository.URLValidatorConfig {
	return d.urlValidatorConfig
}

func (d *PGXDriver) Migrate(_ context.Context) error {
	if d.migrator == nil {
		return fmt.Errorf("migrator is not initialized")
	}

	return d.migrator.Up()
}

func (d *PGXDriver) Drop(_ context.Context) error {
	if d.migrator == nil {
		return fmt.Errorf("migrator is not initialized")
	}

	return d.migrator.Drop()
}

func (d *PGXDriver) Truncate(ctx context.Context, tables []string) error {
	for _, table := range tables {
		if err := d.db.WithContext(ctx).Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}
	return nil
}

func (d *PGXDriver) Close() {
	if d.migrator != nil {
		d.migrator.Close()
	}
	sqlDB, err := d.db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}

func (d *PGXDriver) initDB(connString string, logLevel logger.LogLevel) (err error) {
	d.db, err = gorm.Open(postgres.Open(connString), &gorm.Config{
		PrepareStmt:    true,
		TranslateError: true,
	})
	if err != nil {
		return fmt.Errorf("failed to open gorm connection: %w", err)
	}

	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql db from gorm: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	d.db.Logger = d.db.Logger.LogMode(logLevel)

	return nil
}

func (d *PGXDriver) initRepos() {
	d.posts = NewPostsRepository(d.db)
	d.votes = NewVotesRepository(d.db)
	d.favorites = NewFavoritesRepository(d.db)
	d.comments = NewCommentsRepository(d.db)
	d.moderation = NewModerationRepository(d.db)
	d.urlValidatorConfig = NewURLValidatorConfigRepository(d.db)
}

func (d *PGXDriver) initMigrator(connString, migrationsPath string) (err error) {
	d.migrator, err = NewMigrator(
		fmt.Sprintf("file://%s", migrationsPath),
		connString,
	)

	return err
}

var _ repository.Driver = (*PGXDriver)(nil)
