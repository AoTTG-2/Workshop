package entity

import "workshop/internal/types"

type Vote struct {
	ID      types.VoteID `json:"id"`
	PostID  types.PostID `json:"post_id"`
	VoterID types.UserID `json:"voter_id"`
	Vote    int          `json:"vote"`
}
