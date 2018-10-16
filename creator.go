package ggpool

import "context"

//Creator is interface which Factory (see Config) must implement
type Creator interface {
	//Type of object which Create returns must be "pointer"
	Create(ctx context.Context) (interface{}, error)
}
