package postgres

import (
	"context"
	"fmt"
	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/values"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"strings"
	"sync"
	"testing"
	"time"
	"workshop/internal/repository"
	"workshop/internal/repository/entity"
	repoErrors "workshop/internal/repository/errors"
	"workshop/internal/types"
	"workshop/pkg/appender"
)

type testConfig struct {
	ConnString     string `env:"TEST_POSTGRES_CONN_STRING" default:"postgres://postgres:12345@postgres:5432/test_workshop?sslmode=disable"`
	MigrationsPath string `env:"TEST_POSTGRES_MIGRATIONS_PATH" default:"./migrations/postgres"`
}

var (
	testCfg         *testConfig
	testCfgOnce     sync.Once
	testMigrateOnce sync.Once
)

func setupTestDB(t *testing.T) *GORMDriver {
	testCfgOnce.Do(func() {
		testCfg = new(testConfig)
		cfgManager := gocfg.NewDefault()
		if dotEnvProvider, err := values.NewDotEnvProvider(); err == nil {
			cfgManager = cfgManager.AddValueProviders(dotEnvProvider)
		}

		err := cfgManager.Unmarshal(testCfg)
		require.NoError(t, err)
	})

	driver, err := NewGORMDriver(testCfg.ConnString, testCfg.MigrationsPath)
	require.NoError(t, err)

	testMigrateOnce.Do(func() {
		err = driver.Migrate(context.Background())
		require.NoError(t, err)
	})

	return driver
}

func teardownTestDB(t *testing.T, driver *GORMDriver) {
	err := driver.Truncate(context.Background(), []string{
		PostsTableName,
		FavoritesTableName,
		PostContentsTableName,
		VotesTableName,
		ModerationActionsTableName,
		PostTagsTagsTableName,
		CommentsTableName,
		URLValidatorConfigsTableName,
	})
	require.NoError(t, err)

	driver.Close()
}

func TestGORMDriver_CreatePost(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "12345",
		Title:       "Portal 2",
		Description: "Portal 2 Game Mode And Maps",
		PreviewURL:  "urlhere",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Mission"},
			{Name: "Other"},
			{Name: "Non-canonical"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "longurl",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	err := driver.CreatePostWithContentsAndTags(context.Background(), post)
	require.NoError(t, err)

	assert.NotZero(t, post.ID)
	assert.Equal(t, post.Title, "Portal 2")
	assert.Equal(t, post.Description, "Portal 2 Game Mode And Maps")
	assert.Equal(t, post.PreviewURL, "urlhere")
	assert.EqualValues(t, post.PostType, types.PostTypeGameMode)
	assert.Len(t, post.Tags, 3)
	for _, tag := range post.Tags {
		assert.NotZero(t, tag.ID)
	}
	assert.Len(t, post.Contents, 2)
	for _, pc := range post.Contents {
		assert.NotZero(t, pc.ID)
	}

	t.Logf("Created post %+v", post)
}

func TestGORMDriver_UpdatePost(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "12345",
		Title:       "Portal 2",
		Description: "Portal 2 Game Mode And Maps",
		PreviewURL:  "urlhere",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Mission"},
			{Name: "Other"},
			{Name: "Non-canonical"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "longurl",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		post.Title = "Updated Portal 2"
		post.Description = "Updated Portal 2 Game Mode And Maps"
		post.PreviewURL = "updatedurl"
		err := driver.UpdatePost(context.Background(), post)
		require.NoError(t, err)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{
			PostID: post.ID,
		})
		assert.Equal(t, "Updated Portal 2", p.Title)
		assert.Equal(t, "Updated Portal 2 Game Mode And Maps", p.Description)
		assert.Equal(t, "updatedurl", p.PreviewURL)
		assert.WithinDuration(t, p.UpdatedAt, time.Now(), time.Second)
	})

	t.Run("Invalid post PostID", func(t *testing.T) {
		updatedPost := &entity.Post{
			ID:          838383,
			Title:       "Some Title",
			Description: "Some Description",
			PreviewURL:  "someurl",
		}

		err := driver.UpdatePost(context.Background(), updatedPost)
		require.ErrorIs(t, err, repoErrors.ErrNotFound)
	})
}

func TestGORMDriver_DeletePost(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "12345",
		Title:       "Portal 2",
		Description: "Portal 2 Game Mode And Maps",
		PreviewURL:  "urlhere",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Mission"},
			{Name: "Other"},
			{Name: "Non-canonical"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "longurl",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}
	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	err := driver.DeletePost(context.Background(), post.ID, false)
	require.NoError(t, err)

	err = driver.db.Unscoped().Where("id =?", post.ID).First(post).Error
	assert.NoError(t, err)
	assert.True(t, post.DeletedAt.Valid)
	assert.WithinDuration(t, post.DeletedAt.Time, time.Now(), time.Second)

	err = driver.DeletePost(context.Background(), post.ID, true)
	require.NoError(t, err)

	err = driver.db.Unscoped().Where("id = ?", post.ID).First(post).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestGORMDriver_GetPost(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "12345",
		Title:       "Portal 2",
		Description: "Portal 2 Game Mode And Maps",
		PreviewURL:  "urlhere",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Mission"},
			{Name: "Other"},
			{Name: "Non-canonical"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "longurl",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}
	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	retrievedPost, err := driver.GetPost(context.Background(), repository.GetPostFilter{
		PostID:              post.ID,
		IncludePostContents: false,
		IncludeTags:         false,
	})
	assert.NoError(t, err)
	assert.Equal(t, retrievedPost.ID, post.ID)
	assert.Equal(t, retrievedPost.Title, post.Title)
	assert.Equal(t, retrievedPost.Description, post.Description)
	assert.Equal(t, retrievedPost.PreviewURL, post.PreviewURL)
	assert.EqualValues(t, retrievedPost.PostType, post.PostType)
	assert.Len(t, retrievedPost.Tags, 0)
	assert.Len(t, retrievedPost.Contents, 0)
	assert.Nil(t, retrievedPost.LastModeration)

	if err = driver.CreateModerationAction(context.Background(), &entity.ModerationAction{
		PostID:      post.ID,
		ModeratorID: "moderator1",
		Action:      types.ModerationActionTypeApprove,
		Note:        "",
	}); err != nil {
		t.Fatal(err)
	}

	t.Run("With post contents and tags", func(t *testing.T) {
		retrievedPost, err = driver.GetPost(context.Background(), repository.GetPostFilter{
			PostID:              post.ID,
			IncludePostContents: true,
			IncludeTags:         true,
		})
		assert.NoError(t, err)
		assert.Equal(t, retrievedPost.ID, post.ID)
		assert.Equal(t, retrievedPost.Title, post.Title)
		assert.Equal(t, retrievedPost.Description, post.Description)
		assert.Equal(t, retrievedPost.PreviewURL, post.PreviewURL)
		assert.EqualValues(t, retrievedPost.PostType, post.PostType)
		assert.Len(t, retrievedPost.Tags, len(post.Tags))
		assert.Len(t, retrievedPost.Contents, len(post.Contents))
		assert.NotNil(t, retrievedPost.LastModeration)
		assert.EqualValues(t, retrievedPost.LastModeration.ModeratorID, "moderator1")
		assert.EqualValues(t, retrievedPost.LastModeration.Action, types.ModerationActionTypeApprove)
		assert.WithinDuration(t, retrievedPost.LastModeration.CreatedAt, time.Now(), time.Second)
	})

	t.Run("Declined/Approved post", func(t *testing.T) {
		t.Run("Declined post", func(t *testing.T) {
			if err = driver.CreateModerationAction(context.Background(), &entity.ModerationAction{
				PostID:      post.ID,
				ModeratorID: "moderator2",
				Action:      types.ModeratorActionTypeDecline,
				Note:        "Exploit detected",
			}); err != nil {
				t.Fatal(err)
			}

			retrievedPost, err = driver.GetPost(context.Background(), repository.GetPostFilter{
				PostID:              post.ID,
				IncludePostContents: true,
				IncludeTags:         true,
			})
			assert.ErrorIs(t, err, repoErrors.ErrNotFound)
		})

		t.Run("Approved post", func(t *testing.T) {
			if err = driver.CreateModerationAction(context.Background(), &entity.ModerationAction{
				PostID:      post.ID,
				ModeratorID: "moderator1",
				Action:      types.ModerationActionTypeApprove,
				Note:        "",
			}); err != nil {
				t.Fatal(err)
			}

			retrievedPost, err = driver.GetPost(context.Background(), repository.GetPostFilter{
				PostID:              post.ID,
				IncludePostContents: true,
				IncludeTags:         true,
			})
			assert.NoError(t, err)
			assert.Equal(t, retrievedPost.ID, post.ID)
		})
	})

	t.Run("Soft-deleted post", func(t *testing.T) {
		if err = driver.DeletePost(context.Background(), post.ID, false); err != nil {
			t.Fatal(err)
		}

		retrievedPost, err = driver.GetPost(context.Background(), repository.GetPostFilter{
			PostID:              post.ID,
			IncludePostContents: true,
			IncludeTags:         true,
		})
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)
		if err := driver.RestorePost(context.Background(), post.ID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("User interactions", func(t *testing.T) {
		t.Run("Preload favorite", func(t *testing.T) {
			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{
				PostID:       post.ID,
				ForUserID:    "interactionTester1",
				ShowDeclined: true,
			})
			assert.NoError(t, err)
			assert.Nil(t, p.MyFavorite)

			favorite := &entity.Favorite{
				PostID: post.ID,
				UserID: "interactionTester1",
			}
			if err := driver.AddPostToFavorites(context.Background(), favorite); err != nil {
				t.Fatal(err)
			}

			p, err = driver.GetPost(context.Background(), repository.GetPostFilter{
				PostID:       post.ID,
				ForUserID:    "interactionTester1",
				ShowDeclined: true,
			})
			assert.NoError(t, err)
			assert.NotNil(t, p.MyFavorite)
			assert.EqualValues(t, favorite.ID, p.MyFavorite.ID)
		})

		t.Run("Preload vote", func(t *testing.T) {
			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{
				PostID:       post.ID,
				ForUserID:    "interactionTester1",
				ShowDeclined: true,
			})
			assert.NoError(t, err)
			assert.Nil(t, p.MyVote)

			vote := &entity.Vote{
				PostID:  post.ID,
				VoterID: "interactionTester1",
				Vote:    1,
			}
			if err := driver.RatePost(context.Background(), vote); err != nil {
				t.Fatal(err)
			}

			p, err = driver.GetPost(context.Background(), repository.GetPostFilter{
				PostID:       post.ID,
				ForUserID:    "interactionTester1",
				ShowDeclined: true,
			})
			assert.NoError(t, err)
			assert.NotNil(t, p.MyVote)
			assert.EqualValues(t, vote.ID, p.MyVote.ID)
		})
	})
}

func TestGORMDriver_GetPosts(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post1 := &entity.Post{
		AuthorID:    "Jagerente",
		Title:       "Portal 2",
		Description: "Original Game by Valve",
		PreviewURL:  "url1",
		PostType:    types.PostTypeMapAndGameMode,
		Tags: []*entity.Tag{
			{Name: "Singleplayer"},
			{Name: "Mission"},
			{Name: "Misc"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw1",
				IsLink:      false,
			},
		},
	}
	post2 := &entity.Post{
		AuthorID:    "Jagerente",
		Title:       "Counter Strike",
		Description: "Original Game by Valve",
		PreviewURL:  "url2",
		PostType:    types.PostTypeMapSuite,
		Tags: []*entity.Tag{
			{Name: "Multiplayer"},
			{Name: "PVP"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "longurl2",
				IsLink:      true,
			},
		},
	}
	post3 := &entity.Post{
		AuthorID:    "Steve",
		Title:       "Beast Titan",
		Description: "I'm Steve",
		PreviewURL:  "url3",
		PostType:    types.PostTypeMapAndGameMode,
		Tags: []*entity.Tag{
			{Name: "Multiplayer"},
			{Name: "RP"},
			{Name: "Mission"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomAsset,
				ContentData: "longurl3",
				IsLink:      true,
			},
		},
	}

	posts := []*entity.Post{post1, post2, post3}
	for _, post := range posts {
		if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
			t.Fatal(err)
		}
	}

	if err := driver.CreateModerationAction(context.Background(), &entity.ModerationAction{
		PostID:      post1.ID,
		ModeratorID: "moderator1",
		Action:      types.ModerationActionTypeApprove,
		Note:        "",
	}); err != nil {
		t.Fatal(err)
	}

	if err := driver.CreateModerationAction(context.Background(), &entity.ModerationAction{
		PostID:      post2.ID,
		ModeratorID: "moderator2",
		Action:      types.ModeratorActionTypeDecline,
		Note:        "Something went wrong",
	}); err != nil {
		t.Fatal(err)
	}

	t.Run("No filters", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 2)
		assert.Contains(t, mapRes.Map(), post1.ID)
		assert.Contains(t, mapRes.Map(), post3.ID)

		assert.Nil(t, mapRes.Map()[post1.ID].Tags)
		assert.Nil(t, mapRes.Map()[post1.ID].Contents)
		assert.NotNil(t, mapRes.Map()[post1.ID].LastModeration)

		assert.Nil(t, mapRes.Map()[post3.ID].Tags)
		assert.Nil(t, mapRes.Map()[post3.ID].Contents)
		assert.Nil(t, mapRes.Map()[post3.ID].LastModeration)
	})

	t.Run("ShowDeclined", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			ShowDeclined: true,
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 3)
		assert.Contains(t, mapRes.Map(), post1.ID)
		assert.Contains(t, mapRes.Map(), post2.ID)
		assert.Contains(t, mapRes.Map(), post3.ID)
	})

	t.Run("OnlyApproved", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			OnlyApproved: true,
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 1)
		_, ok := mapRes.Map()[post1.ID]
		assert.True(t, ok)
	})

	t.Run("Include Tags", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			IncludeTags: true,
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 2)
		assert.Contains(t, mapRes.Map(), post1.ID)
		assert.Contains(t, mapRes.Map(), post3.ID)

		assert.NotNil(t, mapRes.Map()[post1.ID].Tags)
		assert.Nil(t, mapRes.Map()[post1.ID].Contents)
		assert.NotNil(t, mapRes.Map()[post1.ID].LastModeration)

		assert.NotNil(t, mapRes.Map()[post3.ID].Tags)
		assert.Nil(t, mapRes.Map()[post3.ID].Contents)
		assert.Nil(t, mapRes.Map()[post3.ID].LastModeration)
	})

	t.Run("Include Post Contents", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			IncludePostContents: true,
		})
		assert.NoError(t, err)

		out := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			out.Append(post)
		}

		assert.Len(t, out.Map(), 2)
		assert.Contains(t, out.Map(), post1.ID)
		assert.Contains(t, out.Map(), post3.ID)

		assert.Nil(t, out.Map()[post1.ID].Tags)
		assert.NotNil(t, out.Map()[post1.ID].Contents)
		assert.NotNil(t, out.Map()[post1.ID].LastModeration)

		assert.Nil(t, out.Map()[post3.ID].Tags)
		assert.NotNil(t, out.Map()[post3.ID].Contents)
		assert.Nil(t, out.Map()[post3.ID].LastModeration)
	})

	t.Run("Limit and Offset", func(t *testing.T) {
		testCases := []struct {
			Limit         int
			Offset        int
			ExpectedCount int
		}{
			{Limit: 1, Offset: 0, ExpectedCount: 1},
			{Limit: 2, Offset: 0, ExpectedCount: 2},
			{Limit: 3, Offset: 0, ExpectedCount: 3},
			{Limit: 10, Offset: 0, ExpectedCount: 3},
			{Limit: 1, Offset: 1, ExpectedCount: 1},
			{Limit: 2, Offset: 1, ExpectedCount: 2},
			{Limit: 3, Offset: 1, ExpectedCount: 2},
			{Limit: 10, Offset: 1, ExpectedCount: 2},
			{Limit: 1, Offset: 2, ExpectedCount: 1},
			{Limit: 10, Offset: 2, ExpectedCount: 1},
			{Limit: 10, Offset: 4, ExpectedCount: 0},
		}

		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("Limit: %d, Offset: %d", testCase.Limit, testCase.Offset), func(t *testing.T) {
				res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
					BaseFilter: repository.BaseFilter{
						Limit:  testCase.Limit,
						Offset: testCase.Offset,
					},
					ShowDeclined: true,
				})
				assert.NoError(t, err)

				mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
					return v.ID
				})
				for _, post := range res {
					mapRes.Append(post)
				}

				assert.Len(t, mapRes.Map(), testCase.ExpectedCount)
			})
		}
	})

	t.Run("Search by Author", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			AuthorID:     "Jagerente",
			ShowDeclined: true,
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}
		assert.Len(t, mapRes.Map(), 2)
		assert.Contains(t, mapRes.Map(), post1.ID)
		assert.Contains(t, mapRes.Map(), post2.ID)

		res, err = driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			AuthorID:     "Steve",
			ShowDeclined: true,
		})
		assert.NoError(t, err)

		mapRes = appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 1)
		assert.Contains(t, mapRes.Map(), post3.ID)
	})

	t.Run("Search by Name and Description", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			Query:        strings.ToLower("Valve"),
			ShowDeclined: true,
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 2)
		assert.Contains(t, mapRes.Map(), post1.ID)
		assert.Contains(t, mapRes.Map(), post2.ID)

		res, err = driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			Query:        strings.ToLower("Portal"),
			ShowDeclined: true,
		})
		assert.NoError(t, err)

		mapRes = appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 1)
		assert.Contains(t, mapRes.Map(), post1.ID)

		res, err = driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			Query:        strings.ToLower("Steve"),
			ShowDeclined: true,
		})
		assert.NoError(t, err)

		mapRes = appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 1)
		assert.Contains(t, mapRes.Map(), post3.ID)
	})

	t.Run("Search by Type", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			ShowDeclined: true,
			PostType:     types.PostTypeMapAndGameMode,
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 2)
		assert.Contains(t, mapRes.Map(), post1.ID)
		assert.Contains(t, mapRes.Map(), post3.ID)

		res, err = driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			ShowDeclined: true,
			PostType:     types.PostTypeMapSuite,
		})
		assert.NoError(t, err)

		mapRes = appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 1)
		assert.Contains(t, mapRes.Map(), post2.ID)
	})

	t.Run("Search by Tags", func(t *testing.T) {
		res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			ShowDeclined: true,
			Tags:         []string{"Multiplayer"},
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 2)
		assert.Contains(t, mapRes.Map(), post2.ID)
		assert.Contains(t, mapRes.Map(), post3.ID)

		res, err = driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			ShowDeclined: true,
			Tags:         []string{"Mission"},
		})
		assert.NoError(t, err)

		mapRes = appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 2)
		assert.Contains(t, mapRes.Map(), post1.ID)
		assert.Contains(t, mapRes.Map(), post3.ID)

		res, err = driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			ShowDeclined: true,
			Tags:         []string{"RP"},
		})
		assert.NoError(t, err)

		mapRes = appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 1)
		assert.Contains(t, mapRes.Map(), post3.ID)

		res, err = driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			ShowDeclined: true,
			Tags:         []string{"Multiplayer", "PVP"},
		})
		assert.NoError(t, err)

		mapRes = appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 2)
		assert.Contains(t, mapRes.Map(), post2.ID)
		assert.Contains(t, mapRes.Map(), post3.ID)

		res, err = driver.GetPosts(context.Background(), repository.GetPostsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			ShowDeclined: true,
			Tags:         []string{"Multiplayer", "PVP", "Singleplayer"},
		})
		assert.NoError(t, err)

		mapRes = appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
			return v.ID
		})
		for _, post := range res {
			mapRes.Append(post)
		}

		assert.Len(t, mapRes.Map(), 3)
		assert.Contains(t, mapRes.Map(), post1.ID)
		assert.Contains(t, mapRes.Map(), post2.ID)
		assert.Contains(t, mapRes.Map(), post3.ID)
	})

	t.Run("CreatedAt order", func(t *testing.T) {
		t.Run("Ascending", func(t *testing.T) {
			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined:   true,
				CreatedAtOrder: repository.OrderAsc,
			})
			assert.NoError(t, err)

			assert.Len(t, res, 3)
			for i := 0; i < len(res)-1; i++ {
				assert.True(t, res[i].CreatedAt.Before(res[i+1].CreatedAt))
			}
		})

		t.Run("Descending", func(t *testing.T) {
			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined:   true,
				CreatedAtOrder: repository.OrderDesc,
			})
			assert.NoError(t, err)
			assert.Len(t, res, 3)
			for i := 0; i < len(res)-1; i++ {
				assert.True(t, res[i].CreatedAt.After(res[i+1].CreatedAt))
			}
		})
	})

	t.Run("Rating order", func(t *testing.T) {
		ratings := []struct {
			postID types.PostID
			rating int
		}{
			{
				postID: post1.ID,
				rating: -5,
			},
			{
				postID: post2.ID,
				rating: 10,
			},
			{
				postID: post3.ID,
				rating: -10,
			},
		}

		for _, r := range ratings {
			for i := 0; i < r.rating; i++ {
				voteVal := 1
				if r.rating < 0 {
					voteVal = -1
				}
				vote := &entity.Vote{
					PostID:  r.postID,
					VoterID: types.UserID(fmt.Sprintf("voter%d", i)),
					Vote:    voteVal,
				}
				if err := driver.RatePost(context.Background(), vote); err != nil {
					t.Fatal(err)
				}
			}
		}

		t.Run("Ascending", func(t *testing.T) {
			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined:   true,
				CreatedAtOrder: repository.OrderAsc,
				RatingOrder:    repository.OrderAsc,
			})
			assert.NoError(t, err)

			assert.Len(t, res, 3)
			for i := 0; i < len(res)-1; i++ {
				assert.LessOrEqual(t, res[i].Rating, res[i+1].Rating)
			}
		})

		t.Run("Descending", func(t *testing.T) {
			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined:   true,
				CreatedAtOrder: repository.OrderAsc,
				RatingOrder:    repository.OrderDesc,
			})
			assert.NoError(t, err)

			assert.Len(t, res, 3)
			for i := 0; i < len(res)-1; i++ {
				assert.GreaterOrEqual(t, res[i].Rating, res[i+1].Rating)
			}
		})
	})

	t.Run("Comments order", func(t *testing.T) {
		comments := []struct {
			postID types.PostID
			count  int
		}{
			{
				postID: post1.ID,
				count:  5,
			},
			{
				postID: post2.ID,
				count:  10,
			},
			{
				postID: post3.ID,
				count:  3,
			},
		}

		for _, c := range comments {
			for i := 0; i < c.count; i++ {
				comment := &entity.Comment{
					PostID:   c.postID,
					AuthorID: types.UserID(fmt.Sprintf("commenter%d", i)),
					Content:  fmt.Sprintf("Comment %d", i),
				}
				if err := driver.AddComment(context.Background(), comment); err != nil {
					t.Fatal(err)
				}
			}
		}

		t.Run("Ascending", func(t *testing.T) {
			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined:       true,
				CreatedAtOrder:     repository.OrderAsc,
				CommentsCountOrder: repository.OrderAsc,
			})
			assert.NoError(t, err)

			assert.Len(t, res, 3)
			for i := 0; i < len(res)-1; i++ {
				assert.LessOrEqual(t, res[i].CommentsCount, res[i+1].CommentsCount)
			}
		})

		t.Run("Descending", func(t *testing.T) {
			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined:       true,
				CreatedAtOrder:     repository.OrderAsc,
				CommentsCountOrder: repository.OrderDesc,
			})
			assert.NoError(t, err)

			assert.Len(t, res, 3)
			for i := 0; i < len(res)-1; i++ {
				assert.GreaterOrEqual(t, res[i].CommentsCount, res[i+1].CommentsCount)
			}
		})
	})

	t.Run("User interactions", func(t *testing.T) {
		t.Run("Preload favorite", func(t *testing.T) {
			favorite := &entity.Favorite{
				PostID: post1.ID,
				UserID: "interactionTester1",
			}
			if err := driver.AddPostToFavorites(context.Background(), favorite); err != nil {
				t.Fatal(err)
			}

			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined: true,
				ForUserID:    "interactionTester1",
			})
			assert.NoError(t, err)

			mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
				return v.ID
			})
			for _, post := range res {
				mapRes.Append(post)
			}
			p, ok := mapRes.Map()[post1.ID]
			assert.True(t, ok)
			assert.NotNil(t, p.MyFavorite)
			assert.Equal(t, favorite.ID, p.MyFavorite.ID)
			assert.Nil(t, p.MyVote)
		})

		t.Run("Preload vote", func(t *testing.T) {
			vote := &entity.Vote{
				PostID:  post1.ID,
				VoterID: "interactionTester1",
				Vote:    1,
			}
			if err := driver.RatePost(context.Background(), vote); err != nil {
				t.Fatal(err)
			}

			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined: true,
				ForUserID:    "interactionTester1",
			})
			assert.NoError(t, err)

			mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
				return v.ID
			})
			for _, post := range res {
				mapRes.Append(post)
			}
			p, ok := mapRes.Map()[post1.ID]
			assert.True(t, ok)
			assert.NotNil(t, p.MyFavorite)
			assert.NotNil(t, p.MyVote)
			assert.Equal(t, vote.ID, p.MyVote.ID)
		})

		t.Run("Only favorites", func(t *testing.T) {
			res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				ShowDeclined:  true,
				OnlyFavorites: true,
				ForUserID:     "interactionTester1",
			})
			assert.NoError(t, err)

			assert.Len(t, res, 1)
			p := res[0]
			assert.NotNil(t, p.MyFavorite)
		})

		t.Run("Only voted", func(t *testing.T) {
			vote := &entity.Vote{
				PostID:  post2.ID,
				VoterID: "interactionTester1",
				Vote:    -1,
			}
			if err := driver.RatePost(context.Background(), vote); err != nil {
				t.Fatal(err)
			}
			t.Run("Any", func(t *testing.T) {
				res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
					BaseFilter: repository.BaseFilter{
						Limit:  10,
						Offset: 0,
					},
					ShowDeclined: true,
					RatingFilter: types.RateTypeVoted,
					ForUserID:    "interactionTester1",
				})
				assert.NoError(t, err)

				mapRes := appender.NewMapAppender(0, func(v *entity.Post) types.PostID {
					return v.ID
				})
				for _, post := range res {
					mapRes.Append(post)
				}
				_, ok1 := mapRes.Map()[post1.ID]
				_, ok2 := mapRes.Map()[post2.ID]
				assert.True(t, ok1 && ok2)
			})

			t.Run("Upvoted", func(t *testing.T) {
				res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
					BaseFilter: repository.BaseFilter{
						Limit:  10,
						Offset: 0,
					},
					ShowDeclined: true,
					RatingFilter: types.RateTypeUpvoted,
					ForUserID:    "interactionTester1",
				})
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, post1.ID, res[0].ID)
			})

			t.Run("Downvoted", func(t *testing.T) {
				res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
					BaseFilter: repository.BaseFilter{
						Limit:  10,
						Offset: 0,
					},
					ShowDeclined: true,
					RatingFilter: types.RateTypeDownvoted,
					ForUserID:    "interactionTester1",
				})
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, post2.ID, res[0].ID)
			})

			t.Run("Voted and favorite", func(t *testing.T) {
				favorite := &entity.Favorite{
					PostID: post3.ID,
					UserID: "interactionTester1",
				}
				if err := driver.AddPostToFavorites(context.Background(), favorite); err != nil {
					t.Fatal(err)
				}

				res, err := driver.GetPosts(context.Background(), repository.GetPostsFilter{
					BaseFilter: repository.BaseFilter{
						Limit:  10,
						Offset: 0,
					},
					ShowDeclined:  true,
					RatingFilter:  types.RateTypeVoted,
					OnlyFavorites: true,
					ForUserID:     "interactionTester1",
				})
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, post1.ID, res[0].ID)
			})
		})
	})
}

func TestGORMDriver_PostAggregatedCounters(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Tag1"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	t.Run("Favorites Counter", func(t *testing.T) {
		t.Run("Increase", func(t *testing.T) {
			for i := 0; i < 83; i++ {
				favorite := &entity.Favorite{
					PostID: post.ID,
					UserID: types.UserID(fmt.Sprintf("user%d", i)),
				}
				if err := driver.AddPostToFavorites(context.Background(), favorite); err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, p.FavoritesCount, 83)
		})

		t.Run("Decrease", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				favorite := &entity.Favorite{
					PostID: post.ID,
					UserID: types.UserID(fmt.Sprintf("user%d", i)),
				}
				err := driver.RemovePostFromFavoritesByPostAndUser(context.Background(), favorite)
				if err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, p.FavoritesCount, 73)
		})
	})

	t.Run("Rating", func(t *testing.T) {
		t.Run("UpVote", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				vote := &entity.Vote{
					PostID:  post.ID,
					VoterID: types.UserID(fmt.Sprintf("goodVoter%d", i)),
					Vote:    1,
				}
				err := driver.RatePost(context.Background(), vote)
				if err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, 100, p.Rating)
		})

		t.Run("DownVote", func(t *testing.T) {
			for i := 0; i < 50; i++ {
				vote := &entity.Vote{
					PostID:  post.ID,
					VoterID: types.UserID(fmt.Sprintf("badVoter%d", i)),
					Vote:    -1,
				}
				err := driver.RatePost(context.Background(), vote)
				if err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, 50, p.Rating)
		})

		t.Run("UpVoters unvote", func(t *testing.T) {
			for i := 0; i < 25; i++ {
				vote := &entity.Vote{
					PostID:  post.ID,
					VoterID: types.UserID(fmt.Sprintf("goodVoter%d", i)),
				}
				err := driver.RemovePostRateByPostAndUser(context.Background(), vote)
				if err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, 25, p.Rating)
		})

		t.Run("DownVoters unvote", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				vote := &entity.Vote{
					PostID:  post.ID,
					VoterID: types.UserID(fmt.Sprintf("badVoter%d", i)),
				}
				err := driver.RemovePostRateByPostAndUser(context.Background(), vote)
				if err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, 35, p.Rating)
		})

		t.Run("DownVoters re-vote", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				vote := &entity.Vote{
					PostID:  post.ID,
					VoterID: types.UserID(fmt.Sprintf("badVoter%d", i+10)),
					Vote:    1,
				}
				err := driver.RatePost(context.Background(), vote)
				if err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, 55, p.Rating)
		})
	})

	t.Run("Comments Counter", func(t *testing.T) {
		t.Run("Increase", func(t *testing.T) {
			for i := 0; i < 183; i++ {
				comment := &entity.Comment{
					PostID:   post.ID,
					AuthorID: types.UserID(fmt.Sprintf("user%d", i)),
					Content:  fmt.Sprintf("Content %d", i),
				}
				if err := driver.AddComment(context.Background(), comment); err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, p.CommentsCount, 183)
		})

		t.Run("Decrease", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				if err := driver.DeleteComment(context.Background(), &entity.Comment{ID: types.CommentID(i + 1)}); err != nil {
					t.Fatal(err)
				}
			}

			p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, p.CommentsCount, 83)
		})
	})
}

func TestGORMDriver_CreateModerationAction(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Mission"},
			{Name: "Other"},
			{Name: "Non-canonical"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	err := driver.CreatePostWithContentsAndTags(context.Background(), post)
	require.NoError(t, err)

	t.Run("Create approve moderation action", func(t *testing.T) {
		action := &entity.ModerationAction{
			PostID:      post.ID,
			ModeratorID: "moderator1",
			Action:      types.ModerationActionTypeApprove,
		}

		err := driver.CreateModerationAction(context.Background(), action)
		assert.NoError(t, err)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		require.NoError(t, err)
		assert.NotNil(t, p.LastModerationID)
		assert.NotNil(t, p.LastModeration)
		assert.EqualValues(t, *p.LastModerationID, action.ID)
		assert.EqualValues(t, p.LastModeration.ID, action.ID)
		assert.EqualValues(t, "moderator1", p.LastModeration.ModeratorID)
		assert.EqualValues(t, p.LastModeration.Action, types.ModerationActionTypeApprove)
		assert.WithinDuration(t, time.Now(), p.LastModeration.CreatedAt, time.Second)
	})

	t.Run("Create decline moderation action", func(t *testing.T) {
		action := &entity.ModerationAction{
			PostID:      post.ID,
			ModeratorID: "moderator2",
			Action:      types.ModeratorActionTypeDecline,
			Note:        "Some reason for decline",
		}

		err := driver.CreateModerationAction(context.Background(), action)
		assert.NoError(t, err)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID, ShowDeclined: true})
		require.NoError(t, err)
		assert.NotNil(t, p.LastModerationID)
		assert.NotNil(t, p.LastModeration)
		assert.EqualValues(t, *p.LastModerationID, action.ID)
		assert.EqualValues(t, p.LastModeration.ID, action.ID)
		assert.EqualValues(t, "moderator2", p.LastModeration.ModeratorID)
		assert.EqualValues(t, p.LastModeration.Action, types.ModeratorActionTypeDecline)
		assert.EqualValues(t, p.LastModeration.Note, "Some reason for decline")
		assert.WithinDuration(t, time.Now(), p.LastModeration.CreatedAt, time.Second)
	})

	t.Run("Create moderation action for non-existent post", func(t *testing.T) {
		action := &entity.ModerationAction{
			PostID:      838383,
			ModeratorID: "moderator1",
			Action:      types.ModerationActionTypeApprove,
		}

		err := driver.CreateModerationAction(context.Background(), action)
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)
	})
}

func TestGORMDriver_GetModerationActions(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	posts := make(map[int]*entity.Post, 5)
	for i := 0; i < 5; i++ {
		posts[i] = &entity.Post{
			AuthorID:    types.UserID(fmt.Sprintf("Author%d", i+1)),
			Title:       fmt.Sprintf("Title %d", i+1),
			Description: fmt.Sprintf("Description %d", i+1),
			PreviewURL:  fmt.Sprintf("url%d", i+1),
			PostType:    types.PostTypeGameMode,
			Tags: []*entity.Tag{
				{Name: "Tag"},
			},
			Contents: []*entity.PostContent{
				{
					ContentType: types.ContentTypeCustomLogic,
					ContentData: fmt.Sprintf("url%d", i+1),
					IsLink:      true,
				},
			},
		}
	}

	for _, post := range posts {
		if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("No moderation actions", func(t *testing.T) {
		res, err := driver.GetModerationActions(context.Background(), repository.GetModerationActionsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
		})
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	m1 := &entity.ModerationAction{
		PostID:      posts[0].ID,
		ModeratorID: "moderator1",
		Action:      types.ModeratorActionTypeDecline,
	}
	m2 := &entity.ModerationAction{
		PostID:      posts[0].ID,
		ModeratorID: "moderator2",
		Action:      types.ModerationActionTypeApprove,
	}
	m3 := &entity.ModerationAction{
		PostID:      posts[1].ID,
		ModeratorID: "moderator1",
		Action:      types.ModerationActionTypeApprove,
	}
	m4 := &entity.ModerationAction{
		PostID:      posts[2].ID,
		ModeratorID: "moderator4",
		Action:      types.ModerationActionTypeApprove,
	}
	m5 := &entity.ModerationAction{
		PostID:      posts[2].ID,
		ModeratorID: "moderator1",
		Action:      types.ModeratorActionTypeDecline,
	}
	m6 := &entity.ModerationAction{
		PostID:      posts[2].ID,
		ModeratorID: "moderator1",
		Action:      types.ModerationActionTypeApprove,
	}

	mList := []*entity.ModerationAction{m1, m2, m3, m4, m5, m6}

	for _, m := range mList {
		if err := driver.CreateModerationAction(context.Background(), m); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Get moderation actions for a single post", func(t *testing.T) {
		res, err := driver.GetModerationActions(context.Background(), repository.GetModerationActionsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			PostID: posts[0].ID,
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.ModerationAction) types.ModerationActionID {
			return v.ID
		})
		for _, m := range res {
			mapRes.Append(m)
		}

		assert.Len(t, mapRes.Map(), 2)
		_, ok1 := mapRes.Map()[m1.ID]
		_, ok2 := mapRes.Map()[m2.ID]
		assert.True(t, ok1 && ok2)
	})

	t.Run("Get moderation actions for moderator", func(t *testing.T) {
		res, err := driver.GetModerationActions(context.Background(), repository.GetModerationActionsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			ModeratorID: "moderator1",
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.ModerationAction) types.ModerationActionID {
			return v.ID
		})
		for _, m := range res {
			mapRes.Append(m)
		}

		assert.Len(t, mapRes.Map(), 4)
		_, ok1 := mapRes.Map()[m1.ID]
		_, ok2 := mapRes.Map()[m3.ID]
		_, ok3 := mapRes.Map()[m5.ID]
		_, ok4 := mapRes.Map()[m6.ID]
		assert.True(t, ok1 && ok2 && ok3 && ok4)
	})

	t.Run("Get moderation actions by action type", func(t *testing.T) {
		res, err := driver.GetModerationActions(context.Background(), repository.GetModerationActionsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  10,
				Offset: 0,
			},
			Action: types.ModerationActionTypeApprove,
		})
		assert.NoError(t, err)

		mapRes := appender.NewMapAppender(0, func(v *entity.ModerationAction) types.ModerationActionID {
			return v.ID
		})
		for _, m := range res {
			mapRes.Append(m)
		}

		assert.Len(t, mapRes.Map(), 4)
		_, ok1 := mapRes.Map()[m2.ID]
		_, ok2 := mapRes.Map()[m3.ID]
		_, ok3 := mapRes.Map()[m4.ID]
		_, ok4 := mapRes.Map()[m6.ID]
		assert.True(t, ok1 && ok2 && ok3 && ok4)
	})

	t.Run("Get moderation actions ordered by created at", func(t *testing.T) {
		t.Run("Ascending order", func(t *testing.T) {
			res, err := driver.GetModerationActions(context.Background(), repository.GetModerationActionsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				CreatedAtOrder: repository.OrderAsc,
			})
			assert.NoError(t, err)

			assert.Len(t, res, 6)
			for i := 0; i < len(res)-1; i++ {
				assert.True(t, res[i].CreatedAt.Before(res[i+1].CreatedAt))
			}
		})

		t.Run("Descending order", func(t *testing.T) {
			res, err := driver.GetModerationActions(context.Background(), repository.GetModerationActionsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				CreatedAtOrder: repository.OrderDesc,
			})
			assert.NoError(t, err)
			assert.Len(t, res, 6)
			for i := 0; i < len(res)-1; i++ {
				assert.True(t, res[i].CreatedAt.After(res[i+1].CreatedAt))
			}
		})
	})

	t.Run("Preload post", func(t *testing.T) {
		t.Run("With", func(t *testing.T) {
			filter := repository.GetModerationActionsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
				IncludePost: true,
			}

			res, err := driver.GetModerationActions(context.Background(), filter)
			assert.NoError(t, err)
			assert.Len(t, res, 6)
			for _, m := range res {
				assert.NotNil(t, m.Post)
				assert.EqualValues(t, m.Post.ID, m.PostID)
			}
		})

		t.Run("Without", func(t *testing.T) {
			filter := repository.GetModerationActionsFilter{
				BaseFilter: repository.BaseFilter{
					Limit:  10,
					Offset: 0,
				},
			}

			res, err := driver.GetModerationActions(context.Background(), filter)
			assert.NoError(t, err)
			assert.Len(t, res, 6)
			for _, m := range res {
				assert.Nil(t, m.Post)
			}
		})
	})

	t.Run("Pagination", func(t *testing.T) {
		filter := repository.GetModerationActionsFilter{
			BaseFilter: repository.BaseFilter{
				Limit:  2,
				Offset: 1,
			},
			CreatedAtOrder: repository.OrderAsc,
		}

		res, err := driver.GetModerationActions(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.EqualValues(t, res[0].ID, m2.ID)
		assert.EqualValues(t, res[1].ID, m3.ID)
	})
}

func TestGORMDriver_AddPostToFavorites(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Tag1"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		favorite := &entity.Favorite{
			PostID: post.ID,
			UserID: "user1",
		}

		err := driver.AddPostToFavorites(context.Background(), favorite)
		assert.NoError(t, err)
		assert.NotZero(t, favorite.ID)
		assert.WithinDuration(t, time.Now(), favorite.CreatedAt, time.Second)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		assert.NoError(t, err)
		assert.Equal(t, p.FavoritesCount, 1)
	})

	t.Run("Already favorited", func(t *testing.T) {
		favorite := &entity.Favorite{
			PostID: post.ID,
			UserID: "user1",
		}

		err := driver.AddPostToFavorites(context.Background(), favorite)
		assert.ErrorIs(t, err, repoErrors.ErrAlreadyExists)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		assert.NoError(t, err)
		assert.Equal(t, p.FavoritesCount, 1)
	})

	t.Run("Invalid post PostID", func(t *testing.T) {
		favorite := &entity.Favorite{
			PostID: 838383,
			UserID: "user1",
		}

		err := driver.AddPostToFavorites(context.Background(), favorite)
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		assert.NoError(t, err)
		assert.Equal(t, p.FavoritesCount, 1)
	})
}

func TestGORMDriver_RemovePostFromFavorites(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Tag1"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	favorite := &entity.Favorite{
		PostID: post.ID,
		UserID: "user1",
	}

	if err := driver.AddPostToFavorites(context.Background(), favorite); err != nil {
		t.Fatal(err)
	}

	p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
	assert.NoError(t, err)
	assert.Equal(t, p.FavoritesCount, 1)

	t.Run("Success", func(t *testing.T) {
		err := driver.RemovePostFromFavoritesByPostAndUser(context.Background(), &entity.Favorite{
			PostID: post.ID,
			UserID: "user1",
		})
		assert.NoError(t, err)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		assert.NoError(t, err)
		assert.Equal(t, p.FavoritesCount, 0)
	})

	t.Run("Not in favorite", func(t *testing.T) {
		err := driver.RemovePostFromFavoritesByPostAndUser(context.Background(), &entity.Favorite{
			PostID: post.ID,
			UserID: "user1",
		})
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		assert.NoError(t, err)
		assert.Equal(t, p.FavoritesCount, 0)
	})

	t.Run("Invalid post PostID", func(t *testing.T) {
		err := driver.RemovePostFromFavoritesByPostAndUser(context.Background(), &entity.Favorite{
			PostID: 838383,
			UserID: "user1",
		})
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		assert.NoError(t, err)
		assert.Equal(t, p.FavoritesCount, 0)
	})
}

func TestGORMDriver_RatePost(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Tag1"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		vote := &entity.Vote{
			PostID:  post.ID,
			VoterID: "user1",
			Vote:    1,
		}
		err := driver.RatePost(context.Background(), vote)
		assert.NoError(t, err)
		assert.NotZero(t, vote.ID)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, p.Rating, 1)
	})

	t.Run("Re-Vote", func(t *testing.T) {
		rate := &entity.Vote{
			PostID:  post.ID,
			VoterID: "user1",
			Vote:    -1,
		}
		err := driver.RatePost(context.Background(), rate)
		assert.NoError(t, err)
		assert.NotZero(t, rate.ID)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, p.Rating, -1)
	})

	t.Run("Invalid post PostID", func(t *testing.T) {
		rate := &entity.Vote{
			PostID:  838383,
			VoterID: "user1",
			Vote:    1,
		}
		err := driver.RatePost(context.Background(), rate)
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)
	})

	t.Run("Invalide vote value", func(t *testing.T) {
		rate := &entity.Vote{
			PostID:  post.ID,
			VoterID: "user1",
			Vote:    2,
		}
		err := driver.RatePost(context.Background(), rate)
		assert.Error(t, err)

		rate = &entity.Vote{
			PostID:  post.ID,
			VoterID: "user1",
			Vote:    -2,
		}

		err = driver.RatePost(context.Background(), rate)
		assert.Error(t, err)

		rate = &entity.Vote{
			PostID:  post.ID,
			VoterID: "user1",
			Vote:    0,
		}

		err = driver.RatePost(context.Background(), rate)
		assert.Error(t, err)
	})
}

func TestGORMDriver_UnRatePost(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Tag1"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	vote := &entity.Vote{
		PostID:  post.ID,
		VoterID: "user1",
		Vote:    1,
	}
	if err := driver.RatePost(context.Background(), vote); err != nil {
		t.Fatal(err)
	}

	p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
	if err != nil {
		t.Fatal(err)
	}
	if p.Rating != 1 {
		t.Fatal("Expected rating to be 1, got:", p.Rating)
	}

	t.Run("Success", func(t *testing.T) {
		err := driver.RemovePostRateByPostAndUser(context.Background(), &entity.Vote{
			PostID:  post.ID,
			VoterID: "user1",
		})
		assert.NoError(t, err)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, p.Rating, 0)
	})

	t.Run("Not rated", func(t *testing.T) {
		err := driver.RemovePostRateByPostAndUser(context.Background(), &entity.Vote{
			PostID:  838383,
			VoterID: "user1",
		})
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)
	})
}

func TestGORMDriver_AddComment(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Tag1"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		comment := &entity.Comment{
			PostID:   post.ID,
			AuthorID: "commentator1",
			Content:  "Comment Content",
		}

		err := driver.AddComment(context.Background(), comment)
		assert.NoError(t, err)
		assert.NotZero(t, comment.ID)

		p, err := driver.GetPost(context.Background(), repository.GetPostFilter{PostID: post.ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, p.CommentsCount, 1)
	})

	t.Run("Invalid post PostID", func(t *testing.T) {
		comment := &entity.Comment{
			PostID:   838383,
			AuthorID: "commentator1",
			Content:  "Comment Content",
		}

		err := driver.AddComment(context.Background(), comment)
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)
	})
}

func TestGORMDriver_UpdateComment(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Tag1"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		comment := &entity.Comment{
			PostID:   post.ID,
			AuthorID: "commentator1",
			Content:  "Original Comment Content",
		}

		err := driver.AddComment(context.Background(), comment)
		assert.NoError(t, err)

		updatedComment := &entity.Comment{
			ID:       comment.ID,
			PostID:   post.ID,
			AuthorID: "commentator1",
			Content:  "Updated Comment Content",
		}

		err = driver.UpdateComment(context.Background(), updatedComment)
		assert.NoError(t, err)
		assert.Equal(t, comment.ID, updatedComment.ID)

		c := &entity.Comment{}
		if err := driver.db.Where("id = ?", comment.ID).First(c).Error; err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, c.ID, updatedComment.ID)
		assert.Equal(t, c.Content, updatedComment.Content)
	})

	t.Run("Update with the same content", func(t *testing.T) {
		comment := &entity.Comment{
			PostID:   post.ID,
			AuthorID: "commentator1",
			Content:  "Original Comment Content",
		}

		err := driver.AddComment(context.Background(), comment)
		assert.NoError(t, err)

		updatedComment := &entity.Comment{
			ID:       comment.ID,
			PostID:   post.ID,
			AuthorID: "commentator1",
			Content:  "Original Comment Content",
		}

		err = driver.UpdateComment(context.Background(), updatedComment)
		assert.NoError(t, err)
		assert.Equal(t, comment.ID, updatedComment.ID)

		c := &entity.Comment{}
		if err := driver.db.Where("id =?", comment.ID).First(c).Error; err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, c.ID, updatedComment.ID)
		assert.Equal(t, c.Content, updatedComment.Content)
	})

	t.Run("Invalid comment PostID", func(t *testing.T) {
		comment := &entity.Comment{
			ID:       838383,
			PostID:   post.ID,
			AuthorID: "commentator1",
			Content:  "Updated Comment Content",
		}

		err := driver.UpdateComment(context.Background(), comment)
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)
	})
}

func TestGORMDriver_DeleteComment(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)

	post := &entity.Post{
		AuthorID:    "authorID",
		Title:       "Post Title",
		Description: "Post Description",
		PreviewURL:  "preview.url",
		PostType:    types.PostTypeGameMode,
		Tags: []*entity.Tag{
			{Name: "Tag1"},
		},
		Contents: []*entity.PostContent{
			{
				ContentType: types.ContentTypeCustomLogic,
				ContentData: "url",
				IsLink:      true,
			},
			{
				ContentType: types.ContentTypeCustomMap,
				ContentData: "mapraw",
				IsLink:      false,
			},
		},
	}

	if err := driver.CreatePostWithContentsAndTags(context.Background(), post); err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		comment := &entity.Comment{
			PostID:   post.ID,
			AuthorID: "commentator1",
			Content:  "Original Comment Content",
		}

		err := driver.AddComment(context.Background(), comment)
		assert.NoError(t, err)

		err = driver.DeleteComment(context.Background(), comment)
		assert.NoError(t, err)

		c := &entity.Comment{}
		err = driver.db.Where("id = ?", comment.ID).First(c).Error
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Invalid comment PostID", func(t *testing.T) {
		err := driver.DeleteComment(context.Background(), &entity.Comment{
			ID: 838383,
		})
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)
	})
}

func TestGORMDriver_GetURLValidatorConfig(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)
	ctx := context.Background()

	err := driver.Truncate(ctx, []string{URLValidatorConfigsTableName})
	require.NoError(t, err)

	fixture := &entity.URLValidatorConfig{
		Type:       "image",
		Protocols:  pq.StringArray([]string{"http", "https"}),
		Domains:    pq.StringArray([]string{"i.imgur.com", "imgur.com", "image.ibb.co"}),
		Extensions: pq.StringArray([]string{".jpg", ".png", ".jpeg"}),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err = driver.db.Create(fixture).Error
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		cfg, err := driver.GetURLValidatorConfig(ctx, "image")
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "image", cfg.Type)
		assert.Equal(t, pq.StringArray([]string{"http", "https"}), cfg.Protocols)
		assert.Equal(t, pq.StringArray([]string{"i.imgur.com", "imgur.com", "image.ibb.co"}), cfg.Domains)
		assert.Equal(t, pq.StringArray([]string{".jpg", ".png", ".jpeg"}), cfg.Extensions)
	})

	t.Run("NotFound", func(t *testing.T) {
		cfg, err := driver.GetURLValidatorConfig(ctx, "nonexistent")
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.ErrorIs(t, err, repoErrors.ErrNotFound)
	})

	t.Run("GenericError", func(t *testing.T) {
		newDriver := setupTestDB(t)
		teardownTestDB(t, newDriver)

		_, err := newDriver.GetURLValidatorConfig(ctx, "image")
		require.Error(t, err)
	})
}

func TestGORMDriver_GetAllURLValidatorConfigs(t *testing.T) {
	driver := setupTestDB(t)
	defer teardownTestDB(t, driver)
	ctx := context.Background()

	fixtures := []entity.URLValidatorConfig{
		{
			Type:       "image",
			Protocols:  []string{"http", "https"},
			Domains:    []string{"i.imgur.com", "imgur.com", "image.ibb.co"},
			Extensions: []string{".jpg", ".png", ".jpeg"},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			Type:       "text",
			Protocols:  []string{"http", "https"},
			Domains:    []string{"pastebin.com", "www.dropbox.com", "raw.githubusercontent.com"},
			Extensions: []string{".txt", ".md", ".csv"},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			Type:       "asset_bundle",
			Protocols:  []string{"http", "https"},
			Domains:    []string{"www.dropbox.com"},
			Extensions: []string{".exe", ".bin", ".dll"},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	err := driver.Truncate(ctx, []string{URLValidatorConfigsTableName})
	require.NoError(t, err)

	for i := range fixtures {
		err := driver.db.Create(&fixtures[i]).Error
		assert.NoError(t, err)
	}

	app := appender.NewMapAppender[string, *entity.URLValidatorConfig](0, func(cfg *entity.URLValidatorConfig) string {
		return cfg.Type
	})

	err = driver.GetAllURLValidatorConfigs(ctx, app)
	assert.NoError(t, err)
	m := app.Map()
	assert.Equal(t, 3, len(m))
	cfg, ok := m["image"]
	assert.True(t, ok)
	assert.Equal(t, "image", cfg.Type)
	cfg, ok = m["text"]
	assert.True(t, ok)
	assert.Equal(t, "text", cfg.Type)
	cfg, ok = m["asset_bundle"]
	assert.True(t, ok)
	assert.Equal(t, "asset_bundle", cfg.Type)
}
