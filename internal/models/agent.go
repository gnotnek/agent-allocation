package models

type Agent struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	IsAvailable          bool   `json:"is_available"`
	CurrentCustomerCount int    `json:"current_customer_count"`
	AvatarURL            string `json:"avatar_url"`
	CreatedAt            string `json:"created_at"`
	ForceOffline         bool   `json:"force_offline"`
	LastLogin            string `json:"last_login,omitempty"`
	SDKEmail             string `json:"sdk_email"`
	SDKKey               string `json:"sdk_key"`
	Type                 int    `json:"type"`
	TypeAsString         string `json:"type_as_string"`
	UserChannels         []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user_channels"`
	UserRoles []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user_roles"`
}

type ResponseData struct {
	Agents []Agent `json:"agents"`
}

type QiscusResponse struct {
	Data ResponseData `json:"data"`
	Meta struct {
		After      interface{} `json:"after"`
		Before     interface{} `json:"before"`
		PerPage    int         `json:"per_page"`
		TotalCount interface{} `json:"total_count"`
	} `json:"meta"`
	Status int `json:"status"`
}
