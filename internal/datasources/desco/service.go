package desco

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/m4hi2/MeterAlertBot/internal/config"
	"github.com/m4hi2/MeterAlertBot/internal/datasources"
	"golang.org/x/time/rate"
)

const name = "desco"

const balancePath = "/api/unified/customer/getBalance"

const (
	paramAccountNo = "accountNo"
	paramMeterNo   = "meterNo"
)

type Service struct {
	client  *datasources.Client
	limiter *rate.Limiter
}

func NewService(cfg config.DescoConfig) *Service {
	return &Service{
		client: datasources.NewClient(&datasources.ClientConfig{
			BasePath:   cfg.BasePath,
			Timeout:    cfg.Timeout,
			Retry:      cfg.Retry,
			RetryDelay: cfg.RetryDelay,
		}),
		limiter: rate.NewLimiter(rate.Limit(cfg.RateLimit), 1),
	}
}

func (s *Service) GetBalance(ctx context.Context, id datasources.Identifier) (datasources.Balance, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return datasources.Balance{}, fmt.Errorf("rate limit wait: %w", err)
	}

	ctx = context.WithValue(ctx, datasources.CtxKeyDatasource, datasources.CtxDatasourceDesco)

	q := url.Values{}
	q.Set(paramAccountNo, id.AccountNumber)
	q.Set(paramMeterNo, id.MeterNumber)
	path := balancePath + "?" + q.Encode()

	var resp GetBalanceResp
	if err := s.client.Do(ctx, http.MethodGet, path, nil, nil, &resp); err != nil {
		return datasources.Balance{}, fmt.Errorf("get balance: %w", err)
	}

	if resp.Code != http.StatusOK {
		return datasources.Balance{}, fmt.Errorf("get balance: upstream code %d: %s", resp.Code, resp.Desc)
	}

	return datasources.Balance{
		Identifier: id,
		Balance:    resp.Data.Balance,
	}, nil
}

func (s *Service) Name() string {
	return name
}
