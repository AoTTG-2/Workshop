package controller

import "workshop/internal/types"

type UserRole string

const (
	RoleGuest         UserRole = "GUEST"
	RoleUser          UserRole = "USER"
	RoleImpersonator  UserRole = "IMPERSONATOR"
	RolePostModerator UserRole = "POST_MODERATOR"
	RolePostCreator   UserRole = "POST_CREATOR"
)

type User struct {
	ID       types.UserID          `json:"id"`
	Username string                `json:"username"`
	Roles    map[UserRole]struct{} `json:"roles"`
}

func (u *User) HasRole(role UserRole) bool {
	_, ok := u.Roles[role]
	return ok
}
