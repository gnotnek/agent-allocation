package models

type MarkAsResolvedPayload struct {
	Service struct {
		ID             int     `json:"id"`
		RoomID         string  `json:"room_id"`
		IsResolved     bool    `json:"is_resolved"`
		Notes          *string `json:"notes"` // Nullable field
		FirstCommentID string  `json:"first_comment_id"`
		LastCommentID  int     `json:"last_comment_id"`
		Source         string  `json:"source"`
	} `json:"service"`
	ResolvedBy struct {
		ID          int    `json:"id"`
		Email       string `json:"email"`
		Name        string `json:"name"`
		Type        string `json:"type"`
		IsAvailable bool   `json:"is_available"`
	} `json:"resolved_by"`
	Customer struct {
		UserID string `json:"user_id"`
	} `json:"customer"`
}
