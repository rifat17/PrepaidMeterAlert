//go:build integration

package dpdc

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/m4hi2/MeterAlertBot/internal/config"
	"github.com/m4hi2/MeterAlertBot/internal/datasources"
)

func TestGetBalance_Real(t *testing.T) {
	id := datasources.Identifier{
		AccountNumber: "26188841",
		// MeterNumber: "",
	}

	svc := NewService(config.Get().Dpdc)
	got, err := svc.GetBalance(context.Background(), id)
	if err != nil {
		t.Fatalf("GetBalance error: %v", err)
	}
	t.Logf("Balance: %+v", got)
}

func TestIntrospection(t *testing.T) {
	ctx := context.Background()
	svc := NewService(config.Get().Dpdc)

	token, err := svc.getToken(ctx)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	headers := make(http.Header)
	headers.Set("Authorization", "Bearer "+token)
	headers.Set("accesstoken", token)
	headers.Set("tenantCode", svc.config.TenantCode)

	const introspectionQuery = `{
		__schema {
			types {
				name
				kind
				fields {
					name
					args { name type { name kind ofType { name kind } } }
					type { name kind ofType { name kind ofType { name kind } } }
				}
			}
		}
	}`

	var result map[string]json.RawMessage
	if err := svc.usageClient.Do(ctx, http.MethodPost, "", headers, BalanceQueryRequest{Query: introspectionQuery}, &result); err != nil {
		t.Fatalf("introspection request failed: %v", err)
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("GraphQL schema:\n%s", out)
}
