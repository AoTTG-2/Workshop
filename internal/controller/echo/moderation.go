package echo

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"workshop/internal/controller"
	_ "workshop/internal/service/workshop"
	_ "workshop/internal/types"
)

// ModeratePost godoc
//
//	@Summary		Moderate Post
//	@Description	Allows a user with the 'POST_MODERATOR' role to moderate a post.
//	@Tags			Posts, Moderation
//	@Accept			json
//	@Produce		json
//	@Param			postID	path	string							true	"Post ID"
//	@Param			request	body	controller.ModeratePostRequest	true	"Request payload containing moderation action and post identifier"
//	@Success		200		"Success – the moderation action has been applied (no content returned)"
//	@Failure		400		{object}	controller.APIError	"Bad Request – invalid payload"
//	@Failure		403		{object}	controller.APIError	"Forbidden – caller lacks post moderator permissions"
//	@Failure		500		{object}	controller.APIError	"Internal Server Error"
//	@Router			/posts/{postID}/moderate [post]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) ModeratePost(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RolePostModerator) {
		return echo.NewHTTPError(http.StatusForbidden, "Not enough permissions to moderate posts")
	}

	req := new(controller.ModeratePostRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	if err := c.ws.ModeratePost(ctx.Request().Context(), req); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

// GetModerationActions godoc
//
//	@Summary		Get Moderation Actions
//	@Description	Returns a list of moderation actions.
//	@Tags			Moderation
//	@Accept			json
//	@Produce		json
//	@Param			postID		query		int							false	"Post ID"
//	@Param			moderatorID	query		string						false	"Moderator ID"
//	@Param			action		query		types.ModerationActionType	false	"Moderation action"
//	@Param			sort_order	query		controller.SortOrder		false	"Sort order by creation time"			default(desc)
//	@Param			page		query		int							true	"Page number"							minimum(1)
//	@Param			limit		query		int							true	"Number of moderation actions per page"	minimum(1)	maximum(100)
//	@Success		200			{array}		workshop.ModerationAction	"List of moderation actions"
//	@Failure		400			{object}	controller.APIError			"Bad Request – invalid payload"
//	@Failure		403			{object}	controller.APIError			"Forbidden – caller lacks post moderator permissions"
//	@Failure		500			{object}	controller.APIError			"Internal Server Error"
//	@Router			/moderation-actions [get]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) GetModerationActions(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RolePostModerator) {
		return echo.NewHTTPError(http.StatusForbidden, "Not enough permissions to view moderation actions")
	}

	req := new(controller.GetModerationActionsRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	if req.SortOrder == "" {
		req.SortOrder = controller.SortOrderDescending
	}

	actions, err := c.ws.GetModerationActions(ctx.Request().Context(), req)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, actions)
}
