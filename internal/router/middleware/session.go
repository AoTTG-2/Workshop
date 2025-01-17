package middleware

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"workshop/internal/controller"
	"workshop/internal/types"
)

type SessionProvider interface {
	Get(ctx context.Context, token string) (controller.User, error)
}

type ExtractionMethod int

const (
	ExtractionMethodQueryParam ExtractionMethod = iota
	ExtractionMethodHeader
	ExtractionMethodBoth
)

const (
	UserRolesDebugHeader = "X-Debug-User-Roles"
	UserIDDebugHeader    = "X-Debug-User-ID"
)

type SessionsMiddlewareConfig struct {
	QueryParam       string
	Header           string
	ExtractionMethod ExtractionMethod
	Debug            bool
}

type SessionMiddleware struct {
	sessionProvider SessionProvider
	cfg             SessionsMiddlewareConfig
}

func NewSessionsMiddleware(
	sessionProvider SessionProvider,
	cfg SessionsMiddlewareConfig,
) echo.MiddlewareFunc {
	m := &SessionMiddleware{
		sessionProvider: sessionProvider,
		cfg:             cfg,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			//ctx := c.Request().Context()
			token := m.extractToken(c)
			if token == "" {
				user := &controller.User{
					Roles: map[controller.UserRole]struct{}{controller.RoleGuest: {}}}

				if m.cfg.Debug {
					rolesHeader := c.Request().Header.Get(UserRolesDebugHeader)
					userIDHeader := c.Request().Header.Get(UserIDDebugHeader)
					if rolesHeader != "" && userIDHeader == "" {
						return echo.NewHTTPError(http.StatusUnauthorized, "User ID header is required for debug mode")
					}

					if rolesHeader != "" {
						delete(user.Roles, controller.RoleGuest)
						for _, role := range strings.Split(rolesHeader, ",") {
							user.Roles[controller.UserRole(role)] = struct{}{}
						}
					}

					if userIDHeader != "" {
						user.ID = types.UserID(userIDHeader)
					}
				}
				c.Set("user", user)
				return next(c)
			}

			//user, err := m.sessionProvider.Get(ctx, token)
			//if err != nil {
			//	if errors.Is(err, sessions.ErrInvalidSession) {
			//		return echo.ErrUnauthorized
			//	}
			//	return err
			//}
			//
			//c.Set("user", user)

			return next(c)
		}
	}
}

func (m *SessionMiddleware) extractToken(c echo.Context) string {
	switch m.cfg.ExtractionMethod {
	case ExtractionMethodQueryParam:
		return c.QueryParam(m.cfg.QueryParam)
	case ExtractionMethodHeader:
		return c.Request().Header.Get(m.cfg.Header)
	case ExtractionMethodBoth:
		if sessionID := c.QueryParam(m.cfg.QueryParam); sessionID != "" {
			return sessionID
		}
		return c.Request().Header.Get(m.cfg.Header)
	default:
		return ""
	}
}
