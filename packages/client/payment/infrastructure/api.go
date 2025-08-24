package infrastructure

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/kajidog/aivis-cloud-cli/client/common/http"
	"github.com/kajidog/aivis-cloud-cli/client/payment/domain"
)

type PaymentAPI struct {
	client http.HTTPClient
}

func NewPaymentAPI(client http.HTTPClient) *PaymentAPI {
	return &PaymentAPI{
		client: client,
	}
}

func (a *PaymentAPI) GetSubscriptions(ctx context.Context, req *domain.SubscriptionListRequest) (*domain.SubscriptionListResponse, error) {
	endpoint := "/v1/payment/subscriptions"
	
	params := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			params.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Offset > 0 {
			params.Set("offset", strconv.Itoa(req.Offset))
		}
	}
	
	var response domain.SubscriptionListResponse
	err := a.client.Get(ctx, endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	
	return &response, nil
}

func (a *PaymentAPI) GetCreditTransactions(ctx context.Context, req *domain.CreditTransactionListRequest) (*domain.CreditTransactionListResponse, error) {
	endpoint := "/v1/payment/credit-transactions"
	
	params := url.Values{}
	if req != nil {
		if req.Type != "" {
			params.Set("type", string(req.Type))
		}
		if req.Status != "" {
			params.Set("status", string(req.Status))
		}
		if req.StartDate != nil {
			params.Set("start_date", req.StartDate.Format("2006-01-02"))
		}
		if req.EndDate != nil {
			params.Set("end_date", req.EndDate.Format("2006-01-02"))
		}
		if req.Limit > 0 {
			params.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Offset > 0 {
			params.Set("offset", strconv.Itoa(req.Offset))
		}
	}
	
	var response domain.CreditTransactionListResponse
	err := a.client.Get(ctx, endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit transactions: %w", err)
	}
	
	return &response, nil
}

func (a *PaymentAPI) GetAPIKeys(ctx context.Context, req *domain.APIKeyListRequest) (*domain.APIKeyListResponse, error) {
	endpoint := "/v1/payment/api-keys"
	
	params := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			params.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Offset > 0 {
			params.Set("offset", strconv.Itoa(req.Offset))
		}
	}
	
	var response domain.APIKeyListResponse
	err := a.client.Get(ctx, endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}
	
	return &response, nil
}

func (a *PaymentAPI) CreateAPIKey(ctx context.Context, req *domain.APIKeyCreateRequest) (*domain.APIKey, error) {
	endpoint := "/v1/payment/api-keys"
	
	var response domain.APIKey
	err := a.client.Post(ctx, endpoint, req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}
	
	return &response, nil
}

func (a *PaymentAPI) DeleteAPIKey(ctx context.Context, keyID string) error {
	endpoint := fmt.Sprintf("/v1/payment/api-keys/%s", keyID)
	
	err := a.client.Delete(ctx, endpoint, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete API key %s: %w", keyID, err)
	}
	
	return nil
}

func (a *PaymentAPI) GetUsageSummaries(ctx context.Context, req *domain.UsageSummaryRequest) (*domain.UsageSummary, error) {
	endpoint := "/v1/payment/usage-summaries"
	
	params := url.Values{}
	if req != nil {
		if req.Period != "" {
			params.Set("period", req.Period)
		}
		if req.StartDate != nil {
			params.Set("start_date", req.StartDate.Format("2006-01-02"))
		}
		if req.EndDate != nil {
			params.Set("end_date", req.EndDate.Format("2006-01-02"))
		}
		if req.ModelID != "" {
			params.Set("model_id", req.ModelID)
		}
	}
	
	var response domain.UsageSummary
	err := a.client.Get(ctx, endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summaries: %w", err)
	}
	
	return &response, nil
}