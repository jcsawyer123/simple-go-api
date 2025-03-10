package aims

// TokenInfo represents the response from AIMS token validation
type TokenInfo struct {
	Roles []Role `json:"roles"`
}

// Role represents an AIMS user role with associated permissions
type Role struct {
	ID          string                 `json:"id"`
	AccountID   string                 `json:"account_id"`
	Name        string                 `json:"name"`
	Permissions map[string]string      `json:"permissions"`
	Version     int                    `json:"version"`
	Created     map[string]interface{} `json:"created"`
	Modified    map[string]interface{} `json:"modified"`
}
