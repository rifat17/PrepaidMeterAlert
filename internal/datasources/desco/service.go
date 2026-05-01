package desco

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/m4hi2/MeterAlertBot/internal/datasources/common"
)

const (
	basePath    = "https://prepaid.desco.org.bd"
	balancePath = "/api/unified/customer/getBalance"
)

const (
	paramAccountNo = "accountNo"
	paramMeterNo   = "meterNo"
)

type Service struct {
	client *common.Client
}

func NewService() *Service {
	return &Service{
		client: common.NewClient(&common.ClientConfig{
			BasePath:   basePath,
			Timeout:    10 * time.Second,
			Retry:      3,
			RetryDelay: time.Second,
		}),
	}
}

func (s *Service) GetBalance(ctx context.Context, id common.Identifier) (common.Balance, error) {
	ctx = context.WithValue(ctx, common.CtxKeyDatasource, common.CtxDatasourceDesco)

	q := url.Values{}
	q.Set(paramAccountNo, id.AccountNumber)
	q.Set(paramMeterNo, id.MeterNumber)
	path := balancePath + "?" + q.Encode()

	var resp GetBalanceResp
	if err := s.client.Do(ctx, http.MethodGet, path, nil, nil, &resp); err != nil {
		return common.Balance{}, fmt.Errorf("get balance: %w", err)
	}

	if resp.Code != http.StatusOK {
		return common.Balance{}, fmt.Errorf("get balance: upstream code %d: %s", resp.Code, resp.Desc)
	}

	return common.Balance{
		Identifier: id,
		Balance:    resp.Data.Balance,
	}, nil
}
