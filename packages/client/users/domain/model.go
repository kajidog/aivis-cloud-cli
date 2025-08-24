package domain

import "time"

// User represents basic user information
type User struct {
	ID          string    `json:"id"`
	Handle      string    `json:"handle"`
	Name        string    `json:"name"`
	Email       string    `json:"email,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	Bio         *string   `json:"bio,omitempty"`
	IsVerified  bool      `json:"is_verified"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserMe represents extended user information for /v1/users/me endpoint
type UserMe struct {
	User
	CreditBalance interface{}   `json:"credit_balance,omitempty"` // Can be number or object
	Settings      *UserSettings `json:"settings,omitempty"`
}

// UserSettings represents user's account settings
type UserSettings struct {
	Language         string `json:"language,omitempty"`
	Timezone         string `json:"timezone,omitempty"`
	EmailNotifications bool  `json:"email_notifications"`
}