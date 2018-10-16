package ggpool

import "context"

//Creator is factory struct interface
type Creator interface {
	Create(ctx context.Context) (interface{}, error)
}
