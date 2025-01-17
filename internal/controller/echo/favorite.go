package echo

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"workshop/internal/controller"
	"workshop/internal/service/workshop"
)

// FavoritePost godoc
//
//	@Summary		Favorite Post
//	@Description	Mark a post as favorite for the current user. Only authorized users with role 'USER' can perform this action.
//	@Tags			Favorites, Interactions
//	@Accept			json
//	@Produce		json
//	@Param			postID	path	string	true	"Post ID"
//	@Success		200		"Success – post was successfully added to favorites (no content returned)"
//	@Failure		400		{object}	controller.APIError	"Bad Request – invalid input payload"
//	@Failure		403		{object}	controller.APIError	"Forbidden – only authorized users can favorite posts"
//	@Failure		404		{object}	controller.APIError	"Not Found – post not found"
//	@Failure		409		{object}	controller.APIError	"Conflict – post is already in favorites"
//	@Failure		500		{object}	controller.APIError	"Internal Server Error"
//	@Router			/posts/{postID}/favorite [post]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) FavoritePost(ctx echo.Context) error {
	// TODO: Allow owned post?
	user := c.getUser(ctx)

	if !user.HasRole(controller.RoleUser) {
		return echo.NewHTTPError(http.StatusForbidden, "Only authorized users can unfavorite posts")
	}

	req := new(controller.FavoritePostRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	if err := c.ws.FavoritePost(ctx.Request().Context(), req); err != nil {
		switch {
		case errors.Is(err, workshop.ErrPostNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		case errors.Is(err, workshop.ErrPostAlreadyFavorite):
			return echo.NewHTTPError(http.StatusConflict, "Post is already in favorites")
		default:
			return err
		}
	}

	return ctx.NoContent(http.StatusOK)
}

// UnfavoritePost godoc
//
//	@Summary		Unfavorite Post
//	@Description	Removes a post from the current user's favorites. Only authorized users with role 'USER' can perform this action.
//	@Tags			Favorites, Interactions
//	@Accept			json
//	@Produce		json
//	@Param			postID	path	string	true	"Post ID"
//	@Success		204		"No Content – the post was successfully removed from favorites"
//	@Failure		400		{object}	controller.APIError	"Bad Request – invalid input payload"
//	@Failure		403		{object}	controller.APIError	"Forbidden – only authorized users can unfavorite posts"
//	@Failure		404		{object}	controller.APIError	"Not Found – post is not marked as favorite or does not exist"	//	(рекомендуемый статус: 412 Precondition Failed, если статус отличается)
//	@Failure		500		{object}	controller.APIError	"Internal Server Error"
//	@Router			/posts/{postID}/favorite [delete]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) UnfavoritePost(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RoleUser) {
		return echo.NewHTTPError(http.StatusForbidden, "Only authorized users can unfavorite posts")
	}

	req := new(controller.UnfavoritePostRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	if err := c.ws.UnfavoritePost(ctx.Request().Context(), req); err != nil {
		switch {
		case errors.Is(err, workshop.ErrPostNotFavorite):
			return echo.NewHTTPError(http.StatusPreconditionFailed, "Post is not in favorites")
		default:
			return err
		}
	}

	return ctx.NoContent(http.StatusNoContent)
}
