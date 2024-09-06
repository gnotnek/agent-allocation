package models

type MarkAsResolvedPayload struct {
	Customer struct {
		Additionaliinfo []string `json:"additional_info"`
		AvatarURL       string   `json:"avatar_url"`
		Name            string   `json:"name"`
		UserID          string   `json:"user_id"`
	}
	ResolvedBy struct {
		Email        string `json:"email"`
		Id           string `json:"id"`
		Is_available bool   `json:"is_available"`
		Name         string `json:"name"`
		Type         string `json:"type"`
	}
	Service struct {
		FirstCommentID string `json:"first_comment_id"`
		ID             int    `json:"id"`
		IsResolved     bool   `json:"is_resolved"`
		LastCommentID  string `json:"last_comment_id"`
		Notes          string `json:"notes"`
		RoomID         string `json:"room_id"`
		Source         string `json:"source"`
	}
}
