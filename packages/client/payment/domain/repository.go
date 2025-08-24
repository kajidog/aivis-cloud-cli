package domain

import "context"

type PaymentRepository interface {
	GetSubscriptions(ctx context.Context, req *SubscriptionListRequest) (*SubscriptionListResponse, error)
	GetCreditTransactions(ctx context.Context, req *CreditTransactionListRequest) (*CreditTransactionListResponse, error)
	GetAPIKeys(ctx context.Context, req *APIKeyListRequest) (*APIKeyListResponse, error)
	CreateAPIKey(ctx context.Context, req *APIKeyCreateRequest) (*APIKey, error)
	DeleteAPIKey(ctx context.Context, keyID string) error
	GetUsageSummaries(ctx context.Context, req *UsageSummaryRequest) (*UsageSummary, error)
}