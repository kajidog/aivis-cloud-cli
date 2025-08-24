package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kajidog/aivis-cloud-cli/client/payment/domain"
)

type PaymentUsecase struct {
	repo domain.PaymentRepository
}

func NewPaymentUsecase(repo domain.PaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{
		repo: repo,
	}
}

func (p *PaymentUsecase) GetSubscriptions(ctx context.Context, limit, offset int) (*domain.SubscriptionListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	req := &domain.SubscriptionListRequest{
		Limit:  limit,
		Offset: offset,
	}
	
	return p.repo.GetSubscriptions(ctx, req)
}

func (p *PaymentUsecase) GetCreditTransactions(ctx context.Context, transactionType domain.TransactionType, status domain.TransactionStatus, startDate, endDate *time.Time, limit, offset int) (*domain.CreditTransactionListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return nil, fmt.Errorf("start date cannot be after end date")
	}
	
	req := &domain.CreditTransactionListRequest{
		Type:      transactionType,
		Status:    status,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     limit,
		Offset:    offset,
	}
	
	return p.repo.GetCreditTransactions(ctx, req)
}

func (p *PaymentUsecase) GetAPIKeys(ctx context.Context, limit, offset int) (*domain.APIKeyListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	req := &domain.APIKeyListRequest{
		Limit:  limit,
		Offset: offset,
	}
	
	return p.repo.GetAPIKeys(ctx, req)
}

func (p *PaymentUsecase) CreateAPIKey(ctx context.Context, name string) (*domain.APIKey, error) {
	if name == "" {
		return nil, fmt.Errorf("API key name is required")
	}
	
	req := &domain.APIKeyCreateRequest{
		Name: name,
	}
	
	return p.repo.CreateAPIKey(ctx, req)
}

func (p *PaymentUsecase) DeleteAPIKey(ctx context.Context, keyID string) error {
	if keyID == "" {
		return fmt.Errorf("API key ID is required")
	}
	
	return p.repo.DeleteAPIKey(ctx, keyID)
}

func (p *PaymentUsecase) GetUsageSummaries(ctx context.Context, period string, startDate, endDate *time.Time, modelID string) (*domain.UsageSummary, error) {
	if period == "" && startDate == nil && endDate == nil {
		period = "month"
	}
	
	if period != "" {
		validPeriods := map[string]bool{
			"day":   true,
			"week":  true,
			"month": true,
			"year":  true,
		}
		if !validPeriods[period] {
			return nil, fmt.Errorf("invalid period: %s", period)
		}
	}
	
	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return nil, fmt.Errorf("start date cannot be after end date")
	}
	
	req := &domain.UsageSummaryRequest{
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
		ModelID:   modelID,
	}
	
	return p.repo.GetUsageSummaries(ctx, req)
}