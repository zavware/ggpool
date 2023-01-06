package ggpool_test

import (
	"context"
	"fmt"
	"time"

	"github.com/zav0x/ggpool"
)

// TODO to create real connection?
type Connection struct{}

func (c *Connection) Destroy() {
	return
}

func (c *Connection) RunCommand() {
	fmt.Println("run command")
}

type Factory struct{}

func (f *Factory) Create(ctx context.Context) (interface{}, error) {
	c, err := CreateConnection()
	return c, err
}

func CreateConnection() (*Connection, error) {
	c := &Connection{}
	return c, nil
}

func Example() {

	factory := &Factory{}

	config := ggpool.Config{
		Capacity:                5,
		MinCapacity:             3,
		ItemLifetime:            20 * time.Second,
		ItemLifetimeCheckPeriod: 3 * time.Second,
		Timeout:                 3 * time.Second,
		Factory:                 factory,
	}

	pool, err := ggpool.NewPool(context.Background(), config)

	if err != nil {
		fmt.Println(err)
		return
	}

	obj, err := pool.Get()

	if err != nil {
		fmt.Println(err)
		return
	}

	connection := (*obj).(*Connection)
	connection.RunCommand()

	pool.Release(obj)
	pool.Close()
}
