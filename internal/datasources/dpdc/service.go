package dpdc

import (
	"bytes"
	"context"
	"encoding/json"
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
	httpClient *http.Client
	config     config.DpdcConfig
	mu         sync.Mutex
	token      string
	tokenExp   time.Time
}

func NewService(cfg config.DpdcConfig) *Service {
	return &Service{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		config:     cfg,
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
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(in); err != nil {
		log.Printf("[DPDC] [GetBalance] Failed to encode JSON payload for account %s: %v", id.AccountNumber, err)
		return datasources.Balance{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.UsageURL, &buf)
	if err != nil {
		log.Printf("[DPDC] [GetBalance] Failed to create request object for account %s: %v", id.AccountNumber, err)
		return datasources.Balance{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("accesstoken", token)
	req.Header.Set("tenantCode", s.config.TenantCode)
	req.Header.Set("Content-Type", "application/json")

	var gqlResp PostBalanceDetailsResponse
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[DPDC] [GetBalance] HTTP request execution failed for account %s: %v", id.AccountNumber, err)
		return datasources.Balance{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[DPDC] [GetBalance] Non-200 response received from usage service for account %s. Status: %s", id.AccountNumber, resp.Status)
		return datasources.Balance{}, errors.New("Non-200 from DPDC GraphQL usage service")
	}

	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		log.Printf("[DPDC] [GetBalance] Failed to decode GraphQL response body for account %s: %v", id.AccountNumber, err)
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

	emptyBody, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		log.Printf("[DPDC] [getToken] Failed to marshal empty JSON body: %v", err)
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.AuthURL, bytes.NewReader(emptyBody))
	if err != nil {
		log.Printf("[DPDC] [getToken] Failed to create auth request: %v", err)
		return "", err
	}
	req.Header.Set("clientId", s.config.ClientID)
	req.Header.Set("clientSecret", s.config.ClientSecret)
	req.Header.Set("tenantCode", s.config.TenantCode)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[DPDC] [getToken] Auth HTTP execution failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	log.Printf("[DPDC] [getToken] Auth endpoint rejected request. Status: %s. Using ClientID: %s", resp.Status, s.config.ClientID)
	// 	return "", fmt.Errorf("Non-200 from DPDC Auth endpoint: %s", resp.Status)
	// }

	// FIX: Accept both 200 OK and 201 Created
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("[DPDC] [getToken] Auth endpoint rejected request. Status: %s. Using ClientID: %s", resp.Status, s.config.ClientID)
		return "", fmt.Errorf("Non-success status from DPDC Auth endpoint: %s", resp.Status)
	}

	var auth AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		log.Printf("[DPDC] [getToken] Failed to decode auth JSON response body: %v", err)
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
