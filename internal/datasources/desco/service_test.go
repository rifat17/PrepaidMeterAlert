package desco

import (
	"context"
	"testing"

	"github.com/m4hi2/MeterAlertBot/internal/datasources/common"
)

func TestGetBalance_Real(t *testing.T) {
	id := common.Identifier{
		AccountNumber: "41378832",
		// MeterNumber:   "METER001",
	}

	svc := NewService()
	got, err := svc.GetBalance(context.Background(), id)
	if err != nil {
		t.Fatalf("GetBalance error: %v", err)
	}
	t.Logf("Balance: %+v", got)
}
