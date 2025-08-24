package domain

import "time"

// Subscription represents a user subscription
type Subscription struct {
	ID                  string              `json:"id"`
	PlanID              string              `json:"plan_id"`
	PlanName            string              `json:"plan_name"`
	Status              SubscriptionStatus  `json:"status"`
	CurrentPeriodStart  time.Time           `json:"current_period_start"`
	CurrentPeriodEnd    time.Time           `json:"current_period_end"`
	CancelAtPeriodEnd   bool                `json:"cancel_at_period_end"`
	CanceledAt          *time.Time          `json:"canceled_at,omitempty"`
	Amount              float64             `json:"amount"`
	Currency            string              `json:"currency"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCanceled  SubscriptionStatus = "canceled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusPending   SubscriptionStatus = "pending"
	SubscriptionStatusTrialing  SubscriptionStatus = "trialing"
)

// CreditTransaction represents a credit transaction
type CreditTransaction struct {
	ID            string            `json:"id"`
	Type          TransactionType   `json:"type"`
	Status        TransactionStatus `json:"status"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	Credits       *int64            `json:"credits,omitempty"`
	Description   string            `json:"description"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type TransactionType string

const (
	TransactionTypeCredit TransactionType = "credit"
	TransactionTypeDebit  TransactionType = "debit"
	TransactionTypeRefund TransactionType = "refund"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCanceled  TransactionStatus = "canceled"
)

// APIKey represents an API key
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key,omitempty"` // Only returned on creation
	KeyPreview  string    `json:"key_preview"`   // Masked version
	IsActive    bool      `json:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UsageSummary represents usage statistics
type UsageSummary struct {
	Period          string         `json:"period"`
	TotalCredits    int64          `json:"total_credits"`
	UsedCredits     int64          `json:"used_credits"`
	TTSRequests     int64          `json:"tts_requests"`
	AudioMinutes    float64        `json:"audio_minutes"`
	BreakdownByModel []UsageByModel `json:"breakdown_by_model,omitempty"`
}

type UsageByModel struct {
	ModelID      string  `json:"model_id"`
	ModelName    string  `json:"model_name"`
	Requests     int64   `json:"requests"`
	Credits      int64   `json:"credits"`
	AudioMinutes float64 `json:"audio_minutes"`
}

// Request/Response types
type SubscriptionListRequest struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type SubscriptionListResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
	Total         int            `json:"total"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
	HasMore       bool           `json:"has_more"`
}

type CreditTransactionListRequest struct {
	Type      TransactionType   `json:"type,omitempty"`
	Status    TransactionStatus `json:"status,omitempty"`
	StartDate *time.Time        `json:"start_date,omitempty"`
	EndDate   *time.Time        `json:"end_date,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
}

type CreditTransactionListResponse struct {
	Transactions []CreditTransaction `json:"transactions"`
	Total        int                 `json:"total"`
	Limit        int                 `json:"limit"`
	Offset       int                 `json:"offset"`
	HasMore      bool                `json:"has_more"`
}

type APIKeyListRequest struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type APIKeyListResponse struct {
	APIKeys []APIKey `json:"api_keys"`
	Total   int      `json:"total"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
	HasMore bool     `json:"has_more"`
}

type APIKeyCreateRequest struct {
	Name string `json:"name"`
}

type UsageSummaryRequest struct {
	Period    string     `json:"period,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	ModelID   string     `json:"model_id,omitempty"`
}