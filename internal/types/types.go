package types

type UserID string
type PostID int
type PostContentID int
type FavoriteID int
type TagID int
type ModerationActionID int
type CommentID int
type VoteID int

type ModerationActionType string

const (
	ModerationActionTypeApprove ModerationActionType = "approve"
	ModeratorActionTypeDecline  ModerationActionType = "decline"
)

type ModerationStatus string

const (
	ModerationStatusApproved ModerationStatus = "approved"
	ModerationStatusDeclined ModerationStatus = "declined"
	ModerationStatusPending  ModerationStatus = "pending"
)

type PostType string

const (
	PostTypeGameMode       = "game_mode"
	PostTypeMap            = "map"
	PostTypeMapAndGameMode = "map_and_game_mode"
	PostTypeMapSuite       = "map_suite"
)

type ContentType string

const (
	ContentTypeSkinSet     = "skin_set"
	ContentTypeCustomMap   = "custom_map"
	ContentTypeCustomLogic = "custom_logic"
	ContentTypeCustomAsset = "custom_asset"
)

type RateType string

const (
	RateTypeUpvoted   RateType = "upvoted"
	RateTypeDownvoted RateType = "downvoted"
	RateTypeVoted     RateType = "voted"
	RateTypeNone      RateType = "none"
)
