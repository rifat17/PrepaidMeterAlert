package dpdc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/m4hi2/MeterAlertBot/internal/config"
	"github.com/m4hi2/MeterAlertBot/internal/datasources"
)

const (
	providerName = "dpdc"
)

type Service struct {
	authClient  *datasources.Client
	usageClient *datasources.Client
	config      config.DpdcConfig
	mu          sync.Mutex
	token       string
	tokenExp    time.Time
}

func NewService(cfg config.DpdcConfig) *Service {
	return &Service{
		authClient: datasources.NewClient(&datasources.ClientConfig{
			BasePath: cfg.AuthURL,
			Timeout:  cfg.Timeout,
		}),
		usageClient: datasources.NewClient(&datasources.ClientConfig{
			BasePath: cfg.UsageURL,
			Timeout:  cfg.Timeout,
		}),
		config: cfg,
	}
}

func (s *Service) GetBalance(ctx context.Context, id datasources.Identifier) (datasources.Balance, error) {
	token, err := s.getToken(ctx)
	if err != nil {
		log.Printf("[DPDC] [GetBalance] Auth step failed for account %s: %v", id.AccountNumber, err)
		return datasources.Balance{}, fmt.Errorf("dpdc auth failed: %w", err)
	}

	gql := fmt.Sprintf(`query { postBalanceDetails(input: {customerNumber: "%s", tenantCode: "DPDC"}) { accountId customerName balanceRemaining } }`, id.AccountNumber)
	in := BalanceQueryRequest{Query: gql}

	headers := make(http.Header)
	headers.Set("Authorization", "Bearer "+token)
	headers.Set("accesstoken", token)
	headers.Set("tenantCode", s.config.TenantCode)

	var gqlResp PostBalanceDetailsResponse
	err = s.usageClient.Do(ctx, http.MethodPost, "", headers, &in, &gqlResp)
	if err != nil {
		log.Printf("[DPDC] [GetBalance] HTTP request execution failed for account %s: %v", id.AccountNumber, err)
		return datasources.Balance{}, err
	}

	output := gqlResp.Data.PostBalanceDetails
	log.Printf("[DPDC] [GetBalance] Successfully retrieved balance for account %s (Returned ID: %s)", id.AccountNumber, output.AccountId)

	return datasources.Balance{
		Identifier: datasources.Identifier{AccountNumber: output.AccountId},
		Balance:    output.BalanceRemaining,
	}, nil
}

func (s *Service) Name() string { return providerName }

func (s *Service) getToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != "" && time.Now().Before(s.tokenExp) {
		log.Printf("[DPDC] [getToken] Using unexpired cached token")
		return s.token, nil
	}

	if s.config.ClientID == "" || s.config.ClientSecret == "" {
		log.Printf("[DPDC] [getToken] Error: credentials are blank. ClientID length: %d, ClientSecret length: %d", len(s.config.ClientID), len(s.config.ClientSecret))
		return "", errors.New("DPDC credentials are blank. Check your environment configuration")
	}

	headers := make(http.Header)
	headers.Set("clientId", s.config.ClientID)
	headers.Set("clientSecret", s.config.ClientSecret)
	headers.Set("tenantCode", s.config.TenantCode)

	var auth AuthResponse
	err := s.authClient.Do(ctx, http.MethodPost, "", headers, map[string]any{}, &auth)
	if err != nil {
		log.Printf("[DPDC] [getToken] Auth HTTP execution failed: %v", err)
		return "", err
	}

	if auth.AccessToken == "" {
		log.Printf("[DPDC] [getToken] Error: Auth succeeded but response token payload was blank")
		return "", errors.New("DPDC auth: no access_token in response")
	}

	s.token = auth.AccessToken
	s.tokenExp = time.Now().Add(25 * time.Minute)

	log.Printf("[DPDC] [getToken] Successfully generated new access token. Expires at: %s", s.tokenExp.Format(time.Kitchen))
	return s.token, nil
}
