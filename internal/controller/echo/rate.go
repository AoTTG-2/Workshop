package echo

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"workshop/internal/controller"
	"workshop/internal/service/workshop"
)

// RatePost godoc
//
//	@Summary		Rate Post
//	@Description	Allows an authorized user (role "user") to rate a post.
//	@Tags			Posts, Interactions
//	@Accept			json
//	@Produce		json
//	@Param			request	body	controller.RatePostRequest	true	"Request payload containing the post identifier and rating data"
//	@Success		200		"Success – the post rating has been processed successfully (no content returned)"
//	@Failure		400		{object}	controller.APIError	"Bad Request – invalid input or payload"
//	@Failure		403		{object}	controller.APIError	"Forbidden – only authorized users can rate posts"
//	@Failure		404		{object}	controller.APIError	"Not Found – the post was not found"
//	@Failure		409		{object}	controller.APIError	"Conflict – the post has already been rated by the user"
//	@Failure		500		{object}	controller.APIError	"Internal Server Error"
//	@Router			/posts/{postID}/rate [post]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) RatePost(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RoleUser) {
		return echo.NewHTTPError(http.StatusForbidden, "Only authorized users can rate posts")
	}

	req := new(controller.RatePostRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	if err := c.ws.RatePost(ctx.Request().Context(), req); err != nil {
		switch {
		case errors.Is(err, workshop.ErrPostNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		case errors.Is(err, workshop.ErrPostNotRated):
			return echo.NewHTTPError(http.StatusConflict, "Post has already been rated")
		default:
			return err
		}
	}

	return ctx.NoContent(http.StatusOK)
}
