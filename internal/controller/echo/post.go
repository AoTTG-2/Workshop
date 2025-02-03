package echo

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
	"workshop/internal/controller"
	"workshop/internal/service/workshop"
	"workshop/internal/types"
	_ "workshop/internal/types"
)

type URLValidator interface {
	Validate(url string) error
}

type PostHandler struct {
	*baseHandler
	ws                      *workshop.Workshop
	imageURLValidator       URLValidator
	textURLValidator        URLValidator
	assetBundleURLValidator URLValidator
}

func NewPostHandler(
	ws *workshop.Workshop,
	imageURLValidator URLValidator,
	textURLValidator URLValidator,
	assetBundleURLValidator URLValidator,
) *PostHandler {
	return &PostHandler{
		baseHandler:             &baseHandler{},
		ws:                      ws,
		imageURLValidator:       imageURLValidator,
		textURLValidator:        textURLValidator,
		assetBundleURLValidator: assetBundleURLValidator,
	}
}

// GetList godoc
//
//	@Summary		Get Posts List
//	@Description	Returns a list of posts, filtered by the request parameters.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			search_query	query		string						false	"Search query"	maxlength(255)
//	@Param			author_id		query		string						false	"Author ID"
//	@Param			only_approved	query		boolean						false	"Filter to include only approved posts"
//	@Param			show_declined	query		boolean						false	"Filter to include declined posts. Requires 'POST_MODERATOR' role."
//	@Param			type			query		string						false	"Post type. Valid values should be enforced by server-side validation."
//	@Param			tags			query		[]string					false	"Tags filter"	collectionFormat(multi)	minlength(1)	maxlength(10)
//	@Param			for_user_id		query		string						false	"User ID for whom posts are retrieved. Requires 'IMPERSONATOR' role."
//	@Param			only_favorites	query		boolean						false	"Filter to include only favorite posts"
//	@Param			rating_filter	query		types.RateType				false	"Whether to show posts that user has rated"
//	@Param			sort_type		query		controller.PostsSortType	false	"Sort type"
//	@Param			sort_order		query		controller.SortOrder		false	"Sort order"				default(desc)
//	@Param			page			query		int							true	"Page number"				minimum(1)
//	@Param			limit			query		int							true	"Number of posts per page"	minimum(1)	maximum(100)
//	@Success		200				{array}		workshop.Post				"List of posts"
//	@Failure		400				{object}	controller.APIError			"Bad Request – invalid filter parameters or conflicting filter options"
//	@Failure		403				{object}	controller.APIError			"Forbidden – insufficient permissions for specified filters"
//	@Failure		500				{object}	controller.APIError			"Internal Server Error"
//	@Router			/posts [get]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) GetList(ctx echo.Context) error {
	user := c.getUser(ctx)

	req := new(controller.GetPostsRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	if req.ForUserID != "" && !user.HasRole(controller.RoleImpersonator) {
		return echo.NewHTTPError(http.StatusForbidden, "Not enough permissions to use 'for_user_id' filter")
	}

	if req.ShowDeclined && !user.HasRole(controller.RolePostModerator) {
		return echo.NewHTTPError(http.StatusForbidden, "Not enough permissions to use 'show_declined' filter")
	}

	if req.ShowDeclined && req.OnlyApproved {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot use 'only_approved' and 'show_declined' filters simultaneously.")
	}

	if req.ForUserID == "" {
		req.ForUserID = user.ID
	}

	if req.SortType != "" && req.SortOrder == "" {
		req.SortOrder = controller.SortOrderDescending
	}

	posts, err := c.ws.GetPosts(ctx.Request().Context(), req)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, posts)
}

// GetOne godoc
//
//	@Summary		Get Post
//	@Description	Retrieves a single post by ID.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			postID		path		string				true	"Post ID"
//	@Param			for_user_id	query		string				false	"User ID for whom posts are retrieved. Requires 'IMPERSONATOR' role"
//	@Success		200			{object}	workshop.Post		"Post details"
//	@Failure		400			{object}	controller.APIError	"Bad Request – invalid parameters"
//	@Failure		403			{object}	controller.APIError	"Forbidden – insufficient permissions for specified filters"
//	@Failure		404			{object}	controller.APIError	"Not Found – post not found"
//	@Failure		500			{object}	controller.APIError	"Internal Server Error"
//	@Router			/posts/{postID} [get]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) GetOne(ctx echo.Context) error {
	user := c.getUser(ctx)

	req := new(controller.GetPostRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	if req.ForUserID != "" && !user.HasRole(controller.RoleImpersonator) {
		return echo.NewHTTPError(http.StatusForbidden, "Not enough permissions to use 'for_user_id' filter")
	}

	if req.ForUserID == "" {
		req.ForUserID = user.ID
	}

	post, err := c.ws.GetPost(ctx.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, workshop.ErrPostNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		default:
			return err
		}
	}

	return ctx.JSON(http.StatusOK, post)
}

// Create godoc
//
//	@Summary		Create Post
//	@Description	Creates a new post. The caller must have the 'post creator' role.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			request	body		controller.CreatePostRequest									true	"Request payload containing new post data"
//	@Success		201		{object}	workshop.Post													"Newly created post"
//	@Failure		400		{object}	controller.APIError												"Bad Request – invalid input data"
//	@Failure		403		{object}	controller.APIError												"Forbidden – insufficient permissions to create posts"
//	@Failure		429		{object}	controller.APIGenericError[controller.APIRateLimitErrorData]	"Rate limit exceeded"
//	@Failure		500		{object}	controller.APIError												"Internal Server Error"
//	@Router			/posts [post]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) Create(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RolePostCreator) {
		return echo.NewHTTPError(http.StatusForbidden, "Not enough permissions to create posts")
	}

	req := new(controller.CreatePostRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	if req.PreviewURL != "" {
		if err := c.imageURLValidator.Validate(req.PreviewURL); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid preview URL '%s': %v", req.PreviewURL, err))
		}
	}

	for _, content := range req.Contents {
		if !content.IsLink {
			continue
		}

		switch content.Type {
		// TODO: Implement how to pass and store skins
		// Probably something like type = "skin:skin_part"
		// Or something more complex if we want to support multiple skin sets / skins of the same type and group it somehow
		case types.ContentTypeSkin:
			if err := c.imageURLValidator.Validate(content.Data); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid skin URL '%s': %v", content.Data, err))
			}
		case types.ContentTypeCustomLogic:
			if err := c.textURLValidator.Validate(content.Data); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid custom logic URL '%s': %v", content.Data, err))
			}
		case types.ContentTypeCustomMap:
			if err := c.textURLValidator.Validate(content.Data); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid custom map URL '%s': %v", content.Data, err))
			}
		case types.ContentTypeCustomAsset:
			if err := c.assetBundleURLValidator.Validate(content.Data); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid asset bundle URL '%s': %v", content.Data, err))
			}
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unsupported content type '%s'", content.Type))
		}
	}

	post, err := c.ws.CreatePost(ctx.Request().Context(), req)
	if err != nil {
		var rlErr *workshop.RateLimitExceededError
		if errors.As(err, &rlErr) {
			return echo.NewHTTPError(http.StatusTooManyRequests, &controller.APIGenericError[controller.APIRateLimitErrorData]{
				Message: fmt.Sprintf("Rate limit exceeded. Try again after %s.", rlErr.Info.ResetAt.Format(time.RFC3339)),
				Data: controller.APIRateLimitErrorData{
					ResetAt: rlErr.Info.ResetAt,
				},
			})
		}
		return err
	}

	return ctx.JSON(http.StatusCreated, post)
}

// Update godoc
//
//	@Summary		Update Post
//	@Description	Updates an existing post.
//	@Description	The caller must have the 'POST_CREATOR' role and must be the owner of the post.
//	@Description	Leave content ID 0 to add new content.
//	@Description	If content ID is not 0, it will try to update the content.
//	@Description	When updating, excluded contents and tags will be removed.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		string							true	"Post ID"
//	@Param			request	body		controller.UpdatePostRequest	true	"Request payload containing updated post data"
//	@Success		200		{object}	workshop.Post					"Updated post details"
//	@Failure		400		{object}	controller.APIError				"Bad Request – invalid input data"
//	@Failure		403		{object}	controller.APIError				"Forbidden – insufficient permissions to update posts"
//	@Failure		404		{object}	controller.APIError				"Not Found – post not found"
//	@Failure		412		{object}	controller.APIError				"Precondition Failed – the post is not owned by the user"
//	@Failure		500		{object}	controller.APIError				"Internal Server Error"
//	@Router			/posts/{postID} [put]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) Update(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RolePostCreator) {
		return echo.NewHTTPError(http.StatusForbidden, "Not enough permissions to update posts")
	}

	req := new(controller.UpdatePostRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	post, err := c.ws.UpdatePost(ctx.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, workshop.ErrPostNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		case errors.Is(err, workshop.ErrPostNotOwned):
			return echo.NewHTTPError(http.StatusPreconditionFailed, "Post not owned by the user")
		case errors.Is(err, workshop.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		default:
			return err
		}
	}

	return ctx.JSON(http.StatusOK, post)
}

// Delete godoc
//
//	@Summary		Delete Post
//	@Description	Deletes an existing post.
//	@Description	The caller must have the 'POST_CREATOR' role and must be the owner of the post.
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path	string							true	"Post ID"
//	@Param			request	body	controller.DeletePostRequest	true	"Request payload containing the identifier of the post to delete"
//	@Success		204		"Success – the post was deleted successfully (no content returned)"
//	@Failure		400		{object}	controller.APIError	"Bad Request – invalid input data"
//	@Failure		403		{object}	controller.APIError	"Forbidden – insufficient permissions to delete posts"
//	@Failure		404		{object}	controller.APIError	"Not Found – post not found"
//	@Failure		412		{object}	controller.APIError	"Precondition Failed – the post is not owned by the user"
//	@Failure		500		{object}	controller.APIError	"Internal Server Error"
//	@Router			/posts/{postID} [delete]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) Delete(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RolePostCreator) {
		return echo.NewHTTPError(http.StatusForbidden, "Not enough permissions to delete posts")
	}

	req := new(controller.DeletePostRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	if err := c.ws.DeletePost(ctx.Request().Context(), req); err != nil {
		switch {
		case errors.Is(err, workshop.ErrPostNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		case errors.Is(err, workshop.ErrPostNotOwned):
			return echo.NewHTTPError(http.StatusPreconditionFailed, "Post not owned by the user")
		default:
			return err
		}
	}

	return ctx.NoContent(http.StatusNoContent)
}
