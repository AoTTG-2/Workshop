package echo

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"workshop/internal/controller"
	"workshop/internal/service/workshop"
)

// AddComment godoc
//
//	@Summary		Add Comment
//	@Description	Add a comment to a post. Only authorized users with role 'USER' can add a comment.
//	@Tags			Comments
//	@Accept			json
//	@Produce		json
//	@Param			request	body		controller.AddCommentRequest	true	"Request payload with comment data"
//	@Success		200		{object}	workshop.Comment				"Comment created successfully"
//	@Failure		400		{object}	controller.APIError				"Bad Request – invalid input payload"
//	@Failure		403		{object}	controller.APIError				"Forbidden – insufficient user rights"
//	@Failure		404		{object}	controller.APIError				"Not Found – associated post not found"
//	@Failure		500		{object}	controller.APIError				"Internal Server Error"
//	@Router			/comments [post]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) AddComment(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RoleUser) {
		return echo.NewHTTPError(http.StatusForbidden, "Only authorized users can comment")
	}

	req := new(controller.AddCommentRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	comment, err := c.ws.AddComment(ctx.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, workshop.ErrPostNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		default:
			return err
		}
	}

	return ctx.JSON(http.StatusOK, comment)
}

// UpdateComment godoc
//
//	@Summary		Update Comment
//	@Description	Update an existing comment. Only authorized users with role 'USER' can update their comments.
//	@Tags			Comments
//	@Accept			json
//	@Produce		json
//	@Param			commentID	path		string							true	"Comment ID"
//	@Param			request		body		controller.UpdateCommentRequest	true	"Request payload with updated comment data"
//	@Success		200			{object}	workshop.Comment				"Comment updated successfully"
//	@Failure		400			{object}	controller.APIError				"Bad Request – invalid input payload"
//	@Failure		403			{object}	controller.APIError				"Forbidden – insufficient user rights"
//	@Failure		404			{object}	controller.APIError				"Not Found – comment not found"
//	@Failure		500			{object}	controller.APIError				"Internal Server Error"
//	@Router			/comments/{commentID} [put]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) UpdateComment(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RoleUser) {
		return echo.NewHTTPError(http.StatusForbidden, "Only authorized users can update comments")
	}

	req := new(controller.UpdateCommentRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	comment, err := c.ws.UpdateComment(ctx.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, workshop.ErrCommentNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		default:
			return err
		}
	}

	return ctx.JSON(http.StatusOK, comment)
}

// DeleteComment godoc
//
//	@Summary		Delete Comment
//	@Description	Delete an existing comment. Only authorized users with role 'USER' can delete their comments.
//	@Tags			Comments
//	@Accept			json
//	@Produce		json
//	@Param			commentID	path	string	true	"Comment ID"
//	@Success		204			"Comment deleted successfully, no content returned"
//	@Failure		400			{object}	controller.APIError	"Bad Request – invalid input payload"
//	@Failure		403			{object}	controller.APIError	"Forbidden – insufficient user rights"
//	@Failure		404			{object}	controller.APIError	"Not Found – comment not found"
//	@Failure		500			{object}	controller.APIError	"Internal Server Error"
//	@Router			/comments/{commentID} [delete]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) DeleteComment(ctx echo.Context) error {
	user := c.getUser(ctx)

	if !user.HasRole(controller.RoleUser) {
		return echo.NewHTTPError(http.StatusForbidden, "Only authorized users can delete comments")
	}

	req := new(controller.DeleteCommentRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	req.UserID = user.ID

	if err := c.ws.DeleteComment(ctx.Request().Context(), req); err != nil {
		switch {
		case errors.Is(err, workshop.ErrCommentNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		default:
			return err
		}
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetComments godoc
//
//	@Summary		Get Comments
//	@Description	Returns a list of post's comments.
//	@Tags			Posts, Comments
//	@Accept			json
//	@Produce		json
//	@Param			postID		query		types.PostID			false	"Post ID"
//	@Param			authorID	query		types.UserID			false	"Author ID"
//	@Param			sort_order	query		controller.SortOrder	false	"Sort order by creation time"	default(desc)
//	@Param			page		query		int						true	"Page number"					minimum(1)
//	@Param			limit		query		int						true	"Number of comments per page"	minimum(1)	maximum(100)
//	@Success		200			{array}		workshop.Comment		"List of comments"
//	@Failure		400			{object}	controller.APIError		"Bad Request – invalid payload"
//	@Failure		500			{object}	controller.APIError		"Internal Server Error"
//	@Router			/comments [get]
//	@Security		DebugUserRoles
//	@Security		DebugUserID
func (c *PostHandler) GetComments(ctx echo.Context) error {
	req := new(controller.GetCommentsRequest)
	if err := c.bindAndValidate(ctx, req); err != nil {
		return err
	}

	if req.SortOrder == "" {
		req.SortOrder = controller.SortOrderDescending
	}

	comments, err := c.ws.GetComments(ctx.Request().Context(), req)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, comments)
}
