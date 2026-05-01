package common

import "context"

type DataFetcher interface {
	GetBalance(context.Context, Identifier) (Balance, error)
}
