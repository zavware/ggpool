package ggpool_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alex-zz/ggpool"
)

func TestGet(t *testing.T) {

	factory := &MockFactory{
		destroyedCount: 0,
		createdCount:   0,
	}

	pool, err := ggpool.NewPool(context.Background(), ggpool.Config{
		Capacity:                5,
		MinCapacity:             3,
		ItemLifetime:            20 * time.Second,
		ItemLifetimeCheckPeriod: 3 * time.Second,
		Timeout:                 3 * time.Second,
		Factory:                 factory,
	})

	if err != nil {
		t.Errorf("Unexpected NewPool method error: %s", err)
	}

	time.Sleep(2 * time.Millisecond)

	object, err := pool.Get()
	if _, ok := (*object).(*MockConnection); !ok {
		t.Fatal("Cannot get mock")
	}

	assertEqual(
		t,
		3,
		factory.createdCount,
		fmt.Sprintf("Test case %d: - Unxpected created items count", 1),
	)

	assertEqual(
		t,
		0,
		factory.destroyedCount,
		fmt.Sprintf("Test case %d: - Unxpected destroyed items count", 1),
	)
}

func TestGetDelay(t *testing.T) {

	factory := &MockFactory{
		destroyedCount: 0,
		createdCount:   0,
	}

	pool, err := ggpool.NewPool(context.Background(), ggpool.Config{
		Capacity:                5,
		MinCapacity:             3,
		ItemLifetime:            20 * time.Second,
		ItemLifetimeCheckPeriod: 3 * time.Second,
		Timeout:                 3 * time.Second,
		Factory:                 factory,
	})

	if err != nil {
		t.Errorf("Unexpected NewPool method error: %s", err)
	}

	time.Sleep(2 * time.Millisecond)

	object, err := pool.Get()
	if _, ok := (*object).(*MockConnection); !ok {
		t.Fatal("Cannot get mock")
	}

	assertEqual(
		t,
		3,
		factory.createdCount,
		fmt.Sprintf("Test case %d: - Unxpected created items count", 1),
	)

	assertEqual(
		t,
		0,
		factory.destroyedCount,
		fmt.Sprintf("Test case %d: - Unxpected destroyed items count", 1),
	)
}
